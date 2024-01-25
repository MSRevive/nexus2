package bbolt

import (
	"fmt"
	"time"
	//"encoding/gob"
	
	"github.com/msrevive/nexus2/internal/database"
	"github.com/msrevive/nexus2/internal/database/schema"

	"github.com/google/uuid"
	"go.etcd.io/bbolt"
)

type bboltDB struct {
	db *bbolt.DB
}

func New() *bboltDB {
	return &bboltDB{}
}

func (d *bboltDB) Connect(cfg database.Config) error {
	timeout := cfg.BBolt.Timeout * time.Second
	db, err := bbolt.Open(cfg.BBolt.File, 0755, &bbolt.Options{Timeout: timeout})
	if err != nil {
		return err
	}

	if err := db.Update(func(tx *bbolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists([]byte("users"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		_, err = tx.CreateBucketIfNotExists([]byte("characters"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
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

func (d *bboltDB) NewCharacter(steamid string, slot int, size int, data string) (uuid.UUID, error) {
	if err := d.db.Update(func(tx *bbolt.Tx) error {
		return nil
	}); err != nil {
		return uuid.Nil, err
	}

	return uuid.Nil, nil
}

func (d *bboltDB) UpdateCharacter(id uuid.UUID, size int, data string, backupMax int, backupTime time.Duration) error {
	return nil
}

func (d *bboltDB) GetUser(steamid string) (user *schema.User, err error) {
	return &schema.User{}, nil
}

func (d *bboltDB) GetCharacter(id uuid.UUID) (char *schema.Character, err error) {
	return &schema.Character{}, nil
}

func (d *bboltDB) GetCharacters(steamid string) (map[int]schema.Character, error) {
	return make(map[int]schema.Character), nil
}

func (d *bboltDB) LookUpCharacterID(steamid string, slot int) (uuid.UUID, error) {
	return uuid.Nil, nil
}

func (d *bboltDB) SoftDeleteCharacter(id uuid.UUID) error {
	return nil
}

func (d *bboltDB) DeleteCharacter(id uuid.UUID) error {
	return nil
}

func (d *bboltDB) DeleteCharacterReference(steamid string, slot int) error {
	return nil
}

func (d *bboltDB) MoveCharacter(id uuid.UUID, steamid string, slot int) error {
	return nil
}

func (d *bboltDB) CopyCharacter(id uuid.UUID, steamid string, slot int) (uuid.UUID, error) {
	return uuid.Nil, nil
}

func (d *bboltDB) RestoreCharacter(id uuid.UUID) error {
	return nil
}

func (d *bboltDB) RollbackCharacter(id uuid.UUID, ver int) error {
	return nil
}

func (d *bboltDB) RollbackCharacterToLatest(id uuid.UUID) error {
	return nil
}

func (d *bboltDB) DeleteCharacterVersions(id uuid.UUID) error {
	return nil
}

func (d *bboltDB) SaveToDatabase() error {
	return nil
}

func (d *bboltDB) ClearCache() {
	
}