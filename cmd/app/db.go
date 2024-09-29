package app

import (
	"fmt"
	"time"

	"github.com/msrevive/nexus2/internal/database/mongodb"
	"github.com/msrevive/nexus2/internal/database/bbolt"

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
	default:
		return fmt.Errorf("database not available.")
	}

	return nil
}

// This is for databases that we use a cache to make writing faster.
// Databases that are embedded like BBolt don't use a cache instead data is entered on demand.
func (a *App) SetupDatabaseAutoSave() {
	if a.Config.Core.DBType != "mongodb" {
		return
	}

	cron := cron.New()
	cron.AddFunc("*/30 * * * *", func(){
		a.Logger.Info("Saving characters from database cache...")
		t1 := time.Now()
		if err := a.DB.SaveToDatabase(); err != nil {
			a.Logger.Error("Failed to save characters!", "error", err)
			return
		}
		a.DB.ClearCache()
		a.Logger.Info("Finished saving to database.", "ping", time.Since(t1))
	})
	cron.Start()
}