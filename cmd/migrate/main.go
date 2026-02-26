// cmd/migrate/main.go
//
// Usage:
//   go run ./cmd/migrate --src pebble --src-dir ./data/pebble \
//                        --dst sqlite --dst-path ./data/nexus.db
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"io"
	"errors"
	"time"

	"github.com/msrevive/nexus2/internal/database"
	"github.com/msrevive/nexus2/internal/migration"
	"github.com/msrevive/nexus2/internal/config"

	// Import whichever backends you have implemented.
	nexusPebble "github.com/msrevive/nexus2/internal/database/pebble"
	nexusSQLite "github.com/msrevive/nexus2/internal/database/sqlite"
	nexusPostgres "github.com/msrevive/nexus2/internal/database/postgres"
)

func main() {
	srcType := flag.String("src", "", "source backend: pebble | sqlite | postgres")
	dstType := flag.String("dst", "", "destination backend: pebble | sqlite | postgres")

	flag.Parse()

	// create logger
	if _, err := os.Stat("./runtime/migration.log"); !errors.Is(err, os.ErrNotExist) {
		os.Remove("./runtime/migration.log")
	}
	file, err := os.OpenFile("./runtime/migration.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	cfg, err := config.Load("./runtime/config.yaml")
	if err != nil {
		panic(err)
	}

	writer := io.MultiWriter(os.Stdout, file)
	log.SetOutput(writer)

	if *srcType == "" || *dstType == "" {
		log.Fatal("--src and --dst are required")
	}

	src, err := openDB(*srcType, cfg.Database)
	if err != nil {
		log.Fatalf("open source: %v", err)
	}
	defer src.Disconnect()

	dst, err := openDB(*dstType, cfg.Database)
	if err != nil {
		log.Fatalf("open destination: %v", err)
	}
	defer dst.Disconnect()

	// actually start migration now
	fmt.Printf("Beginning migration of DB to %s...\n", dstType)
	start := time.Now()

	m := migration.New(src, dst)
	m.OnProgress = func(steamID string, slot int, charID string) {
		log.Printf("  migrated slot %d / char %s for user %s", slot, charID, steamID)
	}

	if err := m.Run(); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	fmt.Printf("Migration finished, took %v\n", time.Since(start))
	os.Exit(0)
}

func openDB(kind string, dbcfg database.Config) (database.Database, error) {
	var db database.Database

	switch kind {
	case "pebble":
		db = nexusPebble.New()

	case "sqlite":
		db = nexusSQLite.New()

	case "postgres":
		db = nexusPostgres.New()

	default:
		return nil, fmt.Errorf("unknown backend %q (supported: pebble, sqlite)", kind)
	}

	if err := db.Connect(dbcfg, database.Options{}); err != nil {
		return nil, fmt.Errorf("connect %s: %w", kind, err)
	}
	return db, nil
}
