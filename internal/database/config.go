package database

import (
	"time"
)

type Config struct {
	MongoDB struct {
		Connection string
	}
	BBolt struct {
		File string
		Timeout time.Duration
	}
	Badger struct {
		Directory string
	}
}