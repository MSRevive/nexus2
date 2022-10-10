package config

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

const sqliteDriver = "sqlite3"

func InitializeDb(conn string) (*sqlx.DB, error) {
	db, err := sql.Open(sqliteDriver, conn)
	if err != nil {
		return nil, err
	}

	return db, createTables(db)
}

func createTables(db *sqlx.DB) error {
	stmt := `
	CREATE TABLE IF NOT EXISTS characters
	CREATE TABLE IF NOT EXISTS nerds
	`
	if _, err := db.ExecContext(context.Background(), stmt); err != nil {
		return err
	}

	return nil
}
