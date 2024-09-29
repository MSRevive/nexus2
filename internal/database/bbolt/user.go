package bbolt

import (
	"fmt"
	
	"github.com/msrevive/nexus2/internal/bitmask"
	"github.com/msrevive/nexus2/pkg/database/bsoncoder"
	"github.com/msrevive/nexus2/pkg/database/schema"

	"go.etcd.io/bbolt"
)

func (d *bboltDB) GetAllUsers() ([]*schema.User, error) {
	var users []*schema.User

	if err := d.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(UserBucket)

		if err := b.ForEach(func(k, v []byte) error {
			var user *schema.User

			if err := bsoncoder.Decode(v, &user); err != nil {
				return fmt.Errorf("bson: failed to unmarshal %v", err)
			}

			users.append(user)
		}); err != nil {
			return err
		}

		return users, nil
	}); err != nil {
		return users, err
	}

	return users, nil
}

func (d *bboltDB) GetUser(steamid string) (user *schema.User, err error) {
	if err = d.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(UserBucket)

		data := b.Get([]byte(steamid))
		if len(data) == 0 {
			return ErrNoDocument
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

	if err = d.db.Update(func(tx *bbolt.Tx) error {
		user.Flags = uint32(flags) // cast it to a uint32 to make the database behave.

		userData, err := bsoncoder.Encode(&user)
		if err != nil {
			return fmt.Errorf("bson: failed to marshal user %v", err)
		}

		bUser := tx.Bucket(UserBucket)

		if err := bUser.Put([]byte(steamid), userData); err != nil {
			return fmt.Errorf("bbolt: failed to put in users", err)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (d *bboltDB) GetUserFlags(steamid string) (bitmask.Bitmask, error) {
	user, err := d.GetUser(steamid)
	if err != nil {
		return 0, err
	}

	return bitmask.Bitmask(user.Flags), nil
}