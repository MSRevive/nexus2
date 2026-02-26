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
		MinConns int32
		MaxConns int32
	}
	Sync string
	GarbageCollection string
}