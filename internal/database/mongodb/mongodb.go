package mongodb

import (
	"fmt"
	"context"
	"time"
	
	"github.com/msrevive/nexus2/internal/database"
	"github.com/msrevive/nexus2/pkg/database/schema"
	"github.com/msrevive/nexus2/pkg/database/bsoncoder"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson"
	"github.com/google/uuid"
	"github.com/saintwish/kv/kv1s"
)

type mongoDB struct {
	Client *mongo.Client
	UserCollection *mongo.Collection
	CharCollection *mongo.Collection

	CharacterCache *kv1s.Cache[uuid.UUID, schema.Character]
}

func New() *mongoDB {
	return &mongoDB{}
}

func (d *mongoDB) Connect(cfg database.Config, opts database.Options) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dOpts := options.Client().ApplyURI(cfg.MongoDB.Connection).SetRegistry(bsoncoder.GetRegistry())
	client, err := mongo.Connect(ctx, dOpts)
	d.Client = client
	if err != nil {
		return fmt.Errorf("error connecting to database, %w", err)
	}

	if err := client.Ping(context.Background(), nil); err != nil {
		return fmt.Errorf("database connection failed, %w", err)
	}

	d.UserCollection = client.Database("msr").Collection("users")
	d.CharCollection = client.Database("msr").Collection("characters")

	// create cache with 1024 size and max of 2 shards
	d.CharacterCache = kv1s.New[uuid.UUID, schema.Character](1024, 2)

	return nil
}

func (d *mongoDB) Disconnect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := d.SyncToDisk(); err != nil {
		return err
	}

	if err := d.Client.Disconnect(ctx); err != nil {
		return err
	}

	return nil
}

func (d *mongoDB) SyncToDisk() error {
	cacheSize := d.CharacterCache.Count()
	queue := make([]mongo.WriteModel, 0, cacheSize)

	if cacheSize > 0 {
		d.CharacterCache.ForEach(func(id uuid.UUID, char schema.Character){
			updateModel := mongo.NewUpdateOneModel().
			SetUpsert(false).
			SetFilter(bson.D{{"_id", id}}).
			SetUpdate(bson.D{
				{ "$set", bson.D{{ "versions", char.Versions }} },
				{ "$set", bson.D{{ "data", char.Data }} },
			})

			queue = append(queue, updateModel)
		})
	}

	if len(queue) > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _,err := d.CharCollection.BulkWrite(ctx, queue); err != nil {
			return err
		}
	}

	return nil
}

func (d *mongoDB) RunGC() error {
	d.CharacterCache.Clear()
	return nil
}