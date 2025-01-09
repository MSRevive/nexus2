package sqlite

import (
	//"fmt"
	
	"github.com/msrevive/nexus2/internal/bitmask"
	//"github.com/msrevive/nexus2/internal/database"
	"github.com/msrevive/nexus2/pkg/database/bsoncoder"
	"github.com/msrevive/nexus2/pkg/database/schema"
)

func (d *sqliteDB) GetAllUsers() ([]*schema.User, error) {
	var users []*schema.User

	return users, nil
}

func (d *sqliteDB) GetUser(steamid string) (user *schema.User, err error) {
	return
}

func (d *sqliteDB) SetUserFlags(steamid string, flags bitmask.Bitmask) (error) {
	return nil
}

func (d *sqliteDB) GetUserFlags(steamid string) (bitmask.Bitmask, error) {
	return bitmask.Bitmask(user.Flags), nil
}