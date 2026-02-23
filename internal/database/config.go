package database

import (
)

type Config struct {
	MongoDB struct {
		Connection string
	}
	BBolt struct {
		File string
		Timeout int
	}
	Badger struct {
		Directory string
	}
	Pebble struct {
		Directory string
	}
	SQLite struct {
		Path string
	}
	Sync string
	GarbageCollection string
}