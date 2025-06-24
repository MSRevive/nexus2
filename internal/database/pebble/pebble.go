package pebble

import (
	"github.com/msrevive/nexus2/internal/database"

	"github.com/cockroachdb/pebble/v2"
)

// The smaller the key prefix the better?
var (
	UserPrefix = []byte("users:")
	CharPrefix = []byte("chars:")
)

type pebbleDB struct {
	db *pebble.DB
}

func New() *pebbleDB {
	return &pebbleDB{}
}

func (d *pebbleDB) Connect(cfg database.Config, opts database.Options) error {
	db, err := pebble.Open(cfg.Badger.Directory, &pebble.Options{
		FormatMajorVersion: pebble.FormatDefault
	})
	if err != nil {
		return nil, err
	}

	d.db = db
	return nil
}

func (d *pebbleDB) Disconnect() error {
	return d.db.Close()
}

func (d *pebbleDB) SyncToDisk() error {
	//return d.db.Flush()
	return nil
}

func (d *pebbleDB) RunGC() error {
	return nil
}