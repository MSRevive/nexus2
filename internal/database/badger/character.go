package badger

import (
	"fmt"
	"time"
	
	"github.com/msrevive/nexus2/pkg/database/bsoncoder"
	"github.com/msrevive/nexus2/pkg/database/schema"
	"github.com/msrevive/nexus2/internal/database"

	"github.com/google/uuid"
	"github.com/dgraph-io/badger/v4"
)

func (d *badgerDB) NewCharacter(steamid string, slot int, size int, data string) (uuid.UUID, error) {
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

	//use a read-write transaction to save CPU time.
	if err := d.db.Update(func(txn *badger.Txn) error {
		userKey := append(UserPrefix, []byte(steamid)...)
		charKey := append(CharPrefix, []byte(charID.String())...)

		item, err := txn.Get(userKey)
		//user doesn't exists so create a new one
		if err != nil {
			user := &schema.User{
				ID: steamid,
				Characters: make(map[int]uuid.UUID),
			}
			user.Characters[slot] = charID

			userData, err := bsoncoder.Encode(&user)
			if err != nil {
				return fmt.Errorf("bson: failed to marshal user %v", err)
			}

			charData, err := bsoncoder.Encode(&char)
			if err != nil {
				return fmt.Errorf("bson: failed to marshal character %v", err)
			}

			//commit new user to DB
			if err := txn.Set(userKey, userData); err != nil {
				return err
			}

			//commit new character to DB
			if err := txn.Set(charKey, charData); err != nil {
				return err
			}

			return nil
		}else{ // user does exists so just create character.
			data, err := item.ValueCopy(nil)
			if err != nil {
				return fmt.Errorf("badger: failed to get value from item")
			}

			var user *schema.User
			if err := bsoncoder.Decode(data, &user); err != nil {
				return fmt.Errorf("bson: failed to unmarshal %v", err)
			}

			user.Characters[slot] = charID

			// we new have to encode the new userdata, this seems like such a waste...
			userData, err := bsoncoder.Encode(&user)
			if err != nil {
				return fmt.Errorf("bson: failed to marshal user %v", err)
			}

			charData, err := bsoncoder.Encode(&char)
			if err != nil {
				return fmt.Errorf("bson: failed to marshal character %v", err)
			}

			//commit new user to DB
			if err := txn.Set(userKey, userData); err != nil {
				return fmt.Errorf("badger: failed to set user %v with key %s", err, userKey)
			}

			//commit new character to DB
			if err := txn.Set(charKey, charData); err != nil {
				return fmt.Errorf("badger: failed to set character %v with key %s", err, charKey)
			}

			return nil
		}

		return nil
	}); err != nil {
		return uuid.Nil, err
	}

	return charID, nil
}

func (d *badgerDB) UpdateCharacter(id uuid.UUID, size int, data string, backupMax int, backupTime time.Duration) error {
	key := append(CharPrefix, []byte(id.String())...)

	if err := d.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		val, err := item.ValueCopy(nil)
		if err != nil {
			return fmt.Errorf("badger: failed to get value from item")
		}

		var char *schema.Character
		if err := bsoncoder.Decode(val, &char); err != nil {
			return fmt.Errorf("bson: failed to unmarshal %v", err)
		}

		//handle character backups for rollback system.
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

		//commit updated character to DB
		charData, err := bsoncoder.Encode(&char)
		if err != nil {
			return fmt.Errorf("bson: failed to encode character %v", err)
		}

		if err := txn.Set(key, charData); err != nil {
			return fmt.Errorf("badger: failed to set character %v with key %s", err, key)
		}

		return nil
	}); err != nil {
		return nil
	}

	return nil
}

func (d *badgerDB) GetCharacter(id uuid.UUID) (char *schema.Character, err error) {
	key := append(CharPrefix, []byte(id.String())...)

	if err = d.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err == badger.ErrKeyNotFound {
			return database.ErrNoDocument
		}else if err != nil {
			return err
		}

		data, err := item.ValueCopy(nil)
		if err != nil {
			return fmt.Errorf("badger: failed to get value from item")
		}

		if err := bsoncoder.Decode(data, &char); err != nil {
			return fmt.Errorf("bson: failed to unmarshal %v", err)
		}

		return nil
	}); err != nil {
		return
	}

	return
}

