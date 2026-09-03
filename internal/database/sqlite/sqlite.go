package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"
	"path/filepath"
	"os"

	"github.com/msrevive/nexus2/internal/database"
	"github.com/google/uuid"
	"github.com/titpetric/oida"
	_ "modernc.org/sqlite"
)

// writeOp is a unit of work sent through the serialized write channel.
// Every mutating DB call goes through here so SQLite's single-writer
// constraint is respected without any external locking.
type writeOp struct {
	ctx  context.Context
	fn   func(ctx context.Context, tx *sql.Tx) error
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
		flushInterval:  5 * time.Second,
		pendingUpdates: make(map[uuid.UUID]pendingUpdate),
		done:           make(chan struct{}),
	}
}

func (d *sqliteDB) Connect(cfg database.Config, opts database.Options) error {
	// Ensure all parent directories exist before opening the SQLite file.
	if dir := filepath.Dir(cfg.SQLite.Path); dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("sqlite mkdir: %w", err)
		}
	}

	// modernc.org/sqlite takes pragmas as _pragma=name(value); the _journal /
	// _synchronous / _busy_timeout spelling is mattn/go-sqlite3's and is silently
	// ignored by this driver, which left the database in the default rollback
	// journal with no busy timeout.
	dsn := fmt.Sprintf("%s?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=busy_timeout(5000)", cfg.SQLite.Path)
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
	d.Options = opts

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

// SyncToDisk drains the coalescing buffer, then issues a passive WAL checkpoint
// so data in the WAL file is folded back into the main database file. Without
// the flush the buffered updates aren't in the WAL yet, so there'd be nothing
// for the checkpoint to fold back.
func (d *sqliteDB) SyncToDisk(ctx context.Context) error {
	return d.observe(ctx, "sqlite SyncToDisk", func(ctx context.Context) error {
		if err := d.flushPendingUpdates(ctx); err != nil {
			return err
		}

		ctx, span := oida.Start(ctx, "PRAGMA wal_checkpoint", oida.KindDatabase)
		defer span.End()

		_, err := d.db.ExecContext(ctx, "PRAGMA wal_checkpoint(PASSIVE)")
		span.RecordError(err)
		return err
	})
}

// RunGC flushes pending updates and then purges any soft-deleted characters
// whose expiration timestamp has passed.
func (d *sqliteDB) RunGC(ctx context.Context) error {
	return d.observe(ctx, "sqlite RunGC", func(ctx context.Context) error {
		// Flush first: purging a row that still has a queued update would make the
		// next flush fail on it, and a failed flush drops the whole snapshot.
		if err := d.flushPendingUpdates(ctx); err != nil {
			return err
		}

		ctx, span := oida.Start(ctx, "DELETE expired characters", oida.KindDatabase)
		defer span.End()

		err := d.exec(ctx, func(ctx context.Context, tx *sql.Tx) error {
			res, err := tx.ExecContext(ctx,
				`DELETE FROM characters WHERE expires_at IS NOT NULL AND expires_at <= datetime('now')`,
			)
			if err != nil {
				return err
			}
			if n, err := res.RowsAffected(); err == nil {
				span.SetAttribute("deleted", n)
			}
			return nil
		})
		span.RecordError(err)
		return err
	})
}

// observe runs fn inside its own trace so background work that has no request
// behind it still shows up on the dashboard. Without a tracer it just runs fn.
func (d *sqliteDB) observe(ctx context.Context, name string, fn func(context.Context) error) error {
	if d.Tracer == nil {
		return fn(ctx)
	}

	return d.Tracer.Observe(ctx, name, fn)
}

// exec is the public helper for ad-hoc write operations. It packages the
// function into a writeOp, ships it to the single writer goroutine, and
// blocks until the result comes back.
func (d *sqliteDB) exec(ctx context.Context, fn func(ctx context.Context, tx *sql.Tx) error) error {
	resp := make(chan error, 1)

	select {
	case d.writeCh <- writeOp{ctx: ctx, fn: fn, resp: resp}:
	case <-ctx.Done():
		return ctx.Err()
	}

	select {
	case err := <-resp:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// writeWorker is the ONLY goroutine that opens transactions and writes to
// the database. This gives SQLite a single writer at all times.
func (d *sqliteDB) writeWorker() {
	defer d.wg.Done()

	runOp := func(op writeOp) {
		ctx := op.ctx
		if ctx == nil {
			ctx = context.Background()
		}

		tx, err := d.db.BeginTx(ctx, nil)
		if err != nil {
			op.resp <- err
			return
		}
		if err := op.fn(ctx, tx); err != nil {
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
func (d *sqliteDB) flushWorker() {
	defer d.wg.Done()
	ticker := time.NewTicker(d.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := d.flushPendingUpdates(context.Background()); err != nil && d.Logger != nil {
				d.Logger.Error("sqlite: flush error", "error", err)
			}

		case <-d.done:
			if err := d.flushPendingUpdates(context.Background()); err != nil && d.Logger != nil {
				d.Logger.Error("sqlite: final flush error", "error", err)
			}
			return
		}
	}
}

// flushPendingUpdates atomically swaps the coalescing map for a fresh one,
// then commits all coalesced updates in a single transaction. N calls to
// UpdateCharacter for the same character between ticks become exactly 1
// database write.
func (d *sqliteDB) flushPendingUpdates(ctx context.Context) error {
	d.coalesceMu.Lock()
	if len(d.pendingUpdates) == 0 {
		d.coalesceMu.Unlock()
		return nil
	}
	// Swap out the map so callers can keep writing while we flush.
	snapshot := d.pendingUpdates
	d.pendingUpdates = make(map[uuid.UUID]pendingUpdate)
	d.coalesceMu.Unlock()

	ctx, span := oida.Start(ctx, "UPDATE characters (flush)", oida.KindDatabase)
	defer span.End()
	span.SetAttribute("characters", len(snapshot))

	err := d.exec(ctx, func(ctx context.Context, tx *sql.Tx) error {
		for id, upd := range snapshot {
			if err := applyCharacterUpdate(ctx, tx, id, upd); err != nil {
				return fmt.Errorf("flush update for %s: %w", id, err)
			}
		}
		return nil
	})
	span.RecordError(err)
	return err
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
