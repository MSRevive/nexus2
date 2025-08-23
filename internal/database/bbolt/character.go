package bbolt

import (
	"fmt"
	"time"
	
	"github.com/msrevive/nexus2/internal/database"
	"github.com/msrevive/nexus2/pkg/database/schema"

	"github.com/bwmarrin/snowflake"
	"go.etcd.io/bbolt"
	"github.com/fxamacker/cbor/v2"
)

func (d *bboltDB) NewCharacter(steamid string, slot int, size int, data string) (snowflake.ID, error) {
	charID := uuid.New()
	char := schema.Character{
		ID: charID,
		SteamID: steamid,
		Slot: slot,
		CreatedAt: time.Now().UTC(),
		Data: schema.CharacterData{
			CreatedAt: time.Now().UTC(),
			Size: size,
			Data: data,
		},
	}

	user, err := d.GetUser(steamid)
	if err == database.ErrNoDocument {
		if err = d.db.Update(func(tx *bbolt.Tx) error {
			user = &schema.User{
				ID: steamid,
				Characters: make(map[int]snowflake.ID),
			}
			user.Characters[slot] = charID

			userData, err := cbor.Marshal(&user)
			if err != nil {
				return fmt.Errorf("failed to marshal user %v", err)
			}

			charData, err := cbor.Marshal(&char)
			if err != nil {
				return fmt.Errorf("failed to marshal character %v", err)
			}

			bUser := tx.Bucket(UserBucket)
			bChar := tx.Bucket(CharBucket)

			if err := bUser.Put([]byte(steamid), userData); err != nil {
				return fmt.Errorf("bbolt: failed to put in users", err)
			}

			if err := bChar.Put([]byte(charID.String()), charData); err != nil {
				return fmt.Errorf("bbolt: failed to put in characters", err)
			}
	
			return nil
		}); err != nil {
			return snowflake.ID, err
		}
	}else if err != nil {
		return snowflake.ID, err
	}

	if user != nil {
		if err = d.db.Update(func(tx *bbolt.Tx) error {
			user.Characters[slot] = charID
	
			userData, err := cbor.Marshal(&user)
			if err != nil {
				return fmt.Errorf("failed to marshal user %v", err)
			}
	
			charData, err := cbor.Marshal(&char)
			if err != nil {
				return fmt.Errorf("failed to marshal character %v", err)
			}
	
			bUser := tx.Bucket(UserBucket)
			bChar := tx.Bucket(CharBucket)
	
			if err := bUser.Put([]byte(steamid), userData); err != nil {
				return fmt.Errorf("bbolt: failed to put in users", err)
			}
	
			if err := bChar.Put([]byte(charID.String()), charData); err != nil {
				return fmt.Errorf("bbolt: failed to put in characters", err)
			}
	
			return nil
		}); err != nil {
			return snowflake.ID, err
		}
	}

	return charID, nil
}

func (d *bboltDB) UpdateCharacter(id snowflake.ID, size int, data string, backupMax int, backupTime time.Duration) error {
	if err := d.db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket(CharBucket)

		// Get the character data.
		item := b.Get([]byte(id.String()))
		if len(item) == 0 {
			return database.ErrNoDocument
		}

		// Decode the data
		var char *schema.Character
		if err := cbor.Unmarshal(item, &char); err != nil {
			return fmt.Errorf("failed to unmarshal %v", err)
		}

		// Handle backups for characters
		bCharsLen := len(char.Versions)
		if backupMax > 0 {
			// we remove the oldest backup here
			if bCharsLen >= backupMax {
				copy(char.Versions, char.Versions[1:])
				char.Versions = char.Versions[:bCharsLen-1]
				bCharsLen--
			}

			if bCharsLen > 0 {
				bNewest := char.Versions[bCharsLen-1] //latest backup

				timeCheck := bNewest.CreatedAt.Add(backupTime)
				if char.Data.CreatedAt.After(timeCheck) {
					char.Versions = append(char.Versions, char.Data)
				}
			}else{
				char.Versions = append(char.Versions, char.Data)
			}
		}

		char.Data = schema.CharacterData{
			CreatedAt: time.Now().UTC(), 
			Size: size, 
			Data: data,
		}

		charData, err := cbor.Marshal(&char)
		if err != nil {
			return fmt.Errorf("bson: failed to encode character %v", err)
		}

		if err := b.Put([]byte(char.ID.String()), charData); err != nil {
			return fmt.Errorf("bbolt: failed to update character %v", err)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (d *bboltDB) GetCharacter(id snowflake.ID) (char *schema.Character, err error) {
	if err = d.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(CharBucket)

		data := b.Get([]byte(id.String()))
		if len(data) == 0 {
			return database.ErrNoDocument
		}

		if err := cbor.Unmarshal(data, &char); err != nil {
			return fmt.Errorf("bson: failed to decode %v", err)
		}

		return nil
	}); err != nil {
		return
	}

	return
}

