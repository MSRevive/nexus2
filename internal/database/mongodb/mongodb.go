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
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
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
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err := d.Client.Disconnect(ctx); err != nil {
		return err
	}

	return nil
}

func (d *mongoDB) NewCharacter(steamid string, slot int, size int, data string) (*uuid.UUID, error) {
	filter := bson.D{{"_id", steamid}}
	var user schema.User
	charID := uuid.New()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := d.UserCollection.FindOne(ctx, filter).Decode(&user); err == mongo.ErrNoDocuments {
		user = schema.User{
			ID: steamid,
			Characters: make(map[int]uuid.UUID),
			DeletedCharacters: make(map[int]schema.DeletedCharacter),
		}
		user.Characters[slot] = charID

		if _, err := d.UserCollection.InsertOne(ctx, &user); err != nil {
			return nil, err
		}

		return &charID, nil
	} else if err != nil {
		return nil, err
	}

	user.Characters[slot] = charID
	opts := options.Update().SetUpsert(false)
	update := bson.D{
		{ "$set", bson.D{{ "characters", user.Characters }} },
	}
	_, err := d.UserCollection.UpdateByID(ctx, steamid, update, opts)
	if err != nil {
		return nil, err
	}

	// _,err := d.UserCollection.ReplaceOne(ctx, filter, user)
	// if err != nil {
	// 	return err
	// }

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
	if _, err := d.CharCollection.InsertOne(ctx, &char); err != nil {
		return nil, err
	}

	return &charID, nil
}

func (d *mongoDB) UpdateCharacter(id uuid.UUID, size int, data string, backupMax int, backupTime string) error {
	filter := bson.D{{"_id", id}}
	var char schema.Character

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := d.CharCollection.FindOne(ctx, filter).Decode(&char); err == nil {
		return err
	}

	newChar := schema.CharacterData{
		CreatedAt: time.Now(),
		Size: size,
		Data: data,
	}

	verLen := len(char.Versions)
	if verLen > 0 {
		time, err := time.ParseDuration(backupTime)
		if err != nil {
			return err
		}

		if verLen > backupMax {
			char.Versions = char.Versions[:verLen-1]
		}

		newest := char.Versions[0]
		timeCheck := newest.CreatedAt.Add(time)
		if (newest.CreatedAt.After(timeCheck)) {
			char.Versions = append(char.Versions, newest)
		}

		char.Versions[0] = newChar
	}else{
		char.Versions = append(char.Versions, newChar)
	}

	opts := options.Update().SetUpsert(false)
	update := bson.D{
		{ "$set", bson.D{{ "versions", char.Versions }} },
		{ "$set", bson.D{{ "updated_at", time.Now() }} },
	}
	_, err := d.CharCollection.UpdateByID(ctx, id, update, opts)
	if err != nil {
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
	
	chars := make(map[int]schema.Character, len(user.Characters))
	for k,v := range user.Characters {
		char, err := d.GetCharacter(v)
		if err != nil {
			return nil, err
		}
		chars[k] = *char
	}

	return chars, nil
}

func (d *mongoDB) LookUpCharacterID(steamid string, slot int) (*uuid.UUID, error) {
	user, err := d.GetUser(steamid)
	if err != nil {
		return nil, err
	}

	uuid := user.Characters[slot]
	return &uuid, nil
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
	user.DeletedCharacters[char.Slot] = schema.DeletedCharacter{
		CreatedAt: char.CreatedAt,
		DeletedAt: time.Now(),
		Data: char.Versions[0],
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	opts := options.Update().SetUpsert(false)
	update := bson.D{
		{ "$set", bson.D{{ "characters", user.Characters }} },
		{ "$set", bson.D{{ "deleted_characters", user.DeletedCharacters }} },
	}
	if _, err := d.UserCollection.UpdateByID(ctx, char.SteamID, update, opts); err != nil {
		return err
	}

	filter := bson.D{{"_id", id}}
	if _, err := d.CharCollection.DeleteOne(ctx, filter); err != nil {
		return err
	}

	return nil
}