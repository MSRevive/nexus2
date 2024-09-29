package mongodb

import (
	"context"
	"time"
	
	"github.com/msrevive/nexus2/internal/bitmask"
	"github.com/msrevive/nexus2/pkg/database/schema"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

func (d *mongoDB) GetUser() ([]*schema.User, error) {
	var users []*schema.User
	return users, nil
}

func (d *mongoDB) GetUser(steamid string) (user *schema.User, err error) {
	filter := bson.D{{"_id", steamid}}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = d.UserCollection.FindOne(ctx, filter).Decode(&user)
	return
}

func (d *mongoDB) SetUserFlags(steamid string, flags bitmask.Bitmask) (error) {
	filter := bson.D{{"_id", steamid}}
	var user schema.User
	uFlags := uint32(flags) // cast it to a uint32 to make database behave.

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := d.UserCollection.FindOne(ctx, filter).Decode(&user); err == mongo.ErrNoDocuments {
		user.Flags = uFlags

		if _, err := d.UserCollection.InsertOne(ctx, &user); err != nil {
			return err
		}

		return nil
	} else if err != nil {
		return err
	}

	return nil
}

func (d *mongoDB) GetUserFlags(steamid string) (bitmask.Bitmask, error) {
	user, err := d.GetUser(steamid)
	if err != nil {
		return 0, err
	}

	return bitmask.Bitmask(user.Flags), nil
}