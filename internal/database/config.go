package database

import (
	"time"
)

type Config struct {
	SQLite struct {
		Path string
	}
	Postgres struct {
		Conn string
		MinConns int32
		MaxConns int32
		RetryDelay time.Duration
		MaxRetries int
	}
	Sync string
	GarbageCollection string
}