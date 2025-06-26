package pebble

import (
	"time"
	"io"
	"fmt"
	"encoding/binary"

	"github.com/msrevive/nexus2/internal/database"

	"github.com/cockroachdb/pebble/v2"
)

// The smaller the key prefix the better?
// Each time this changes the database needs to be recreated.
var (
	UserPrefix = []byte("users:")
	CharPrefix = []byte("chars:")
)

type pebbleDB struct {
	db *pebble.DB
	batch *pebble.Batch
}

func New() *pebbleDB {
	return &pebbleDB{}
}

func (d *pebbleDB) Connect(cfg database.Config, opts database.Options) error {
	db, err := pebble.Open(cfg.Badger.Directory, &pebble.Options{
		FormatMajorVersion: pebble.FormatDefault,
	})
	if err != nil {
		return err
	}

	d.batch = db.NewBatch()

	d.db = db
	return nil
}

func (d *pebbleDB) Disconnect() (err error) {
	err = d.batch.Close()
	err = d.db.Close()
	return
}

func (d *pebbleDB) SyncToDisk() error {
	//return d.db.Flush()

	// this should be the most optimal way to sync the data in memory https://github.com/cockroachdb/pebble/issues/4598
	return d.db.LogData(nil, pebble.Sync)
}

func (d *pebbleDB) RunGC() error {
	it, err := d.db.NewIter(nil)
	if err != nil {
		return err
	}
	defer it.Close()

	if d.batch.Count() > 0 {
		if err := d.batch.Commit(pebble.Sync); err != nil {
			return err
		}
	}
	d.batch.Reset()

	for it.First(); it.Valid(); it.Next() {
		value := it.Value()
		exTime := int64(binary.BigEndian.Uint64(value[:timestampSize])) // unix is int64 so this needs to be int64
		fmt.Printf("key %s, expire: %d\n", it.Key(), exTime)

		if (len(value) >= timestampSize) && (exTime > 0) {
			fmt.Println("has TTL")
			if time.Now().Unix() > exTime {
				fmt.Printf("key whipped %s\n", it.Key())
				// if err := d.batch.Delete(it.Key(), nil); err != nil {
				// 	return err
				// }
			}
		}
	}

	return nil
}

/*
	Helper functions
	- functions to implement our own TTL implementation and prefix handling for iterators.
*/
//8 bytes for a Unix timestamp
const timestampSize = 8

func (d *pebbleDB) setWithTTL(key, value []byte, ttl time.Duration, opts *pebble.WriteOptions) error {
	expiration := time.Now().Add(ttl).Unix()
	fmt.Println(uint64(expiration))
	buf := make([]byte, timestampSize+len(value)) //8 bytes for a Unix timestamp
	binary.BigEndian.PutUint64(buf, uint64(expiration))
	copy(buf[timestampSize:], value)

	return d.db.Set(key, buf, opts)
}

func (d *pebbleDB) get(key []byte) ([]byte, io.Closer, error) {
	value, closer, err := d.db.Get(key)
	exTime := binary.BigEndian.Uint64(value[:timestampSize])
	if err != nil {
		return nil, nil, err
	}

	// if there's no timestamp then just return.
	if (len(value) < timestampSize) || (exTime == 0) {
		return value, closer, nil
	}

	// we don't check if the entry is expired because we don't need to for this.
	// we need to get everything after the first timestamp size.
	return value[timestampSize:], closer, nil
}

// keyUpperBound returns the smallest key that is lexicographically greater than the given prefix.
// This is used to define the exclusive upper bound for a prefix iteration.
func keyUpperBound(b []byte) []byte {
	end := make([]byte, len(b))
	copy(end, b)
	for i := len(end) - 1; i >= 0; i-- {
		end[i]++
		if end[i] != 0 {
			return end[:i+1]
		}
	}
	return nil // The prefix is all 0xFF bytes, which is the end of the key space.
}