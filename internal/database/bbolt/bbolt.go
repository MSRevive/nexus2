package bbolt

import (
	"fmt"
	"time"
	"errors"
	
	"github.com/msrevive/nexus2/internal/database"

	"go.etcd.io/bbolt"
)

var (
	ErrNoDocument = errors.New("no document")

	UserBucket = []byte("users")
	CharBucket = []byte("characters")
)

type bboltDB struct {
	db *bbolt.DB
}

func New() *bboltDB {
	return &bboltDB{}
}

func (d *bboltDB) Connect(cfg database.Config, opts database.Options) error {
	timeout := cfg.BBolt.Timeout * time.Second
	db, err := bbolt.Open(cfg.BBolt.File, 0755, &bbolt.Options{Timeout: timeout})
	if err != nil {
		return err
	}

	if err := db.Update(func(tx *bbolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists(UserBucket)
		if err != nil {
			return fmt.Errorf("failed to create users bucket: %s", err)
		}

		_, err = tx.CreateBucketIfNotExists(CharBucket)
		if err != nil {
			return fmt.Errorf("failed to create characters bucket: %s", err)
		}

		return nil
	}); err != nil {
		return err
	}

	d.db = db
	return nil
}

func (d *bboltDB) Disconnect() error {
	return d.db.Close()
}

func (d *bboltDB) SyncToDisk() error {
	return nil
}

func (d *bboltDB) RunGC() error {
	return nil
}