package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/msrevive/nexus2/internal/database"
	"github.com/msrevive/nexus2/internal/database/coalesce"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/titpetric/oida"
)

type postgresDB struct {
	db *pgxpool.Pool

	// buf coalesces character updates so repeated saves for the same character
	// collapse into one write. See internal/database/coalesce.
	buf *coalesce.Buffer

	database.Options
}

func New() *postgresDB {
	return &postgresDB{}
}

func (d *postgresDB) Connect(cfg database.Config, opts database.Options) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel();

	poolCfg, err := pgxpool.ParseConfig(cfg.Postgres.Conn)
	if err != nil {
		return fmt.Errorf("postgres: parse dsn: %w", err)
	}

	poolCfg.MinConns = cfg.Postgres.MinConns
	poolCfg.MaxConns = cfg.Postgres.MaxConns

	// Health-check idle connections periodically so stale connections to a
	// remote instance (which may be behind a load-balancer or firewall with
	// idle timeouts) are replaced before they cause query failures.
	poolCfg.HealthCheckPeriod = 30 * time.Second

	// Keep idle connections alive for a reasonable window. Managed instances
	// (e.g. RDS, Cloud SQL) often terminate connections idle > 10 min.
	poolCfg.MaxConnIdleTime = 5 * time.Minute
	poolCfg.MaxConnLifetime = 30 * time.Minute

	// Per-connection timeouts protect against network partitions to the
	// remote host.
	poolCfg.ConnConfig.ConnectTimeout = 10 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return fmt.Errorf("postgres: create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return fmt.Errorf("postgres: ping: %w", err)
	}

	d.db = pool
	d.Options = opts

	if cfg.Postgres.CreateTables == true {
		if err := migrate(ctx, pool); err != nil {
			pool.Close()
			return fmt.Errorf("postgres: migrate: %w", err)
		}
	}

	d.buf = coalesce.New(coalesce.Config{
		Interval:  cfg.FlushInterval,
		Threshold: cfg.FlushThreshold,
		Name:      "postgres",
		Logger:    opts.Logger,
	}, d.applyUpdates)
	d.buf.Start()

	return nil
}

func (d *postgresDB) Disconnect() error {
	if d.db == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := d.buf.Stop(ctx); err != nil && d.Logger != nil {
		d.Logger.Error("postgres: final flush error", "error", err)
	}

	d.db.Close()
	return nil
}

// SyncToDisk drains the coalescing buffer. Postgres data is durable the moment
// it commits, so there is nothing to checkpoint — but the buffer still holds
// updates that have not reached a transaction yet, and callers (the sync cron,
// and the migrator between version replays) rely on this to force them out.
func (d *postgresDB) SyncToDisk(ctx context.Context) error {
	// The sync and GC crons are scheduled before Connect runs, so a tick can
	// land while the connection is still being established.
	if d.db == nil {
		return database.ErrNotAvailable
	}

	return d.observe(ctx, "postgres SyncToDisk", func(ctx context.Context) error {
		return d.buf.Flush(ctx)
	})
}

// applyUpdates commits a batch of coalesced character updates in one
// transaction. It is the ApplyFunc handed to the buffer.
func (d *postgresDB) applyUpdates(ctx context.Context, batch []coalesce.Entry) error {
	ctx, span := oida.Start(ctx, "UPDATE characters (flush)", oida.KindDatabase)
	defer span.End()
	span.SetAttribute("characters", len(batch))

	// The buffer hands batches over already sorted by ID, so row locks are
	// always acquired in the same order and overlapping flushes can't deadlock.
	err := d.execTx(ctx, func(tx pgx.Tx) error {
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
func (d *postgresDB) RunGC(ctx context.Context) error {
	if d.db == nil {
		return database.ErrNotAvailable
	}

	return d.observe(ctx, "postgres RunGC", func(ctx context.Context) error {
		if err := d.buf.Flush(ctx); err != nil && d.Logger != nil {
			d.Logger.Warn("postgres: flush before GC failed, collecting anyway", "error", err)
		}

		ctx, span := oida.Start(ctx, "DELETE expired characters", oida.KindDatabase)
		defer span.End()

		tag, err := d.db.Exec(ctx,
			`DELETE FROM characters WHERE expires_at IS NOT NULL AND expires_at <= NOW()`,
		)
		if err != nil {
			span.RecordError(err)
			return err
		}

		span.SetAttribute("deleted", tag.RowsAffected())
		return nil
	})
}

// observe runs fn inside its own trace so background work that has no request
// behind it still shows up on the dashboard. Without a tracer it just runs fn.
func (d *postgresDB) observe(ctx context.Context, name string, fn func(context.Context) error) error {
	if d.Tracer == nil {
		return fn(ctx)
	}

	return d.Tracer.Observe(ctx, name, fn)
}

// execTx runs fn inside a transaction. Postgres supports multiple concurrent
// writers, so there is no need for a serialized write channel.
func (d *postgresDB) execTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := d.db.Begin(ctx)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}

// migrate creates the schema on first run. Uses Postgres-native types.
func migrate(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id         TEXT PRIMARY KEY,
			revision   INTEGER NOT NULL DEFAULT 0,
			flags      INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS characters (
			id              UUID PRIMARY KEY,
			steam_id        TEXT REFERENCES users(id),
			slot            INTEGER,
			created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			deleted_at      TIMESTAMPTZ,
			expires_at      TIMESTAMPTZ,
			data_created_at TIMESTAMPTZ,
			data_size       INTEGER NOT NULL DEFAULT 0,
			data_payload    TEXT NOT NULL DEFAULT '',
			UNIQUE (steam_id, slot)
		);

		CREATE TABLE IF NOT EXISTS deleted_characters (
			steam_id     TEXT NOT NULL REFERENCES users(id),
			slot         INTEGER NOT NULL,
			character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
			deleted_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			PRIMARY KEY (steam_id, slot),
			UNIQUE (character_id)
		);

		CREATE TABLE IF NOT EXISTS character_versions (
			id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
			created_at   TIMESTAMPTZ NOT NULL,
			size         INTEGER NOT NULL,
			data_payload TEXT NOT NULL
		);

		CREATE INDEX IF NOT EXISTS idx_chars_steam_id   ON characters(steam_id);
		CREATE INDEX IF NOT EXISTS idx_charver_char_id  ON character_versions(character_id);
	`)
	return err
}

// pgErr is a helper to check for specific Postgres error codes if needed.
func pgErr(err error) *pgconn.PgError {
	var pgError *pgconn.PgError
	if err != nil {
		if ok := pgx.ErrNoRows; err == ok {
			return nil
		}
		if e, ok := err.(*pgconn.PgError); ok {
			return e
		}
	}
	_ = pgError
	return nil
}
