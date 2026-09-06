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
	"github.com/msrevive/nexus2/internal/database/coalesce"
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

type sqliteDB struct {
	db *sql.DB

	// writeCh is the single-writer channel. Only one goroutine reads from it,
	// so all DB writes are naturally serialized — no locking needed for writes.
	writeCh chan writeOp

	// buf coalesces character updates so repeated saves for the same character
	// collapse into one write. See internal/database/coalesce.
	buf *coalesce.Buffer

	done chan struct{}
	wg   sync.WaitGroup

	database.Options
}

func New() *sqliteDB {
	return &sqliteDB{
		writeCh: make(chan writeOp, 512),
		done:    make(chan struct{}),
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

	d.buf = coalesce.New(coalesce.Config{
		Interval:  cfg.FlushInterval,
		Threshold: cfg.FlushThreshold,
		Name:      "sqlite",
		Logger:    opts.Logger,
	}, d.applyUpdates)

	d.wg.Add(1)
	go d.writeWorker()
	d.buf.Start()

	return nil
}

func (d *sqliteDB) Disconnect() error {
	if d.db == nil {
		return nil
	}

	// Stop the buffer first. Its final flush is a write, and every write goes
	// through writeWorker — so the writer has to still be alive to serve it.
	// Tearing them down in the other order deadlocks: the flush parks on a
	// response that nobody is left to send.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := d.buf.Stop(ctx); err != nil && d.Logger != nil {
		d.Logger.Error("sqlite: final flush error", "error", err)
	}

	close(d.done)
	d.wg.Wait()
	return d.db.Close()
}

// applyUpdates commits a batch of coalesced character updates in one
// transaction. It is the ApplyFunc handed to the buffer.
func (d *sqliteDB) applyUpdates(ctx context.Context, batch []coalesce.Entry) error {
	ctx, span := oida.Start(ctx, "UPDATE characters (flush)", oida.KindDatabase)
	defer span.End()
	span.SetAttribute("characters", len(batch))

	err := d.exec(ctx, func(ctx context.Context, tx *sql.Tx) error {
		for _, e := range batch {
			if err := applyCharacterUpdate(ctx, tx, e.ID, e.Update); err != nil {
				return fmt.Errorf("flush update for %s: %w", e.ID, err)
			}
		}
		return nil
	})
	span.RecordError(err)
	return err
}

// SyncToDisk drains the coalescing buffer, then issues a passive WAL checkpoint
// so data in the WAL file is folded back into the main database file. Without
// the flush the buffered updates aren't in the WAL yet, so there'd be nothing
// for the checkpoint to fold back.
func (d *sqliteDB) SyncToDisk(ctx context.Context) error {
	// The sync and GC crons are scheduled before Connect runs, so a tick can
	// land while the connection is still being established.
	if d.db == nil {
		return database.ErrNotAvailable
	}

	return d.observe(ctx, "sqlite SyncToDisk", func(ctx context.Context) error {
		if err := d.buf.Flush(ctx); err != nil {
			return err
		}

		ctx, span := oida.Start(ctx, "PRAGMA wal_checkpoint", oida.KindDatabase)
		defer span.End()

		_, err := d.db.ExecContext(ctx, "PRAGMA wal_checkpoint(PASSIVE)")
		span.RecordError(err)
		return err
	})
}

// RunGC purges any soft-deleted characters whose expiration timestamp has passed.
//
// It flushes first so updates queued before the purge are written rather than
// discarded, but that flush is advisory: the buffer tolerates a character
// vanishing underneath it, so a flush failure must not stop garbage collection
// from running.
//
// There is deliberately no attempt to evict the purged IDs from the buffer.
// Only soft-deleted characters can expire, SoftDeleteCharacter already drops
// whatever was queued for them, and nothing can queue an update for a character
// it can no longer look up — so there would be nothing to evict.
func (d *sqliteDB) RunGC(ctx context.Context) error {
	if d.db == nil {
		return database.ErrNotAvailable
	}

	return d.observe(ctx, "sqlite RunGC", func(ctx context.Context) error {
		if err := d.buf.Flush(ctx); err != nil && d.Logger != nil {
			d.Logger.Warn("sqlite: flush before GC failed, collecting anyway", "error", err)
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
