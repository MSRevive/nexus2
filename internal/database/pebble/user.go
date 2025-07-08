package pebble

import (
	"fmt"
	"context"
	"time"
	
	"github.com/msrevive/nexus2/internal/bitmask"
	"github.com/msrevive/nexus2/internal/database"
	"github.com/msrevive/nexus2/pkg/database/schema"

	"github.com/cockroachdb/pebble/v2"
	"github.com/fxamacker/cbor/v2"
)

func (d *pebbleDB) GetAllUsers() ([]*schema.User, error) {
	var users []*schema.User

	// better to use contexts https://github.com/cockroachdb/pebble/issues/1609
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	it, err := d.db.NewIterWithContext(ctx, &pebble.IterOptions{
		LowerBound: UserPrefix,
		UpperBound: keyUpperBound(UserPrefix),
	})

	if err != nil {
		it.Close()
		return nil, err
	}
	defer it.Close()

	for it.First(); it.Valid(); it.Next() {
		var user *schema.User
		if err := cbor.Unmarshal(it.Value(), &user); err != nil {
			return nil, fmt.Errorf("failed to unmarshal %v", err)
		}

		users = append(users, user)
	}

	return users, nil
}

func (d *pebbleDB) GetUser(steamid string) (*schema.User, error) {
	var user *schema.User = nil
	key := append(UserPrefix, []byte(steamid)...)

	data, io, err := d.db.Get(key)
	if err == pebble.ErrNotFound {
		return user, database.ErrNoDocument
	}else if err != nil {
		return user, err
	}

	defer io.Close()

	if err := cbor.Unmarshal(data, &user); err != nil {
		return user, fmt.Errorf("failed to unmarshal %v", err)
	}

	return user, nil
}

func (d *pebbleDB) SetUserFlags(steamid string, flags bitmask.Bitmask) (error) {
	user, err := d.GetUser(steamid)
	if err != nil {
		return err
	}

	user.Flags = uint32(flags)

	userData, err := cbor.Marshal(&user)
	if err != nil {
		return fmt.Errorf("failed to marshal user %v", err)
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