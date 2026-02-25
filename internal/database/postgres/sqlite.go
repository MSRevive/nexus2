package postgres

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/msrevive/nexus2/internal/database"
	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

// writeOp is a unit of work sent through the serialized write channel.
// Every mutating DB call goes through here so SQLite's single-writer
// constraint is respected without any external locking.
type writeOp struct {
	fn   func(tx *sql.Tx) error
	resp chan error
}

// pendingUpdate holds the latest state for a character that has been
// updated but not yet flushed to the database.
type pendingUpdate struct {
	size       int
	data       string
	backupMax  int
	backupTime time.Duration
}

type sqliteDB struct {
	db *sql.DB

	// writeCh is the single-writer channel. Only one goroutine reads from it,
	// so all DB writes are naturally serialized — no locking needed for writes.
	writeCh chan writeOp

	// flushInterval controls how often the coalescing buffer is drained.
	flushInterval time.Duration

	// pendingUpdates is the coalescing map. When UpdateCharacter is called,
	// we just overwrite the entry for that character ID. On each flush tick,
	// all pending entries are committed in a single transaction.
	coalesceMu     sync.Mutex
	pendingUpdates map[uuid.UUID]pendingUpdate

	done chan struct{}
	wg   sync.WaitGroup

	database.Options
}

func New() *sqliteDB {
	return &sqliteDB{
		writeCh:        make(chan writeOp, 512),
		flushInterval:  500 * time.Millisecond,
		pendingUpdates: make(map[uuid.UUID]pendingUpdate),
		done:           make(chan struct{}),
	}
}

func (d *sqliteDB) Connect(cfg database.Config, opts database.Options) error {
	// WAL mode + NORMAL sync gives the best write throughput while still
	// being crash-safe. busy_timeout prevents "database is locked" errors
	// during the brief windows where SQLite is checkpointing.
	dsn := fmt.Sprintf("%s?_journal=WAL&_synchronous=NORMAL&_busy_timeout=5000", cfg.SQLite.Path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return err
	}

	// Crucial: limit to a single open connection so SQLite's file-level
	// write lock is never contended from within our own process.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if err := db.Ping(); err != nil {
		return fmt.Errorf("sqlite ping: %w", err)
	}

	if err := migrate(db); err != nil {
		return fmt.Errorf("sqlite migrate: %w", err)
	}

	d.db = db

	d.wg.Add(2)
	go d.writeWorker()
	go d.flushWorker()

	return nil
}

func (d *sqliteDB) Disconnect() error {
	close(d.done)
	d.wg.Wait()
	return d.db.Close()
}

// SyncToDisk issues a passive WAL checkpoint so data in the WAL file
// is folded back into the main database file.
func (d *sqliteDB) SyncToDisk() error {
	_, err := d.db.Exec("PRAGMA wal_checkpoint(PASSIVE)")
	return err
}

// RunGC flushes pending updates and then purges any soft-deleted characters
// whose expiration timestamp has passed.
func (d *sqliteDB) RunGC() error {
	if err := d.flushPendingUpdates(); err != nil {
		return err
	}

	return d.exec(func(tx *sql.Tx) error {
		_, err := tx.Exec(
			`DELETE FROM characters WHERE expires_at IS NOT NULL AND expires_at <= datetime('now')`,
		)
		return err
	})
}

// exec is the public helper for ad-hoc write operations. It packages the
// function into a writeOp, ships it to the single writer goroutine, and
// blocks until the result comes back.
func (d *sqliteDB) exec(fn func(tx *sql.Tx) error) error {
	resp := make(chan error, 1)
	d.writeCh <- writeOp{fn: fn, resp: resp}
	return <-resp
}

// writeWorker is the ONLY goroutine that opens transactions and writes to
// the database. This gives SQLite a single writer at all times.
func (d *sqliteDB) writeWorker() {
	defer d.wg.Done()

	runOp := func(op writeOp) {
		tx, err := d.db.Begin()
		if err != nil {
			op.resp <- err
			return
		}
		if err := op.fn(tx); err != nil {
			_ = tx.Rollback()
			op.resp <- err
			return
		}
		op.resp <- tx.Commit()
	}

	for {
		select {
		case op := <-d.writeCh:
			runOp(op)

		case <-d.done:
			// Drain any remaining ops that arrived before shutdown.
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
// On shutdown it performs one final flush so no updates are lost.
func (d *sqliteDB) flushWorker() error {
	defer d.wg.Done()
	ticker := time.NewTicker(d.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := d.flushPendingUpdates(); err != nil {
				return fmt.Errorf("sqlite: flush error: %v", err)
			}

		case <-d.done:
			if err := d.flushPendingUpdates(); err != nil {
				return fmt.Errorf("sqlite: final flush error: %v", err)
			}
			return nil
		}
	}
}

// flushPendingUpdates atomically swaps the coalescing map for a fresh one,
// then commits all coalesced updates in a single transaction. N calls to
// UpdateCharacter for the same character between ticks become exactly 1
// database write.
func (d *sqliteDB) flushPendingUpdates() error {
	d.coalesceMu.Lock()
	if len(d.pendingUpdates) == 0 {
		d.coalesceMu.Unlock()
		return nil
	}
	// Swap out the map so callers can keep writing while we flush.
	snapshot := d.pendingUpdates
	d.pendingUpdates = make(map[uuid.UUID]pendingUpdate)
	d.coalesceMu.Unlock()

	return d.exec(func(tx *sql.Tx) error {
		for id, upd := range snapshot {
			if err := applyCharacterUpdate(tx, id, upd); err != nil {
				return fmt.Errorf("flush update for %s: %w", id, err)
			}
		}
		return nil
	})
}

// migrate creates the schema on first run. Queries are idempotent (IF NOT EXISTS).
// When moving to Postgres: swap TEXT for UUID, DATETIME for TIMESTAMPTZ,
// AUTOINCREMENT for GENERATED ALWAYS AS IDENTITY, and ? for $N placeholders.
func migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id         TEXT PRIMARY KEY,
			revision   INTEGER NOT NULL DEFAULT 0,
			flags      INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS characters (
			id              TEXT PRIMARY KEY,
			steam_id        TEXT REFERENCES users(id),
			slot            INTEGER,
			created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			deleted_at      DATETIME,
			expires_at      DATETIME,      -- populated on soft-delete for GC
			data_created_at DATETIME,
			data_size       INTEGER NOT NULL DEFAULT 0,
			data_payload    TEXT NOT NULL DEFAULT ''
		);

		CREATE TABLE IF NOT EXISTS deleted_characters (
			steam_id     TEXT NOT NULL REFERENCES users(id),
			slot         INTEGER NOT NULL,
			character_id TEXT NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
			deleted_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (steam_id, slot)
			UNIQUE (character_id)
		);

		-- Stores the version history (Versions []CharacterData on the schema struct).
		-- Ordered by autoincrement id to preserve insertion order.
		CREATE TABLE IF NOT EXISTS character_versions (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			character_id TEXT NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
			created_at   DATETIME NOT NULL,
			size         INTEGER NOT NULL,
			data_payload TEXT NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_chars_steam_id   ON characters(steam_id);
		CREATE INDEX IF NOT EXISTS idx_charver_char_id  ON character_versions(character_id);
	`)
	return err
}
