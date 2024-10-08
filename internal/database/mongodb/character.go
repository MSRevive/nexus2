package mongodb

import (
	"fmt"
	"context"
	"time"

	"github.com/msrevive/nexus2/pkg/database/schema"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"github.com/google/uuid"
)

func (d *mongoDB) NewCharacter(steamid string, slot int, size int, data string) (uuid.UUID, error) {
	filter := bson.D{{"_id", steamid}}
	var user schema.User
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := d.UserCollection.FindOne(ctx, filter).Decode(&user); err == mongo.ErrNoDocuments {
		user = schema.User{
			ID: steamid,
			Characters: make(map[int]uuid.UUID),
		}
		user.Characters[slot] = charID

		if _, err := d.UserCollection.InsertOne(ctx, &user); err != nil {
			return uuid.Nil, err
		}

		if _, err := d.CharCollection.InsertOne(ctx, &char); err != nil {
			return uuid.Nil, err
		}

		return charID, nil
	} else if err != nil {
		return uuid.Nil, err
	}

	user.Characters[slot] = charID
	update := bson.D{
		{ "$set", bson.D{{ "characters", user.Characters }} },
	}
	_, err := d.UserCollection.UpdateByID(ctx, steamid, update)
	if err != nil {
		return uuid.Nil, err
	}

	if _, err := d.CharCollection.InsertOne(ctx, &char); err != nil {
		return uuid.Nil, err
	}

	return charID, nil
}

func (d *mongoDB) UpdateCharacter(id uuid.UUID, size int, data string, backupMax int, backupTime time.Duration) error {
	var char schema.Character

	if d.CharacterCache.Has(id) {
		char = d.CharacterCache.Get(id)
	}else{
		filter := bson.D{{"_id", id}}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := d.CharCollection.FindOne(ctx, filter).Decode(&char); err != nil {
			return err
		}

		d.CharacterCache.SetOrUpdate(id, char)
	}

	bCharsLen := len(char.Versions)
	if backupMax > 0 {
		// remove oldest backup which is the first element.
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
	if err := d.CharacterCache.Update(id, char); err != nil {
		return err
	}

	return nil
}

func (d *mongoDB) GetCharacter(id uuid.UUID) (char *schema.Character, err error) {
	filter := bson.D{{"_id", id}}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = d.CharCollection.FindOne(ctx, filter).Decode(&char)
	return
}

func (d *mongoDB) GetCharacters(steamid string) (map[int]schema.Character, error) {
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

func (d *mongoDB) LookUpCharacterID(steamid string, slot int) (uuid.UUID, error) {
	user, err := d.GetUser(steamid)
	if err != nil {
		return uuid.Nil, err
	}

	uuid := user.Characters[slot]
	return uuid, nil
}

func (d *mongoDB) SoftDeleteCharacter(id uuid.UUID, expiration time.Duration) error {
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	update := bson.D{
		{ "$set", bson.D{{ "characters", user.Characters }} },
		{ "$set", bson.D{{ "deleted_characters", user.DeletedCharacters }} },
	}
	if _, err := d.UserCollection.UpdateByID(ctx, char.SteamID, update); err != nil {
		return err
	}

	update = bson.D{
		{ "$set", bson.D{{ "deleted_at", time.Now().UTC() }} },
	}
	if _, err := d.CharCollection.UpdateByID(ctx, id, update); err != nil {
		return err
	}

	return nil
}

func (d *mongoDB) DeleteCharacter(id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := d.CharCollection.DeleteOne(ctx, bson.D{{"_id", id}}); err != nil {
		return err
	}

	return nil
}

func (d *mongoDB) DeleteCharacterReference(steamid string, slot int) error {
	user, err := d.GetUser(steamid)
	if err != nil {
		return err
	}

	delete(user.Characters, slot)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	update := bson.D{
		{ "$set", bson.D{{ "characters", user.Characters }} },
	}
	if _, err := d.UserCollection.UpdateByID(ctx, steamid, update); err != nil {
		return err
	}

	return nil
}

func (d *mongoDB) MoveCharacter(id uuid.UUID, steamid string, slot int) error {
	user, err := d.GetUser(steamid)
	if err != nil {
		return err
	}

	user.Characters[slot] = id

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.D{
		{ "$set", bson.D{{ "characters", user.Characters }} },
	}
	if _, err := d.UserCollection.UpdateByID(ctx, steamid, update); err != nil {
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

	// Update character with new user information.
	update = bson.D{
		{ "$set", bson.D{{ "steamid", steamid }} },
		{ "$set", bson.D{{ "slot", slot }} },
	}
	if _, err := d.CharCollection.UpdateByID(ctx, id, update); err != nil {
		return err
	}

	return nil
}

func (d *mongoDB) CopyCharacter(id uuid.UUID, steamid string, slot int) (uuid.UUID, error) {
	charID := uuid.New()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create reference to "new" character.
	user, err := d.GetUser(steamid)
	if err != nil {
		return uuid.Nil, err
	}

	user.Characters[slot] = charID

	update := bson.D{
		{ "$set", bson.D{{ "characters", user.Characters }} },
	}
	if _, err := d.UserCollection.UpdateByID(ctx, steamid, update); err != nil {
		return uuid.Nil, err
	}

	// Insert new character data.
	char, err := d.GetCharacter(id)
	if err != nil {
		return uuid.Nil, err
	}
	char.ID = charID
	char.SteamID = steamid
	char.Slot = slot
	char.CreatedAt = time.Now().UTC()

	if _, err := d.CharCollection.InsertOne(ctx, &char); err != nil {
		return uuid.Nil, err
	}

	return charID, nil
}

func (d *mongoDB) RestoreCharacter(id uuid.UUID) error {
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	update := bson.D{
		{ "$set", bson.D{{ "characters", user.Characters }} },
		{ "$set", bson.D{{ "deleted_characters", user.DeletedCharacters }} },
	}
	if _, err := d.UserCollection.UpdateByID(ctx, char.SteamID, update); err != nil {
		return err
	}

	update = bson.D{
		{ "$unset", bson.D{{ "deleted_at", nil }} },
	}
	if _, err := d.CharCollection.UpdateByID(ctx, id, update); err != nil {
		return err
	}

	return nil
}

func (d *mongoDB) RollbackCharacter(id uuid.UUID, ver int) error {
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	update := bson.D{
		{ "$set", bson.D{{ "data", char.Data }} },
	}
	if _, err := d.CharCollection.UpdateByID(ctx, id, update); err != nil {
		return err
	}

	return nil
}

func (d *mongoDB) RollbackCharacterToLatest(id uuid.UUID) error {
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	update := bson.D{
		{ "$set", bson.D{{ "data", char.Data }} },
	}
	if _, err := d.CharCollection.UpdateByID(ctx, id, update); err != nil {
		return err
	}

	return nil
}

func (d *mongoDB) DeleteCharacterVersions(id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	update := bson.D{
		{ "$unset", bson.D{{ "versions", nil }} },
	}
	if _, err := d.CharCollection.UpdateByID(ctx, id, update); err != nil {
		return err
	}

	return nil
}