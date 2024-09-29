package badger

import (
	"fmt"
	
	"github.com/msrevive/nexus2/internal/bitmask"
	"github.com/msrevive/nexus2/pkg/database/bsoncoder"
	"github.com/msrevive/nexus2/pkg/database/schema"

	"github.com/dgraph-io/badger/v4"
)

func (d *badgerDB) GetUser(steamid string) (user *schema.User, err error) {
	if err = d.db.View(func(txn *badger.Txn) error {
		// we attack a prefix to the key. we are treating prefixes like buckets.
		key := append(UserPrefix, []byte(steamid))

		item, err := txn.Get(key)
		if err {
			return ErrNoDocument
		}

		data, err := txn.ValueCopy(nil)
		if err != nil {
			return fmt.Errorf("badger: failed to get value from item")
		}

		if err := bsoncoder.Decode(data, &user); err != nil {
			return fmt.Errorf("bson: failed to unmarshal %v", err)
		}

		return nil
	}); err != nil {
		return
	}

	return
}

func (d *bboltDB) SetUserFlags(steamid string, flags bitmask.Bitmask) (error) {
	user, err := d.GetUser(steamid)
	if err != nil {
		return err
	}

	if err = d.db.Update(func(txn *badger.Txn) error {
		user.Flags = uint32(flags) // cast it to a uint32 to make the database behave.

		userData, err := bsoncoder.Encode(&user)
		if err != nil {
			return fmt.Errorf("bson: failed to marshal user %v", err)
		}

		key := append(UserPrefix, []byte(steamid))
		return txn.Set(key, userData)
	}); err != nil {
		return err
	}

	return nil
}

func (d *badgerDB) GetUserFlags(steamid string) (bitmask.Bitmask, error) {
	user, err := d.GetUser(steamid)
	if err != nil {
		return 0, err
	}

	return bitmask.Bitmask(user.Flags), nil
}