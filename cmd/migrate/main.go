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

	// Import whichever backends you have implemented.
	nexusPebble "github.com/msrevive/nexus2/internal/database/pebble"
	nexusSQLite "github.com/msrevive/nexus2/internal/database/sqlite"
)

func main() {
	srcType := flag.String("src", "", "source backend: pebble | sqlite")
	srcDir  := flag.String("src-dir", "", "source pebble directory")
	srcPath := flag.String("src-path", "", "source sqlite file path")

	dstType := flag.String("dst", "", "destination backend: pebble | sqlite")
	dstDir  := flag.String("dst-dir", "", "destination pebble directory")
	dstPath := flag.String("dst-path", "", "destination sqlite file path")

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

	writer := io.MultiWriter(os.Stdout, file)
	log.SetOutput(writer)

	if *srcType == "" || *dstType == "" {
		log.Fatal("--src and --dst are required")
	}

	src, err := openDB(*srcType, *srcDir, *srcPath)
	if err != nil {
		log.Fatalf("open source: %v", err)
	}
	defer src.Disconnect()

	dst, err := openDB(*dstType, *dstDir, *dstPath)
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

func openDB(kind, dir, path string) (database.Database, error) {
	var db database.Database
	var cfg database.Config

	switch kind {
	case "pebble":
		if dir == "" {
			return nil, fmt.Errorf("pebble backend requires --src-dir or --dst-dir")
		}
		cfg.Pebble.Directory = dir
		db = nexusPebble.New()

	case "sqlite":
		if path == "" {
			return nil, fmt.Errorf("sqlite backend requires --src-path or --dst-path")
		}
		cfg.SQLite.Path = path
		db = nexusSQLite.New()

	default:
		return nil, fmt.Errorf("unknown backend %q (supported: pebble, sqlite)", kind)
	}

	if err := db.Connect(cfg, database.Options{}); err != nil {
		return nil, fmt.Errorf("connect %s: %w", kind, err)
	}
	return db, nil
}
