package pebble

import (
	"time"
	"io"
	"fmt"
	"context"
	"encoding/binary"
	"sync"

	"github.com/msrevive/nexus2/internal/database"

	"github.com/cockroachdb/pebble/v2"
	"github.com/cockroachdb/pebble/v2/vfs"
)

// The smaller the key prefix the better?
// Each time this changes the database needs to be recreated.
var (
	UserPrefix = []byte("users:")
	CharPrefix = []byte("chars:")
)

// pendingUpdate holds the latest coalesced state for a character.
// Only the most recent call to UpdateCharacter per character ID between
// flush ticks is kept — intermediate states are discarded.
type pendingUpdate struct {
	size       int
	data       string
	backupMax  int
	backupTime time.Duration
}

// writeOp is a unit of work for the single-writer goroutine.
// fn receives an empty Batch; when fn returns nil the batch is committed.
type writeOp struct {
	fn   func(b *pebble.Batch) error
	resp chan error
}

type pebbleDB struct {
	db *pebble.DB
	
	// writeCh serializes all mutations through a single goroutine so
	// Pebble's batch commits are never raced.
	writeCh chan writeOp

	// flushInterval controls how often the coalescing buffer is drained.
	flushInterval time.Duration

	// coalesceMu protects pendingUpdates. Only UpdateCharacter writes to it;
	// the flush goroutine swaps it out under the lock.
	coalesceMu sync.Mutex
	pendingUpdates map[string]pendingUpdate // key = char UUID string

	done chan struct{}
	wg sync.WaitGroup
	
	database.Options
}

func New() *pebbleDB {
	return &pebbleDB{
		writeCh: make(chan writeOp, 512),
		flushInterval: 500 * time.Millisecond,
		pendingUpdates: make(map[string]pendingUpdate),
		done: make(chan struct{}),
	}
}

func (d *pebbleDB) Connect(cfg database.Config, opts database.Options) error {
	var db *pebble.DB
	var err error

	if cfg.Pebble.Directory == "" {
		db, err = pebble.Open("", &pebble.Options{
			FormatMajorVersion: pebble.FormatColumnarBlocks,
			FS: vfs.NewMem(),
		})
	} else {
		db, err = pebble.Open(cfg.Pebble.Directory, &pebble.Options{
			FormatMajorVersion: pebble.FormatColumnarBlocks,
		})
	}

	if err != nil {
		return err
	}

	d.wg.Add(2)
	go d.writeWorker()
	go d.flushWorker()

	d.db = db
	return nil
}

func (d *pebbleDB) Disconnect() (err error) {
	close(d.done)
	d.wg.Wait()
	return d.db.Close()
}

func (d *pebbleDB) SyncToDisk() error {
	//return d.db.Flush()

	// this should be the most optimal way to sync the data in memory https://github.com/cockroachdb/pebble/issues/4598
	return d.db.LogData(nil, pebble.Sync)
}

func (d *pebbleDB) RunGC() error {
	// Flush pending updates before iterating so nothing stale is left behind.
	if err := d.flushPendingUpdates(); err != nil {
		return err
	}

	return d.exec(func(b *pebble.Batch) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		it, err := d.db.NewIterWithContext(ctx, nil)
		if err != nil {
			return err
		}
		defer it.Close()

		for it.First(); it.Valid(); it.Next() {
			value := it.Value()
			if len(value) > 0 && value[0] == ttlMagic && len(value) >= 1+timestampSize {
				exTime := int64(binary.BigEndian.Uint64(value[1 : 1+timestampSize]))
				if time.Now().Unix() > exTime {
					if err := b.Delete(it.Key(), nil); err != nil {
						return err
					}
				}
			}
		}
		return nil
	})
}

/*
	Helper functions
	- functions to implement our own TTL implementation and prefix handling for iterators.
*/
//8 bytes for a Unix timestamp, this is cause in 2032 8 bytes will be needed to store unix timestamp
const (
    timestampSize = 8
    ttlMagic = byte(0xFF) // marker to indicate this value has a TTL prefix
)

