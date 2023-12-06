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
)

type mongoDB struct {
	Client *mongo.Client
	UserCollection *mongo.Collection
}

func New() *mongoDB {
	return &mongoDB{}
}

func (d *mongoDB) Connect(conn string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(conn))
	d.Client = client
	if err != nil {
		return fmt.Errorf("error connecting to database, %w", err)
	}

	if err := client.Ping(context.Background(), nil); err != nil {
		return fmt.Errorf("database connection failed, %w", err)
	}

	d.UserCollection = client.Database("msr").Collection("users")

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

func (d *mongoDB) NewCharacter(steamid string, slot int, size int, data string) error {
	filter := bson.D{{"_id", steamid}}
	var user schema.User
	char := schema.Character{
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
			Characters: make(map[int]schema.Character),
		}
		user.Characters[slot] = char

		if _, err := d.UserCollection.InsertOne(ctx, &user); err != nil {
			return err
		}

		return nil
	} else if err != nil {
		return err
	}

	user.Characters[slot] = char
	opts := options.Update().SetUpsert(false)
	update := bson.D{
		{ "$set", bson.D{{ "characters", user.Characters }} },
	}
	_, err := d.UserCollection.UpdateByID(ctx, steamid, update, opts)
	if err != nil {
		return err
	}

	// _,err := d.UserCollection.ReplaceOne(ctx, filter, user)
	// if err != nil {
	// 	return err
	// }
	return nil
}

func (d *mongoDB) UpdateCharacter(steamid string, slot int, size int, data string, backupMax int, backupTime string) error {
	filter := bson.D{{"_id", steamid}}
	var user schema.User

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := d.UserCollection.FindOne(ctx, filter).Decode(&user); err == nil {
		return err
	}

	newChar := schema.CharacterData{
		CreatedAt: time.Now(),
		Size: size,
		Data: data,
	}

	updatedChar := user.Characters[slot]
	updatedChar.UpdatedAt = time.Now()
	verLen := len(updatedChar.Versions)
	if verLen > 0 {
		time, err := time.ParseDuration(backupTime)
		if err != nil {
			return err
		}

		if verLen > backupMax {
			updatedChar.Versions = updatedChar.Versions[:verLen-1]
		}

		newest := updatedChar.Versions[0]
		timeCheck := newest.CreatedAt.Add(time)
		if (newest.CreatedAt.After(timeCheck)) {
			updatedChar.Versions = append(updatedChar.Versions, newest)
		}

		updatedChar.Versions[0] = newChar
	}else{
		updatedChar.Versions = append(updatedChar.Versions, newChar)
	}

	user.Characters[slot] = updatedChar
	opts := options.Update().SetUpsert(false)
	update := bson.D{
		{ "$set", bson.D{{ "characters", user.Characters }} },
	}
	_, err := d.UserCollection.UpdateByID(ctx, steamid, update, opts)
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

func (d *mongoDB) GetCharacters(steamid string) (map[int]schema.Character, error) {
	user, err := d.GetUser(steamid)
	if err != nil {
		return nil, err
	}
	return user.Characters, nil
}