package badger

import (
	"errors"

	"github.com/msrevive/nexus2/internal/database"

	"github.com/dgraph-io/badger/v4"
)

var (
	ErrNoDocument = errors.New("no document")

	UserPrefix = []byte("users:")
	CharPrefix = []byte("characters:")
)

type badgerDB struct {
	db *badger.DB
}

func New() *badgerDB {
	return &badgerDB{}
}

func (d *badgerDB) Connect(cfg database.Config, opts database.Options) error {
	dOpts := badger.DefaultOptions(cfg.Badger.Directory)
	// opts.ValueLogFileSize = 256 * 1024 * 1024 // 256 MB
	// opts.LevelFanout = 10
	// opts.LevelFringeSize = 100
	// opts.BloomFalsePositive = 0.01
	// opts.Compression = badger.Snappy
	// opts.SyncWrites = false
	opts.Logger = opts.Logger

	db, err := badger.Open(dOpts)
	if err != nil {
		return err
	}

	d.db = db
	return nil
}

func (d *badgerDB) Disconnect() error {
	return d.db.Close()
}

func (d *badgerDB) SaveToDatabase() error {
	return nil
}

func (d *badgerDB) ClearCache() {
	return
}