func (d *bboltDB) GetCharacters(steamid string) (map[int]schema.Character, error) {
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

func (d *bboltDB) LookUpCharacterID(steamid string, slot int) (snowflake.ID, error) {
	user, err := d.GetUser(steamid)
	if err != nil {
		return 0, err
	}

	uuid := user.Characters[slot]
	return uuid, nil
}

func (d *bboltDB) SoftDeleteCharacter(id snowflake.ID, expiration time.Duration) error {
	char, err := d.GetCharacter(id)
	if err != nil {
		return err
	}

	user, err := d.GetUser(char.SteamID)
	if err != nil {
		return err
	}

	delete(user.Characters, char.Slot)
	user.DeletedCharacters = make(map[int]snowflake.ID, 1)
	user.DeletedCharacters[char.Slot] = id

	timeNow := time.Now().UTC()
	char.DeletedAt = &timeNow

	if err = d.db.Update(func(tx *bbolt.Tx) error {
		userData, err := cbor.Marshal(&user)
		if err != nil {
			return fmt.Errorf("bson: failed to encode user %v", err)
		}

		charData, err := cbor.Marshal(&char)
		if err != nil {
			return fmt.Errorf("bson: failed to encode character %v", err)
		}

		bUser := tx.Bucket(UserBucket)
		bChar := tx.Bucket(CharBucket)

		if err := bUser.Put([]byte(char.SteamID), userData); err != nil {
			return fmt.Errorf("bbolt: failed to update user %v", err)
		}

		if err := bChar.Put([]byte(id.String()), charData); err != nil {
			return fmt.Errorf("bbolt: failed to update character %v", err)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (d *bboltDB) DeleteCharacter(id snowflake.ID) error {
	if err := d.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(CharBucket)

		if err := b.Delete([]byte(id.String())); err != nil {
			return fmt.Errorf("bbolt: failed to delete character %v", err)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (d *bboltDB) DeleteCharacterReference(steamid string, slot int) error {
	user, err := d.GetUser(steamid)
	if err != nil {
		return err
	}

	delete(user.Characters, slot)

	if err = d.db.Update(func(tx *bbolt.Tx) error {
		userData, err := cbor.Marshal(&user)
		if err != nil {
			return fmt.Errorf("bson: failed to encode user %v", err)
		}

		b := tx.Bucket(UserBucket)

		if err := b.Put([]byte(steamid), userData); err != nil {
			return fmt.Errorf("bbolt: failed to update user %v", err)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (d *bboltDB) MoveCharacter(id snowflake.ID, steamid string, slot int) error {
	user, err := d.GetUser(steamid)
	if err != nil {
		return err
	}

	char, err := d.GetCharacter(id)
	if err != nil {
		return err
	}

	// Delete reference to the character from via old user
	if err := d.DeleteCharacterReference(char.SteamID, char.Slot); err != nil {
		return err
	}

	// Assign character ID to the new account.
	user.Characters[slot] = id

	// Update the character information with new account data.
	char.SteamID = steamid
	char.Slot = slot

	if err = d.db.Update(func(tx *bbolt.Tx) error {
		userData, err := cbor.Marshal(&user)
		if err != nil {
			return fmt.Errorf("bson: failed to encode user %v", err)
		}

		charData, err := cbor.Marshal(&char)
		if err != nil {
			return fmt.Errorf("bson: failed to encode character %v", err)
		}

		bUser := tx.Bucket(UserBucket)
		bChar := tx.Bucket(CharBucket)

		if err := bUser.Put([]byte(steamid), userData); err != nil {
			return fmt.Errorf("bbolt: failed to update user %v", err)
		}

		if err := bChar.Put([]byte(id.String()), charData); err != nil {
			return fmt.Errorf("bbolt: failed to update character %v", err)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (d *bboltDB) CopyCharacter(id snowflake.ID, steamid string, slot int) (snowflake.ID, error) {
	// Create reference to "new" character.
	targetUser, err := d.GetUser(steamid)
	if err != nil {
		return 0, err
	}

	// Insert new character data.
	char, err := d.GetCharacter(id)
	if err != nil {
		return 0, err
	}

	charID := uuid.New()
	targetUser.Characters[slot] = charID

	char.ID = charID
	char.SteamID = steamid
	char.Slot = slot
	char.CreatedAt = time.Now().UTC()

	if err = d.db.Update(func(tx *bbolt.Tx) error {
		userData, err := cbor.Marshal(&targetUser)
		if err != nil {
			return fmt.Errorf("bson: failed to encode user %v", err)
		}

		charData, err := cbor.Marshal(&char)
		if err != nil {
			return fmt.Errorf("bson: failed to encode character %v", err)
		}

		bUser := tx.Bucket(UserBucket)
		bChar := tx.Bucket(CharBucket)

		if err := bUser.Put([]byte(steamid), userData); err != nil {
			return fmt.Errorf("bbolt: failed to update user %v", err)
		}

		if err := bChar.Put([]byte(charID.String()), charData); err != nil {
			return fmt.Errorf("bbolt: failed to update character %v", err)
		}

		return nil
	}); err != nil {
		return 0, err
	}

	return charID, nil
}

func (d *bboltDB) RestoreCharacter(id snowflake.ID) error {
	char, err := d.GetCharacter(id)
	if err != nil {
		return err
	}

	user, err := d.GetUser(char.SteamID)
	if err != nil {
		return err
	}

	user.Characters[char.Slot] = id
	delete(user.DeletedCharacters, char.Slot)

	char.DeletedAt = nil

	if err = d.db.Update(func(tx *bbolt.Tx) error {
		userData, err := cbor.Marshal(&user)
		if err != nil {
			return fmt.Errorf("bson: failed to encode user %v", err)
		}

		charData, err := cbor.Marshal(&char)
		if err != nil {
			return fmt.Errorf("bson: failed to encode character %v", err)
		}

		bUser := tx.Bucket(UserBucket)
		bChar := tx.Bucket(CharBucket)

		if err := bUser.Put([]byte(char.SteamID), userData); err != nil {
			return fmt.Errorf("bbolt: failed to update user %v", err)
		}

		if err := bChar.Put([]byte(id.String()), charData); err != nil {
			return fmt.Errorf("bbolt: failed to update character %v", err)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (d *bboltDB) RollbackCharacter(id snowflake.ID, ver int) error {
	char, err := d.GetCharacter(id)
	if err != nil {
		return err
	}

	bCharsLen := len(char.Versions)
	if bCharsLen > ver {
		// Replace the active character with the selected version
		char.Data = char.Versions[ver]
	}else{
		return fmt.Errorf("no character version at index %d", ver)
	}

	if err = d.db.Update(func(tx *bbolt.Tx) error {
		charData, err := cbor.Marshal(&char)
		if err != nil {
			return fmt.Errorf("bson: failed to encode character %v", err)
		}

		b := tx.Bucket(CharBucket)

		if err := b.Put([]byte(id.String()), charData); err != nil {
			return fmt.Errorf("bbolt: failed to update char %v", err)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (d *bboltDB) RollbackCharacterToLatest(id snowflake.ID) error {
	char, err := d.GetCharacter(id)
	if err != nil {
		return err
	}

	bCharsLen := len(char.Versions)
	if bCharsLen > 0 {
		// Replace the active character with the selected version
		char.Data = char.Versions[bCharsLen-1]
	}else{
		return fmt.Errorf("no character backups exist")
	}

	if err = d.db.Update(func(tx *bbolt.Tx) error {
		charData, err := cbor.Marshal(&char)
		if err != nil {
			return fmt.Errorf("bson: failed to encode character %v", err)
		}

		b := tx.Bucket(CharBucket)

		if err := b.Put([]byte(id.String()), charData); err != nil {
			return fmt.Errorf("bbolt: failed to update char %v", err)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (d *bboltDB) DeleteCharacterVersions(id snowflake.ID) error {
	char, err := d.GetCharacter(id)
	if err != nil {
		return err
	}

	char.Versions = nil

	if err = d.db.Update(func(tx *bbolt.Tx) error {
		charData, err := cbor.Marshal(&char)
		if err != nil {
			return fmt.Errorf("bson: failed to encode char %v", err)
		}

		b := tx.Bucket(CharBucket)

		if err := b.Put([]byte(id.String()), charData); err != nil {
			return fmt.Errorf("bbolt: failed to update char %v", err)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}