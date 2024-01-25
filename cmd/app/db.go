package app

import (
	"fmt"

	"github.com/msrevive/nexus2/internal/database/mongodb"
	"github.com/msrevive/nexus2/internal/database/bbolt"
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