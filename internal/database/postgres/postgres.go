package postgres

import (
	"context"
	"fmt"
	"sync"
	"time"
	"sort"

	"github.com/msrevive/nexus2/internal/database"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// pendingUpdate holds the latest state for a character that has been
// updated but not yet flushed to the database.
type pendingUpdate struct {
	size       int
	data       string
	backupMax  int
	backupTime time.Duration
}

type postgresDB struct {
	db *pgxpool.Pool

	// flushInterval controls how often the coalescing buffer is drained.
	flushInterval time.Duration

	// pendingUpdates is the coalescing map. When UpdateCharacter is called,
	// we just overwrite the entry for that character ID. On each flush tick,
	// all pending entries are committed in a single transaction.
	coalesceMu     sync.RWMutex
	pendingUpdates map[uuid.UUID]pendingUpdate

	done chan struct{}
	wg   sync.WaitGroup

	database.Options
}

func New() *postgresDB {
	return &postgresDB{
		flushInterval:  3 * time.Second,
		pendingUpdates: make(map[uuid.UUID]pendingUpdate),
		done:           make(chan struct{}),
	}
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
	d.Logger = opts.Logger

	if cfg.Postgres.CreateTables == true {
		if err := migrate(ctx, pool); err != nil {
			pool.Close()
			return fmt.Errorf("postgres: migrate: %w", err)
		}
	}

	d.wg.Add(1)
	go d.flushWorker()

	return nil
}

func (d *postgresDB) Disconnect() error {
	close(d.done)
	d.wg.Wait()
	d.db.Close()
	return nil
}

// SyncToDisk is a no-op for Postgres — data is durable after COMMIT.
func (d *postgresDB) SyncToDisk() error {
	return nil
}

// RunGC flushes pending updates and then purges any soft-deleted characters
// whose expiration timestamp has passed.
func (d *postgresDB) RunGC() error {
	ctx := context.Background()
	_, err := d.db.Exec(ctx,
		`DELETE FROM characters WHERE expires_at IS NOT NULL AND expires_at <= NOW()`,
	)
	return err
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

// flushWorker ticks on flushInterval and drains the coalescing buffer.
// On shutdown it performs one final flush so no updates are lost.
func (d *postgresDB) flushWorker() {
	defer d.wg.Done()
	ticker := time.NewTicker(d.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := d.flushPendingUpdates(); err != nil && d.Logger != nil {
				d.Logger.Fatalln("postgres: flush error", "error", err)
			}

		case <-d.done:
			_ = d.flushPendingUpdates()
			return
		}
	}
}

// flushPendingUpdates atomically swaps the coalescing map for a fresh one,
// then commits all coalesced updates in a single transaction.
func (d *postgresDB) flushPendingUpdates() error {
	d.coalesceMu.Lock()
	if len(d.pendingUpdates) == 0 {
		d.coalesceMu.Unlock()
		return nil
	}
	snapshot := d.pendingUpdates
	d.pendingUpdates = make(map[uuid.UUID]pendingUpdate)
	d.coalesceMu.Unlock()

	// Sort IDs to acquire row locks in a consistent order and prevent deadlocks.
	ids := make([]uuid.UUID, 0, len(snapshot))
	for id := range snapshot {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		return ids[i].String() < ids[j].String()
	})

	ctx := context.Background()
	return d.execTx(ctx, func(tx pgx.Tx) error {
		for _, id := range ids {
			if err := applyCharacterUpdate(ctx, tx, id, snapshot[id]); err != nil {
				return fmt.Errorf("flush update for %s: %w", id, err)
			}
		}
		return nil
	})
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
