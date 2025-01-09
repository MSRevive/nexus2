package sqlite

import (
	"fmt"
	"time"
	"context"
	"embed"
	
	"github.com/msrevive/nexus2/internal/database"

	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

//go:embed sqlite.sql

type sqliteDB struct {
	db *sqlite.Conn
}

func New() *sqliteDB {
	return &sqliteDB{}
}

func (d *sqliteDB) Connect(cfg database.Config, opts database.Options) error {
	db, err := sqlite.OpenConn(cfg.SQLite.Conn, 0)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.SQLite.Timeout*time.Millisecond)
	defer cancel()
	db.SetInterrupt(ctx.Done())

	if err := sqlitex.ExecuteScriptFS(db, embed.FS, "sqlite.sql", nil); err != nil {
		return err
	}

	return nil
}

func (d *sqliteDB) Disconnect() error {
	return d.db.Close()
}

func (d *sqliteDB) SyncToDisk() error {
	return nil
}

func (d *sqliteDB) RunGC() error {
	return nil
}