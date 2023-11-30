package mongodb

import (
	"fmt"
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoDB struct {
	Client *mongo.Client
}

func New() *mongoDB {
	return &mongoDB{}
}

func (d *mongoDB) Connect(conn string) error {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(conn))
	d.Client = client
	if err != nil {
		return fmt.Errorf("error connecting to database, %w", err)
	}

	if err := d.Client.Ping(context.Background(), nil); err != nil {
		return fmt.Errorf("database connection failed, %w", err)
	}

	fmt.Println("Connected to MongoDB!")
	return nil
}

func (d *mongoDB) Disconnect() error {
	if err := d.Client.Disconnect(context.Background()); err != nil {
		return err
	}

	return nil
}