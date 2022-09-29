package config

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

const sqliteDriver = "sqlite3"

func InitializeDb(conn string) (*sqlx.DB, error) {
	db, err := sql.Open(sqliteDriver, conn)
	if err != nil {
		return nil, err
	}

	return db, nil
}
