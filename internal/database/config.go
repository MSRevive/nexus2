package database

import (
)

type Config struct {
	Pebble struct {
		Directory string
	}
	SQLite struct {
		Path string
	}
	Postgres struct {
		Conn string
	}
	Sync string
	GarbageCollection string
}