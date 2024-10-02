package app

import (
	"fmt"
	"time"
	"log"
	"io"
	"os"

	"github.com/msrevive/nexus2/internal/database/mongodb"
	"github.com/msrevive/nexus2/internal/database/bbolt"
	"github.com/msrevive/nexus2/internal/database/badger"

	rw "github.com/saintwish/rotatewriter"
	"github.com/robfig/cron/v3"
)

func (a *App) SetupDatabase() error {
	switch a.Config.Core.DBType {
	case "mongodb":
		a.Logger.Info("Database set to MongoDB!")
		a.DB = mongodb.New()
	case "bbolt":
		a.Logger.Info("Database set to BBolt!")
		a.DB = bbolt.New()
	case "badger":
		a.Logger.Info("Database set to Badger!")
		a.DB = badger.New()
	default:
		return fmt.Errorf("database not available.")
	}

	return nil
}

// This is for databases that we use a cache to make writing faster.
func (a *App) SetupDatabaseAutoSave() {
	syncCron := cron.New()
	syncCron.AddFunc("*/30 * * * *", func(){ //This runs every 30 minutes.
		go func() {
			a.Logger.Info("Syncing data to disk")
			t1 := time.Now()
			if err := a.DB.SyncToDisk(); err != nil {
				a.Logger.Error("Failed to sync data to disk", "error", err)
				return
			}
			a.Logger.Info("Finished syncing data to disk", "ping", time.Since(t1))
		}()
	})
	syncCron.Start()

	gcCron := cron.New()
	gcCron.AddFunc("0 23 * * *", func(){ //This runs at 23:00 every day.
		go func() {
			a.Logger.Info("Running database garbage collection")
			t1 := time.Now()
			if err := a.DB.RunGC(); err != nil {
				a.Logger.Error("Failed to run garbage collection", "error", err)
				return
			}
			a.Logger.Info("Finished running garbage collection", "ping", time.Since(t1))
		}()
	})
	gcCron.Start()
}

// TODO: Move this to database package.
func (a *App) SetUpDatabaseLogger() *log.Logger {
	if a.Config.Core.DBType != "badger" {
		return nil
	}

	iow := io.MultiWriter(os.Stdout, &rw.RotateWriter{
		Dir: a.Config.Log.Dir+"database/",
		Filename: "database.log",
		ExpireTime: a.Config.Log.ExpireTime,
		MaxSize: 5 * rw.Megabyte,
	})

	fmt.Println("Setting up database logger!")
	return log.New(iow, "", log.LstdFlags)
}