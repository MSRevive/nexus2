package sqlite

import (
	//"fmt"
	"time"
	
	//"github.com/msrevive/nexus2/internal/database"
	//"github.com/msrevive/nexus2/pkg/database/bsoncoder"
	"github.com/msrevive/nexus2/pkg/database/schema"

	"github.com/google/uuid"
)

func (d *sqliteDB) NewCharacter(steamid string, slot int, size int, data string) (uuid.UUID, error) {
	charID := uuid.New()
	return charID, nil
}

func (d *sqliteDB) UpdateCharacter(id uuid.UUID, size int, data string, backupMax int, backupTime time.Duration) error {
	return nil
}

func (d *sqliteDB) GetCharacter(id uuid.UUID) (char *schema.Character, err error) {
	return
}

func (d *sqliteDB) GetCharacters(steamid string) (map[int]schema.Character, error) {
	user, err := d.GetUser(steamid)
	if err != nil {
		return nil, err
	}
	
	chars := make(map[int]schema.Character, len(user.Characters)-1)
	for k,v := range user.Characters {
		char, err := d.GetCharacter(v)
		if err != nil {
			return nil, err
		}
		chars[k] = *char
	}

	return chars, nil
}

func (d *sqliteDB) LookUpCharacterID(steamid string, slot int) (uuid.UUID, error) {
	user, err := d.GetUser(steamid)
	if err != nil {
		return uuid.Nil, err
	}

	uuid := user.Characters[slot]
	return uuid, nil
}

// We remove the character from user's active list and set an expiration.
func (d *sqliteDB) SoftDeleteCharacter(id uuid.UUID, expiration time.Duration) error {
	return nil, 0
}

func (d *sqliteDB) DeleteCharacter(id uuid.UUID) error {
	return nil
}

func (d *sqliteDB) DeleteCharacterReference(steamid string, slot int) error {
	return nil
}

func (d *sqliteDB) MoveCharacter(id uuid.UUID, steamid string, slot int) error {
	return nil
}

func (d *sqliteDB) CopyCharacter(id uuid.UUID, steamid string, slot int) (uuid.UUID, error) {
	return nil, nil
}

func (d *sqliteDB) RestoreCharacter(id uuid.UUID) error {
	return nil
}

func (d *sqliteDB) RollbackCharacter(id uuid.UUID, ver int) error {
	return nil
}

func (d *sqliteDB) RollbackCharacterToLatest(id uuid.UUID) error {
	return nil
}

func (d *sqliteDB) DeleteCharacterVersions(id uuid.UUID) error {
	return nil
}