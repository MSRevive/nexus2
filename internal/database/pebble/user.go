package pebble

import (
	"fmt"
	
	"github.com/msrevive/nexus2/internal/bitmask"
	"github.com/msrevive/nexus2/internal/database"
	"github.com/msrevive/nexus2/pkg/database/bsoncoder"
	"github.com/msrevive/nexus2/pkg/database/schema"

	"github.com/cockroachdb/pebble/v2"
)

// TODO
func (d *pebbleDB) GetAllUsers() ([]*schema.User, error) {
	var users []*schema.User

	/*
	if err := d.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		
		for it.Seek(UserPrefix); it.ValidForPrefix(UserPrefix); it.Next() {
			item := it.Item()

			item.Value(func(v []byte) error {
				var user *schema.User
				if err := bsoncoder.Decode(v, &user); err != nil {
					return fmt.Errorf("bson: failed to unmarshal %v", err)
				}

				users = append(users, user)
				return nil
			})
		}

		return nil
	}); err != nil {
		return nil, err
	}

	it, err := d.db.NewIter(&pebble.IterOptions{
		LowerBound: UserPrefix,
		UpperBound: CharPrefix,
	})
	if err != nil {
		it.Close()
		return nil, err
	}*/



	return users, nil
}

func (d *pebbleDB) GetUser(steamid string) (*schema.User, error) {
	var user *schema.User = nil
	key := append(UserPrefix, []byte(steamid)...)

	data, io, err := p.db.Get(key)
	if err == pebble.ErrNotFound {
		return user, database.ErrNoDocument
	}else if err != nil {
		return user, err
	}

	defer io.Close()

	if err := bsoncoder.Decode(data, &user); err != nil {
		return user, fmt.Errorf("bson: failed to unmarshal %v", err)
	}

	return user, nil
}

func (d *pebbleDB) SetUserFlags(steamid string, flags bitmask.Bitmask) (error) {
	user, err := d.GetUser(steamid)
	if err != nil {
		return err
	}

	user.Flags = uint32(flags)

	userData, err := bsoncoder.Encode(&user)
	if err != nil {
		return fmt.Errorf("bson: failed to marshal user %v", err)
	}

	key := append(UserPrefix, []byte(steamid)...)
	return d.db.Set(key, userData, pebble.NoSync)
}

func (d *pebbleDB) GetUserFlags(steamid string) (bitmask.Bitmask, error) {
	user, err := d.GetUser(steamid)
	if err != nil {
		return 0, err
	}

	return bitmask.Bitmask(user.Flags), nil
}