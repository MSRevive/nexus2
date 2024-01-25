package app

import (
	"fmt"

	"github.com/msrevive/nexus2/internal/database/mongodb"
)

func (a *App) SetupDatabase() error {
	switch a.Config.Core.DBType {
	case "mongodb":
		a.Logger.Info("Database set to MongoDB!")
		a.DB = mongodb.New()
	default:
		return fmt.Errorf("database not available.")
	}

	return nil
}