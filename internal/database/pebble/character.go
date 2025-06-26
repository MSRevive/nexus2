package pebble

import (
	"fmt"
	"time"
	//"encoding/binary"
	
	"github.com/msrevive/nexus2/internal/database"
	"github.com/msrevive/nexus2/pkg/database/bsoncoder"
	"github.com/msrevive/nexus2/pkg/database/schema"

	"github.com/google/uuid"
	"github.com/cockroachdb/pebble/v2"
)

/*
	With Pebble we can actually just use the interface functions I.E. GetUser() for the intensive frequently called functions
	because Pebble doesn't use transactions like the others, thus there's no performance penality.
*/

func (d *pebbleDB) NewCharacter(steamid string, slot int, size int, data string) (uuid.UUID, error) {
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

	userKey := append(UserPrefix, []byte(steamid)...)
	charKey := append(CharPrefix, []byte(charID.String())...)

	user, err := d.GetUser(steamid)
	// new user
	if err == database.ErrNoDocument {
		user = &schema.User{
			ID: steamid,
			Characters: make(map[int]uuid.UUID),
		}
		user.Characters[slot] = charID

		userData, err := bsoncoder.Encode(&user)
		if err != nil {
			return uuid.Nil, fmt.Errorf("bson: failed to marshal user %v", err)
		}

		charData, err := bsoncoder.Encode(&char)
		if err != nil {
			return uuid.Nil, fmt.Errorf("bson: failed to marshal character %v", err)
		}

		//commit new user to DB
		if err := d.db.Set(userKey, userData, pebble.NoSync); err != nil {
			return uuid.Nil, err
		}

		//commit new character to DB
		if err := d.db.Set(charKey, charData, pebble.NoSync); err != nil {
			return uuid.Nil, err
		}

	}else{ // user does exists so just create character.
		user.Characters[slot] = charID

		// we new have to encode the new userdata, this seems like such a waste...
		userData, err := bsoncoder.Encode(&user)
		if err != nil {
			return uuid.Nil, fmt.Errorf("bson: failed to marshal user %v", err)
		}

		charData, err := bsoncoder.Encode(&char)
		if err != nil {
			return uuid.Nil, fmt.Errorf("bson: failed to marshal character %v", err)
		}

		//commit new user to DB
		if err := d.db.Set(userKey, userData, pebble.NoSync); err != nil {
			return uuid.Nil, err
		}

		// fmt.Printf("char size: %d\n", len(charData))
		// buf := make([]byte, timestampSize+len(charData)) //8 bytes for a Unix timestamp
		// //binary.BigEndian.PutUint64(buf, uint64(1750813934))
		// //copy(buf[timestampSize:], charData)
		// //buf := make([]byte, timestampSize+len(charData))
		// //binary.BigEndian.PutUint64(buf, uint64(1750813934))
		// copy(buf[timestampSize:], charData)
		// fmt.Printf("char size: %d, timestamp: %v\n", len(buf), binary.BigEndian.Uint64(buf[:8])) //buf[len(buf)-8:] buf[:8]
		// fmt.Printf("buffer: %s\n", buf)

		//commit new character to DB
		if err := d.db.Set(charKey, charData, pebble.NoSync); err != nil {
			return uuid.Nil, err
		}
	}

	return charID, nil
}

func (d *pebbleDB) UpdateCharacter(id uuid.UUID, size int, data string, backupMax int, backupTime time.Duration) error {
	key := append(CharPrefix, []byte(id.String())...)

	val, io, err := d.db.Get(key)
	if err == pebble.ErrNotFound {
		return database.ErrNoDocument
	}else if err != nil {
		return err
	}

	defer io.Close()

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

	return d.db.Set(key, charData, pebble.NoSync)
}

func (d *pebbleDB) GetCharacter(id uuid.UUID) (*schema.Character, error) {
	var char *schema.Character = nil
	key := append(CharPrefix, []byte(id.String())...)

	data, io, err := d.db.Get(key)
	if err == pebble.ErrNotFound {
		return char, database.ErrNoDocument
	}else if err != nil {
		return char, err
	}

	defer io.Close()

	if err := bsoncoder.Decode(data, &char); err != nil {
		return char, fmt.Errorf("bson: failed to unmarshal %v", err)
	}

	return char, nil
}

