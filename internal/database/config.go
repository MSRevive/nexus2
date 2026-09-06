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
		CreateTables bool
	}
	Sync string
	GarbageCollection string

	// FlushInterval is how long a character update may sit in the coalescing
	// buffer before it is written. It is also the worst-case data loss window
	// on an unclean shutdown. Zero uses the package default (5s).
	FlushInterval time.Duration
	// FlushThreshold is the number of queued characters that triggers a flush
	// without waiting for the interval, which bounds buffer memory under load.
	// Zero uses the package default (256).
	FlushThreshold int
}