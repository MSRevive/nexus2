package mongodb

import (
	"fmt"
	"context"
	"time"
	//"strconv"
	
	"github.com/msrevive/nexus2/internal/database/schema"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson"
	"github.com/google/uuid"
)

type mongoDB struct {
	Client *mongo.Client
	UserCollection *mongo.Collection
	CharCollection *mongo.Collection
}

func New() *mongoDB {
	return &mongoDB{}
}

func (d *mongoDB) Connect(conn string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Client().ApplyURI(conn).SetRegistry(mongoRegistry)
	client, err := mongo.Connect(ctx, opts)
	d.Client = client
	if err != nil {
		return fmt.Errorf("error connecting to database, %w", err)
	}

	if err := client.Ping(context.Background(), nil); err != nil {
		return fmt.Errorf("database connection failed, %w", err)
	}

	d.UserCollection = client.Database("msr").Collection("users")
	d.CharCollection = client.Database("msr").Collection("characters")

	fmt.Println("Connected to MongoDB!")
	return nil
}

func (d *mongoDB) Disconnect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := d.Client.Disconnect(ctx); err != nil {
		return err
	}

	return nil
}

func (d *mongoDB) NewCharacter(steamid string, slot int, size int, data string) (uuid.UUID, error) {
	filter := bson.D{{"_id", steamid}}
	var user schema.User
	charID := uuid.New()
	char := schema.Character{
		ID: charID,
		SteamID: steamid,
		Slot: slot,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Versions: []schema.CharacterData{
			schema.CharacterData{
				CreatedAt: time.Now(),
				Size: size,
				Data: data,
			},
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

func (d *mongoDB) UpdateCharacter(id uuid.UUID, size int, data string, backupMax int, backupTime string) error {
	filter := bson.D{{"_id", id}}
	var char schema.Character

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := d.CharCollection.FindOne(ctx, filter).Decode(&char); err != nil {
		return err
	}

	newChar := schema.CharacterData{
		CreatedAt: time.Now(),
		Size: size,
		Data: data,
	}

	// Version 0 is always going to be the current character so it's reserved.
	// Everything after Version 0 are the backups.
	bChars := char.Versions[1:]
	bCharsLen := len(bChars)
	if backupMax > 0 {
		// Remove last backup if over the limit
		if bCharsLen >= backupMax {
			char.Versions = char.Versions[:len(char.Versions)-1]
		}

		time, err := time.ParseDuration(backupTime)
		if err != nil {
			return err
		}

		curChar := char.Versions[0]
		if bCharsLen > 0 {
			bNewest := bChars[0] //latest backup

			timeCheck := bNewest.CreatedAt.Add(time)
			if curChar.CreatedAt.After(timeCheck) {
				char.Versions = append(char.Versions, curChar)
			}
		}else{
			char.Versions = append(char.Versions, curChar)
		}
	}

	// Update the current character.
	char.Versions[0] = newChar

	update := bson.D{
		{ "$set", bson.D{{ "versions", char.Versions }} },
		{ "$set", bson.D{{ "updated_at", newChar.CreatedAt }} },
	}
	if _, err := d.CharCollection.UpdateByID(ctx, id, update); err != nil {
		return err
	}

	return nil
}

func (d *mongoDB) GetUser(steamid string) (user *schema.User, err error) {
	filter := bson.D{{"_id", steamid}}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = d.UserCollection.FindOne(ctx, filter).Decode(&user)
	return
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

func (d *mongoDB) SoftDeleteCharacter(id uuid.UUID) error {
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
		{ "$set", bson.D{{ "deleted_at", time.Now() }} },
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
	char.CreatedAt = time.Now()
	char.UpdatedAt = time.Now()

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