func (d *pebbleDB) GetCharacters(steamid string) (map[int]schema.Character, error) {
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

func (d *pebbleDB) LookUpCharacterID(steamid string, slot int) (uuid.UUID, error) {
	user, err := d.GetUser(steamid)
	if err != nil {
		return uuid.Nil, err
	}

	uuid := user.Characters[slot]
	return uuid, nil
}

func (d *pebbleDB) SoftDeleteCharacter(id uuid.UUID, expiration time.Duration) error {
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

	userData, err := bsoncoder.Encode(&user)
	if err != nil {
		return fmt.Errorf("bson: failed to encode user %v", err)
	}

	charData, err := bsoncoder.Encode(&char)
	if err != nil {
		return fmt.Errorf("bson: failed to encode character %v", err)
	}

	if err := d.db.Set(userKey, userData, pebble.NoSync); err != nil {
		return fmt.Errorf("pebble: failed to set user %v", err)
	}

	if err := d.setWithTTL(charKey, charData, expiration, pebble.NoSync); err != nil {
		return fmt.Errorf("pebble: failed to set character %v", err)
	}

	return nil
}

func (d *pebbleDB) DeleteCharacter(id uuid.UUID) error {
	key := append(CharPrefix, []byte(id.String())...)
	return d.db.Delete(key, pebble.Sync)
}

func (d *pebbleDB) DeleteCharacterReference(steamid string, slot int) error {
	user, err := d.GetUser(steamid)
	if err != nil {
		return err
	}

	delete(user.Characters, slot)
	key := append(UserPrefix, []byte(steamid)...)

	userData, err := bsoncoder.Encode(&user)
	if err != nil {
		return fmt.Errorf("bson: failed to encode user %v", err)
	}

	if err := d.db.Set(key, userData, pebble.Sync); err != nil {
		return fmt.Errorf("pebble: failed to set user %v", err)
	}

	return nil
}

func (d *pebbleDB) MoveCharacter(id uuid.UUID, steamid string, slot int) error {
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

	userData, err := bsoncoder.Encode(&user)
	if err != nil {
		return fmt.Errorf("bson: failed to encode user %v", err)
	}

	charData, err := bsoncoder.Encode(&char)
	if err != nil {
		return fmt.Errorf("bson: failed to encode character %v", err)
	}

	if err := d.db.Set(userKey, userData, pebble.Sync); err != nil {
		return fmt.Errorf("pebble: failed to set user %v", err)
	}

	if err := d.db.Set(charKey, charData, pebble.Sync); err != nil {
		return fmt.Errorf("pebble: failed to set character %v", err)
	}

	return nil
}

func (d *pebbleDB) CopyCharacter(id uuid.UUID, steamid string, slot int) (uuid.UUID, error) {
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

	userData, err := bsoncoder.Encode(&targetUser)
	if err != nil {
		return uuid.Nil, fmt.Errorf("bson: failed to encode user %v", err)
	}

	charData, err := bsoncoder.Encode(&char)
	if err != nil {
		return uuid.Nil, fmt.Errorf("bson: failed to encode character %v", err)
	}

	if err := d.db.Set(userKey, userData, pebble.Sync); err != nil {
		return uuid.Nil, fmt.Errorf("pebble: failed to set user %v", err)
	}

	if err := d.db.Set(charKey, charData, pebble.Sync); err != nil {
		return uuid.Nil, fmt.Errorf("pebble: failed to set character %v", err)
	}

	return charID, nil
}

func (d *pebbleDB) RestoreCharacter(id uuid.UUID) error {
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
	userData, err := bsoncoder.Encode(&user)
	if err != nil {
		return fmt.Errorf("bson: failed to encode user %v", err)
	}

	charData, err := bsoncoder.Encode(&char)
	if err != nil {
		return fmt.Errorf("bson: failed to encode character %v", err)
	}

	if err := d.db.Set(userKey, userData, pebble.Sync); err != nil {
		return fmt.Errorf("pebble: failed to set user %v", err)
	}

	if err := d.db.Set(charKey, charData, pebble.Sync); err != nil {
		return fmt.Errorf("pebble: failed to set character %v", err)
	}

	return nil
}

func (d *pebbleDB) RollbackCharacter(id uuid.UUID, ver int) error {
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
	charData, err := bsoncoder.Encode(&char)
	if err != nil {
		return fmt.Errorf("bson: failed to encode character %v", err)
	}

	if err := d.db.Set(key, charData, pebble.Sync); err != nil {
		return fmt.Errorf("pebble: failed to set char %v", err)
	}

	return nil
}

func (d *pebbleDB) RollbackCharacterToLatest(id uuid.UUID) error {
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
	charData, err := bsoncoder.Encode(&char)
	if err != nil {
		return fmt.Errorf("bson: failed to encode character %v", err)
	}

	if err := d.db.Set(key, charData, pebble.Sync); err != nil {
		return fmt.Errorf("pebble: failed to set char %v", err)
	}

	return nil
}

func (d *pebbleDB) DeleteCharacterVersions(id uuid.UUID) error {
	char, err := d.GetCharacter(id)
	if err != nil {
		return err
	}

	char.Versions = nil

	key := append(CharPrefix, []byte(id.String())...)
	charData, err := bsoncoder.Encode(&char)
	if err != nil {
		return fmt.Errorf("bson: failed to encode character %v", err)
	}

	if err := d.db.Set(key, charData, pebble.Sync); err != nil {
		return fmt.Errorf("pebble: failed to set char %v", err)
	}

	return nil
}