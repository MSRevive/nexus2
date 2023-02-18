package app

import (
	"os"
	"sync"
	"context"
	"errors"
	"fmt"
	"io"
	"time"
	
	"github.com/msrevive/nexus2/ent"
	"github.com/msrevive/nexus2/ent/player"

	"github.com/saintwish/auralog"
	"github.com/goccy/go-json"
	entd "entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/schema"
)

var (
	Version = "canary"

	iPListMutex = new(sync.RWMutex)
	banListMutex = new(sync.RWMutex)
	mapListMutex = new(sync.RWMutex)
	adminListMutex = new(sync.RWMutex)
)

type App struct {
	Config *Config
	Client *ent.Client
	LogCore *auralog.Logger
	LogAPI *auralog.Logger

	IPList map[string]bool
	BanList map[string]bool
	MapList map[string]uint32
	AdminList map[string]bool
}

func New(cfg *Config) *App {
	return &App {
		Config: cfg,
	}
}

func (a *App) newClient(drv *entd.Driver) *ent.Client {
	cfg := a.Config.Database

	db := drv.DB()
	if (cfg.MaxIdleConns > 0) {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}
		
	if (cfg.MaxOpenConns > 0) {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	
	if (cfg.ConnMaxLifetime != "") {
		dur,err := time.ParseDuration(cfg.ConnMaxLifetime)
		if (err != nil) {
			db.SetConnMaxLifetime(dur)
		}
	}

	return ent.NewClient(ent.Driver(drv))
}

func (a *App) SetupClient() error {
	ctx := context.Background()
	dbstring := a.Config.Database.Conn
	dbFileName := "./runtime/chars.db"
	oldDbFileName := "./runtime/old_chars.db"
	dbBakFileName := dbFileName + ".bak"
	dbBakConnStr := "file:" + oldDbFileName + "?cache=shared&mode=rwc&_fk=1"

	// if file doesn't exists then no migration is needed.
	if _, ferr := os.Stat(dbFileName); errors.Is(ferr, os.ErrNotExist) {
		drv, err := entd.Open("sqlite3", dbstring)
		if err != nil {
			return err
		}

		a.Client = a.newClient(drv)
		return nil
	}

	err := func() error {
		/////////////////////////////////////////////////
		/////////// Check if migration is required
		// 1. Check if the 'players' table exists
		sqlClient, err := entd.Open("sqlite3", dbFileName)
		if err != nil {
			return errors.New(fmt.Sprintf("failed to open connection to sqlite3: %v", err))
		}

		rows, err := sqlClient.DB().Query("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='players';")
		if err != nil {
			return errors.New(fmt.Sprintf("failed to get table: %v", err))
		}

		var count int64
		for rows.Next() {
			rows.Scan(&count)
		}
		sqlClient.Close()

		if count > 0 {
			a.LogCore.Debugln("DB already migrated")

			// Set the connection as normal
			drv, err := entd.Open("sqlite3", dbstring)
			if err != nil {
				return errors.New(fmt.Sprintf("failed to open connection to sqlite3: %v", err))
			}

			client := a.newClient(drv)
			if err := client.Schema.Create(ctx, schema.WithAtlas(true)); err != nil {
				return errors.New(fmt.Sprintf("failed to create schema resources: %v", err))
			}
			a.Client = client

			return nil
		}

		a.LogCore.Debugln("DB not migrated")

		/////////////////////////////////////////////////
		/////////// Prepare DB files
		// 1. Copy current DB file (prevents irreversible changes)
		// 2. Rename current DB file (prevents conflicts with new DB)
		currentDbFile, err := os.Open(dbFileName)
		if err != nil {
			return errors.New(fmt.Sprintf("failed to open current DB file: %v", err))
		}

		backupDbFile, err := os.Create(dbBakFileName)
		if err != nil {
			return errors.New(fmt.Sprintf("failed to create backup DB file: %v", err))
		}

		_, err = io.Copy(backupDbFile, currentDbFile)
		if err != nil {
			return errors.New(fmt.Sprintf("failed to backup DB file: %v", err))
		}
		currentDbFile.Close()
		backupDbFile.Close()

		if err := os.Rename(dbFileName, oldDbFileName); err != nil {
			return errors.New(fmt.Sprintf("failed to rename original DB file: %v", err))
		}
		a.LogCore.Debugln("Backed up DB file")

		/////////////////////////////////////////////////
		/////////// Prepare DB schema
		// 1. Rename 'characters' table (allows the use of ORM)
		sqlClient, err = entd.Open("sqlite3", dbBakConnStr)
		if err != nil {
			return errors.New(fmt.Sprintf("failed to open connection to old sqlite3: %v", err))
		}
		a.LogCore.Debugln("Connected to old DB")

		_, err = sqlClient.DB().Exec("ALTER TABLE characters RENAME TO old_characters;")
		if err != nil {
			return errors.New(fmt.Sprintf("failed to rename old table: %v", err))
		}
		sqlClient.Close()
		a.LogCore.Debugln("Renamed table")

		/////////////////////////////////////////////////
		/////////// Collect old data
		// 1. Load all characters into memory using ORM
		client, err := ent.Open("sqlite3", dbBakConnStr)
		if err != nil {
			return errors.New(fmt.Sprintf("failed to open connection to old sqlite3: %v", err))
		}
		a.LogCore.Debugln("Connected to old db")

		oldCharacters, err := client.DeprecatedCharacter.Query().All(ctx)
		if err != nil {
			return errors.New(fmt.Sprintf("failed to get all old characters: %v", err))
		}
		client.Close()
		a.LogCore.Debugln("Loaded all characters")

		/////////////////////////////////////////////////
		/////////// Migrate characters to new DB
		// 1. Connect to new DB (creating file in the process)
		// 2. Loop through old characters
		// 3. Convert old character to new Player + Versioned Character
		// 4. Log any conversion errors
		drv, err := entd.Open("sqlite3", dbstring)
		if err != nil {
			return errors.New(fmt.Sprintf("failed to open connection to sqlite3: %v", err))
		}

		client = a.newClient(drv)
		if err := client.Schema.Create(ctx, schema.WithAtlas(true)); err != nil {
			return errors.New(fmt.Sprintf("failed to create schema resources: %v", err))
		}

		a.Client = client
		a.LogCore.Debugln("connected to new db")

		errLog, err := os.Create("./runtime/logs/migration_errors.log")
		if err != nil {
			return errors.New(fmt.Sprintf("failed to create migration error log: %v", err))
		}

		cLen := len(oldCharacters)
		var failed int
		for i, c := range oldCharacters {
			fmt.Printf("\033[1A\033[Kmigrating %d of %d; steamId='%s'\n", i+1, cLen, c.Steamid)
			player, err := client.Player.Query().Where(player.Steamid(c.Steamid)).Only(ctx)
			if ent.IsNotFound(err) {
				player, err = client.Player.Create().
					SetSteamid(c.Steamid).
					Save(ctx)
				if err != nil {
					return errors.New(fmt.Sprintf("failed to save Player: %v", err))
				}
			}
			_, err = client.Character.Create().
				SetID(c.ID).
				SetPlayer(player).
				SetVersion(1).
				SetSlot(c.Slot).
				SetSize(c.Size).
				SetData(c.Data).
				Save(ctx)
			if err != nil {
				failed++
				errLog.WriteString(fmt.Sprintf("failed to save character! %v\n\t%+v\n", err, c))
			}
		}
		errLog.Close()
		if failed > 0 {
			a.LogCore.Debugf("completed migration with %d errors!", failed)
		}
		a.LogCore.Debugln("Migrated all characters")

		return nil
	}()

	if err != nil {
		// Error detected, revert db changes if possible
		if _, err := os.Stat(dbBakFileName); err == nil {
			os.Remove(dbFileName)
			os.Remove(oldDbFileName)
			os.Rename(dbBakFileName, dbFileName)
		}

		return errors.New(fmt.Sprintf("failed to migrate DB: %v", err))
	}
	
	// happy path cleanup, we don't want to delete old one incase we need to revert suddenly.
	os.Remove(dbBakFileName)
	return nil
}

func (a *App) SetupLoggers(logcore *auralog.Logger, logapi *auralog.Logger) {
	a.LogCore = logcore
	a.LogAPI = logapi
}

func (a *App) SetClient(client *ent.Client) {
	a.Client = client
}

func (a *App) LoadIPList(path string) error {
	file,err := os.ReadFile(path)
	if err != nil {
		return err
	}

	iPListMutex.Lock()
	_ = json.Unmarshal([]byte(file), &a.IPList)
	iPListMutex.Unlock()

	return nil
}

func (a *App) LoadMapList(path string) error {
	file,err := os.ReadFile(path)
	if err != nil {
		return err
	}

	mapListMutex.Lock()
	_ = json.Unmarshal([]byte(file), &a.MapList)
	mapListMutex.Unlock()

	return nil
}

func (a *App) LoadBanList(path string) error {
	file,err := os.ReadFile(path)
	if err != nil {
		return err
	}

	banListMutex.Lock()
	_ = json.Unmarshal([]byte(file), &a.BanList)
	banListMutex.Unlock()

	return nil
}

func (a *App) LoadAdminList(path string) error {
	file,err := os.ReadFile(path)
	if err != nil {
		return err
	}

	adminListMutex.Lock()
	_ = json.Unmarshal([]byte(file), &a.AdminList)
	adminListMutex.Unlock()

	return nil
}