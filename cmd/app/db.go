package app

import (
	"fmt"
	"time"
	"log"
	"io"
	"os"

	"github.com/msrevive/nexus2/internal/database"
	"github.com/msrevive/nexus2/internal/database/sqlite"
	"github.com/msrevive/nexus2/internal/database/postgres"
	"github.com/msrevive/nexus2/pkg/utils"

	rw "github.com/saintwish/rotatewriter"
	"github.com/robfig/cron/v3"
)

func (a *App) SetupDatabase() error {
	switch a.Config.Core.DBType {
	case "sqlite":
		a.Logger.Info("Database set to SQLite!")
		a.DB = sqlite.New()
	case "postgres":
		a.Logger.Info("Database set to PostgreSQL!")
		a.DB = postgres.New()
	default:
		return database.ErrNotAvailable
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

	return nil
}

func (a *App) DatabaseConnect() error {
	maxRetries := a.Config.Database.Postgres.MaxRetries
	baseDelay := a.Config.Database.Postgres.RetryDelay

	if maxRetries > 0 {
		for attempt := 1; attempt <= maxRetries; attempt++ {
			err := a.DB.Connect(a.Config.Database, database.Options{
				Logger: a.SetUpDatabaseLogger(),
			})
			if err == nil {
				return nil
			}

			if attempt == maxRetries {
				return fmt.Errorf("database: failed after %d attempts: %w", maxRetries, err)
			}

			delay := baseDelay * time.Duration(1<<(attempt-1)) // exponential: 2s, 4s, 8s, 16s
			if delay > 60*time.Second {
				delay = 60 * time.Second
			}

			a.Logger.Warn("database connection failed, retrying",
				"attempt", attempt,
				"next_retry_in", delay,
				"error", err,
			)
			time.Sleep(delay)
		}
	}else{
		err := a.DB.Connect(a.Config.Database, database.Options{
			Logger: a.SetUpDatabaseLogger(),
		})
		if err == nil {
			return nil
		}
		attempt := 1

		delay := baseDelay * time.Duration(1<<(attempt-1)) // exponential: 2s, 4s, 8s, 16s
		if delay > 60*time.Second {
			delay = 60 * time.Second
		}

		a.Logger.Warn("database connection failed, retrying",
			"attempt", attempt,
			"next_retry_in", delay,
			"error", err,
		)
		time.Sleep(delay)
		attempt++
	}

	return nil // unreachable
}

// TODO: Move this to database package.
func (a *App) SetUpDatabaseLogger() *log.Logger {
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

	fmt.Println("\t Setting up database logger...")
	return log.New(iow, "", log.LstdFlags)
}