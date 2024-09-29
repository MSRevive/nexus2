package badger

import (
	"errors"

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

func (d *badgerDB) Connect(cfg database.Config) error {
	db, err := badger.Open(badger.DefaultOptions(cfg.Badger.Directory))
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