func (d *badgerDB) GetCharacters(steamid string) (map[int]schema.Character, error) {
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

func (d *badgerDB) LookUpCharacterID(steamid string, slot int) (uuid.UUID, error) {
	user, err := d.GetUser(steamid)
	if err != nil {
		return uuid.Nil, err
	}

	uuid := user.Characters[slot]
	return uuid, nil
}

// We remove the character from user's active list and set an expiration.
func (d *badgerDB) SoftDeleteCharacter(id uuid.UUID, expiration time.Duration) error {
	char, err := d.GetCharacter(id)
	if err != nil {
		return err
	}

	user, err := d.GetUser(char.SteamID)
	if err != nil {
		return err
	}

	delete(user.Characters, char.Slot)
	user.DeletedCharacters = make(map[int]uuid.UUID, 1)
	user.DeletedCharacters[char.Slot] = id

	timeNow := time.Now().UTC()
	char.DeletedAt = &timeNow

	userKey := append(UserPrefix, []byte(char.SteamID)...)
	charKey := append(CharPrefix, []byte(id.String())...)
	if err = d.db.Update(func(txn *badger.Txn) error {
		userData, err := bsoncoder.Encode(&user)
		if err != nil {
			return fmt.Errorf("bson: failed to encode user %v", err)
		}

		charData, err := bsoncoder.Encode(&char)
		if err != nil {
			return fmt.Errorf("bson: failed to encode character %v", err)
		}

		if err := txn.Set(userKey, userData); err != nil {
			return fmt.Errorf("badger: failed to update user %v", err)
		}

		charEntry := badger.NewEntry(charKey, charData)
		charEntry.WithTTL(expiration)
		if err := txn.SetEntry(charEntry); err != nil {
			return fmt.Errorf("badger: failed to update character %v", err)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (d *badgerDB) DeleteCharacter(id uuid.UUID) error {
	key := append(CharPrefix, []byte(id.String())...)
	if err := d.db.Update(func(txn *badger.Txn) error {
		if err := txn.Delete(key); err != nil {
			return fmt.Errorf("badger: failed to delete character %v", err)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (d *badgerDB) DeleteCharacterReference(steamid string, slot int) error {
	user, err := d.GetUser(steamid)
	if err != nil {
		return err
	}

	delete(user.Characters, slot)

	key := append(UserPrefix, []byte(steamid)...)
	if err = d.db.Update(func(txn *badger.Txn) error {
		userData, err := bsoncoder.Encode(&user)
		if err != nil {
			return fmt.Errorf("bson: failed to encode user %v", err)
		}

		if err := txn.Set(key, userData); err != nil {
			return fmt.Errorf("badger: failed to update user %v", err)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (d *badgerDB) MoveCharacter(id uuid.UUID, steamid string, slot int) error {
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

	userKey := append(UserPrefix, []byte(steamid)...)
	charKey := append(CharPrefix, []byte(id.String())...)
	if err = d.db.Update(func(txn *badger.Txn) error {
		userData, err := bsoncoder.Encode(&user)
		if err != nil {
			return fmt.Errorf("bson: failed to encode user %v", err)
		}

		charData, err := bsoncoder.Encode(&char)
		if err != nil {
			return fmt.Errorf("bson: failed to encode character %v", err)
		}

		if err := txn.Set(userKey, userData); err != nil {
			return fmt.Errorf("badger: failed to update user %v", err)
		}

		if err := txn.Set(charKey, charData); err != nil {
			return fmt.Errorf("badger: failed to update character %v", err)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (d *badgerDB) CopyCharacter(id uuid.UUID, steamid string, slot int) (uuid.UUID, error) {
	// Create reference to "new" character.
	targetUser, err := d.GetUser(steamid)
	if err != nil {
		return uuid.Nil, err
	}

	// Insert new character data.
	char, err := d.GetCharacter(id)
	if err != nil {
		return uuid.Nil, err
	}

	charID := uuid.New()
	targetUser.Characters[slot] = charID

	char.ID = charID
	char.SteamID = steamid
	char.Slot = slot
	char.CreatedAt = time.Now().UTC()

	userKey := append(UserPrefix, []byte(steamid)...)
	charKey := append(CharPrefix, []byte(charID.String())...)
	if err = d.db.Update(func(txn *badger.Txn) error {
		userData, err := bsoncoder.Encode(&targetUser)
		if err != nil {
			return fmt.Errorf("bson: failed to encode user %v", err)
		}

		charData, err := bsoncoder.Encode(&char)
		if err != nil {
			return fmt.Errorf("bson: failed to encode character %v", err)
		}

		if err := txn.Set(userKey, userData); err != nil {
			return fmt.Errorf("badger: failed to update user %v", err)
		}

		if err := txn.Set(charKey, charData); err != nil {
			return fmt.Errorf("badger: failed to update character %v", err)
		}

		return nil
	}); err != nil {
		return uuid.Nil, err
	}

	return charID, nil
}

func (d *badgerDB) RestoreCharacter(id uuid.UUID) error {
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

	userKey := append(UserPrefix, []byte(char.SteamID)...)
	charKey := append(CharPrefix, []byte(id.String())...)
	if err = d.db.Update(func(txn *badger.Txn) error {
		userData, err := bsoncoder.Encode(&user)
		if err != nil {
			return fmt.Errorf("bson: failed to encode user %v", err)
		}

		charData, err := bsoncoder.Encode(&char)
		if err != nil {
			return fmt.Errorf("bson: failed to encode character %v", err)
		}

		if err := txn.Set(userKey, userData); err != nil {
			return fmt.Errorf("badger: failed to update user %v", err)
		}

		if err := txn.Set(charKey, charData); err != nil {
			return fmt.Errorf("badger: failed to update character %v", err)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (d *badgerDB) RollbackCharacter(id uuid.UUID, ver int) error {
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

	key := append(CharPrefix, []byte(id.String())...)
	if err = d.db.Update(func(txn *badger.Txn) error {
		charData, err := bsoncoder.Encode(&char)
		if err != nil {
			return fmt.Errorf("bson: failed to encode character %v", err)
		}

		if err := txn.Set(key, charData); err != nil {
			return fmt.Errorf("badger: failed to update char %v", err)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (d *badgerDB) RollbackCharacterToLatest(id uuid.UUID) error {
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

	key := append(CharPrefix, []byte(id.String())...)
	if err = d.db.Update(func(txn *badger.Txn) error {
		charData, err := bsoncoder.Encode(&char)
		if err != nil {
			return fmt.Errorf("bson: failed to encode character %v", err)
		}

		if err := txn.Set(key, charData); err != nil {
			return fmt.Errorf("badger: failed to update char %v", err)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (d *badgerDB) DeleteCharacterVersions(id uuid.UUID) error {
	char, err := d.GetCharacter(id)
	if err != nil {
		return err
	}

	char.Versions = nil

	key := append(CharPrefix, []byte(id.String())...)
	if err = d.db.Update(func(txn *badger.Txn) error {
		charData, err := bsoncoder.Encode(&char)
		if err != nil {
			return fmt.Errorf("bson: failed to encode char %v", err)
		}

		if err := txn.Set(key, charData); err != nil {
			return fmt.Errorf("badger: failed to update char %v", err)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}