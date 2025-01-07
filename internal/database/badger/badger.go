package badger

import (
	"github.com/msrevive/nexus2/internal/database"

	"github.com/dgraph-io/badger/v4"
	"github.com/dgraph-io/badger/v4/options"
)

// The smaller the key prefix the better?
var (
	UserPrefix = []byte("users:")
	CharPrefix = []byte("chars:")
)

type badgerDB struct {
	db *badger.DB
}

func New() *badgerDB {
	return &badgerDB{}
}

func (d *badgerDB) Connect(cfg database.Config, opts database.Options) error {
	bOpts := badger.DefaultOptions(cfg.Badger.Directory)
	bOpts.WithCompression(options.None)
	bOpts.WithBlockSize(0)
	bOpts.WithBlockCacheSize(0)
	bOpts.WithMemTableSize(4 << 20)

	// opts.MemTableSize = 1 << 20
	// opts.BaseTableSize = 1 << 20
	// //opts.NumCompactors = 2
	// opts.NumLevelZeroTables = 1
	// opts.NumLevelZeroTablesStall = 2
	// opts.BlockCacheSize = 10 << 20
	// //opts.NumMemtables = 1
	// opts.ValueThreshold = 1 << 10
	//bOpts.Logger = opts.Logger

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