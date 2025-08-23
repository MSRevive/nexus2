package badger

import (
	"github.com/msrevive/nexus2/internal/database"

	"github.com/dgraph-io/badger/v4"
	"github.com/bwmarrin/snowflake"
)

// The smaller the key prefix the better?
var (
	UserPrefix = []byte("users:")
	CharPrefix = []byte("chars:")
)

type badgerDB struct {
	db *badger.DB

	node *snowflake.Node
}

func New() *badgerDB {
	return &badgerDB{
		node: snowflake.NewNode(1)
	}
}

func (d *badgerDB) Connect(cfg database.Config, opts database.Options) error {
	bOpts := badger.DefaultOptions(cfg.Badger.Directory)

	db, err := badger.Open(bOpts)
	if err != nil {
		return err
	}

	d.db = db
	return nil
}

func (d *badgerDB) Disconnect() error {
	return d.db.Close()
}

func (d *badgerDB) SyncToDisk() error {
	return d.db.Sync()
}

func (d *badgerDB) RunGC() error {
	return d.db.RunValueLogGC(0.5)
}
