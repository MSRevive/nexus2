package bbolt

import (
	"fmt"
	
	"github.com/msrevive/nexus2/pkg/database/bsoncoder"
	"github.com/msrevive/nexus2/pkg/database/schema"

	"go.etcd.io/bbolt"
)

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

func (d *bboltDB) SetUserFlags(steamid string, flag uint32) (error) {
	user, err := d.GetUser(steamid)
	if err != nil {
		return err
	}

	if err = d.db.Update(func(tx *bbolt.Tx) error {
		user.Flags = flag

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

func (d *bboltDB) GetUserFlags(steamid string) (uint32, error) {
	user, err := d.GetUser(steamid)
	if err != nil {
		return 0, err
	}

	return user.Flags, nil
}