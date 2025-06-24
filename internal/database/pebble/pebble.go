package pebble

import (
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

	d.db = db
	return nil
}

func (d *pebbleDB) Disconnect() error {
	return d.db.Close()
}

func (d *pebbleDB) SyncToDisk() error {
	//return d.db.Flush()
	
	// this should be the most optimal way to sync the data in memory https://github.com/cockroachdb/pebble/issues/4598
	return d.db.LogData(nil, pebble.Sync)
}

func (d *pebbleDB) RunGC() error {
	return nil
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