func (d *pebbleDB) setWithTTL(key, value []byte, ttl time.Duration, opts *pebble.WriteOptions) error {
    expiration := time.Now().Add(ttl).Unix()
    // layout: [0xFF magic][8 bytes expiry][value]
    buf := make([]byte, 1+timestampSize+len(value))
    buf[0] = ttlMagic
    binary.BigEndian.PutUint64(buf[1:], uint64(expiration))
    copy(buf[1+timestampSize:], value)

    return d.db.Set(key, buf, opts)
}

func (d *pebbleDB) get(key []byte) ([]byte, io.Closer, error) {
    value, closer, err := d.db.Get(key)
    if err != nil {
        return nil, nil, err
    }
    defer closer.Close()

    // no TTL prefix, return as-is
    if len(value) == 0 || value[0] != ttlMagic {
        out := make([]byte, len(value))
        copy(out, value)
        return out, closer, nil
    }

    // has TTL prefix
    if len(value) < 1+timestampSize {
        return nil, nil, fmt.Errorf("corrupt value for key")
    }

    exTime := int64(binary.BigEndian.Uint64(value[1 : 1+timestampSize]))
    if time.Now().Unix() > exTime {
        closer.Close()
        // expired - delete it and return not found
        _ = d.db.Delete(key, pebble.NoSync)
        return nil, nil, pebble.ErrNotFound
    }

    payload := make([]byte, len(value)-1-timestampSize)
    copy(payload, value[1+timestampSize:])
    return payload, closer, nil
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

// exec sends fn to the single-writer goroutine and blocks until it completes.
// fn receives a fresh Batch; the writer commits it after fn returns nil.
func (d *pebbleDB) exec(fn func(b *pebble.Batch) error) error {
	resp := make(chan error, 1)
	d.writeCh <- writeOp{fn: fn, resp: resp}
	return <-resp
}

// writeWorker is the only goroutine that commits batches to Pebble.
// Centralizing writes here means batches are never committed concurrently,
// which avoids contention on Pebble's commit pipeline.
func (d *pebbleDB) writeWorker() {
	defer d.wg.Done()

	runOp := func(op writeOp) {
		b := d.db.NewBatch()
		if err := op.fn(b); err != nil {
			_ = b.Close()
			op.resp <- err
			return
		}

		op.resp <- b.Commit(pebble.NoSync)
	}

	for {
		select {
		case op := <-d.writeCh:
			runOp(op)
		case <-d.done:
			for {
				select {
				case op := <-d.writeCh:
					runOp(op)
				default:
					return
				}
			}
		}
	}
}

// flushWorker ticks on flushInterval and drains the coalescing buffer.
// On shutdown it performs one last flush so no pending updates are dropped.
func (d *pebbleDB) flushWorker() error {
	defer d.wg.Done()
	ticker := time.NewTicker(d.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := d.flushPendingUpdates(); err != nil {
				return fmt.Errorf("pebble: flush error: %v", err)
			}
		case <-d.done:
			if err := d.flushPendingUpdates(); err != nil {
				return fmt.Errorf("pebble: final flush error: %v", err)
			}
			return nil
		}
	}
}

// flushPendingUpdates atomically swaps the coalescing map for a fresh one,
// then applies every pending character update inside a single Batch commit.
//
// Each call to UpdateCharacter between ticks simply overwrites the map entry
// for that character ID. N calls for the same character → exactly 1 Set in
// Pebble. All of those Sets land in one batch → one WAL append regardless of
// how many characters were updated.
func (d *pebbleDB) flushPendingUpdates() error {
	d.coalesceMu.Lock()
	if len(d.pendingUpdates) == 0 {
		d.coalesceMu.Unlock()
		return nil
	}
	snapshot := d.pendingUpdates
	d.pendingUpdates = make(map[string]pendingUpdate)
	d.coalesceMu.Unlock()

	return d.exec(func(b *pebble.Batch) error {
		for charIDStr, upd := range snapshot {
			if err := d.applyCharacterUpdate(b, charIDStr, upd); err != nil {
				return fmt.Errorf("flush update for %s: %w", charIDStr, err)
			}
		}
		return nil
	})
}