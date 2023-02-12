package service_test

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"

	"entgo.io/ent/dialect/sql/schema"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/msrevive/nexus2/ent"
	"github.com/msrevive/nexus2/cmd/app"
)

var (
	dbOnce sync.Once
)

func NewDb(apps *app.App) func() {
	teardown := func() {}
	dbOnce.Do(func() {
		fileName := uuid.NewString() + ".db"
		client, _ := ent.Open("sqlite3", "file:"+fileName+"?cache=shared&mode=rwc&_fk=1")
		if err := client.Schema.Create(context.Background(), schema.WithAtlas(true)); err != nil {
			panic(fmt.Errorf("initializing database failed: %w", err))
		}

		apps.Client = client
		teardown = func() {
			client.Close()
			os.Remove(fileName)
		}
	})
	return teardown
}

func refreshDb(apps *app.App) {
	if apps.Client == nil {
		panic("database client expected to be initialized")
	}
	apps.Client.Character.Delete().Exec(context.Background())
	apps.Client.Player.Delete().Exec(context.Background())
}

func TestMain(m *testing.M) {
	// Run Setup
	rand.Seed(time.Now().UnixNano())
	teardown := NewDb()

	// Run tests
	code := m.Run()

	// Run Teardown
	teardown()

	os.Exit(code)
}
