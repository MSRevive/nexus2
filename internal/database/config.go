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
	Sync string
	GarbageCollection string
}