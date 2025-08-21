package app

import (
	"fmt"
	"time"
	"log"
	"io"
	"os"
	"path/filepath"

	"github.com/msrevive/nexus2/internal/database/bbolt"
	"github.com/msrevive/nexus2/internal/database/badger"
	"github.com/msrevive/nexus2/internal/database/pebble"
	"github.com/msrevive/nexus2/pkg/utils"

	rw "github.com/saintwish/rotatewriter"
	"github.com/robfig/cron/v3"
)

func (a *App) SetupDatabase() (err error) {
	switch a.Config.Core.DBType {
	case "bbolt":
		a.Logger.Info("Database set to BBolt!")
		// This is needed because BBolt doesn't automatically create the directory.
		err = os.MkdirAll(filepath.Dir(a.Config.Database.BBolt.File), os.ModePerm)
		a.DB = bbolt.New()
	case "badger":
		a.Logger.Info("Database set to Badger!")
		a.DB = badger.New()
	case "pebble":
		a.Logger.Info("Database set to Pebble!")
		a.DB = pebble.New()
	default:
		err = fmt.Errorf("database not available.")
	}

	// Setup database sync
	syncCron := cron.New()
	syncCron.AddFunc(a.Config.Database.Sync, func() {
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

	// Setup database garbage collection
	gcCron := cron.New()
	gcCron.AddFunc(a.Config.Database.GarbageCollection, func() { //This runs on the 10th minute
		go func() {
			a.Logger.Info("Running database garbage collection")
			t1 := time.Now()
			if err := a.DB.RunGC(); err != nil {
				a.Logger.Warn("Unable to run garbage collection", "error", err)
			}
			a.Logger.Info("Finished running garbage collection", "ping", time.Since(t1))
		}()
	})
	gcCron.Start()

	return err
}

// TODO: Move this to database package.
func (a *App) SetUpDatabaseLogger() *log.Logger {
	if a.Config.Core.DBType != "badger" {
		return nil
	}

	if err := os.MkdirAll(a.Config.Log.Dir+"database/", os.ModePerm); err != nil {
		fmt.Println(fmt.Errorf("database error: failed to create logging directory %v", err))
		return nil
	}

	logExpire, err := utils.ParseDuration(a.Config.Log.ExpireTime)
	if err != nil {
		fmt.Println(fmt.Errorf("database error: failed to parse log duration %s : %v", a.Config.Log.ExpireTime, err))
		return nil
	}

	iow := io.MultiWriter(os.Stdout, &rw.RotateWriter{
		Dir: a.Config.Log.Dir+"database/",
		Filename: "database.log",
		ExpireTime: logExpire,
		MaxSize: 5 * rw.Megabyte,
	})

	fmt.Println("Setting up database logger!")
	return log.New(iow, "", log.LstdFlags)
}