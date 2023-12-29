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
	case "clover":
		return fmt.Errorf("database type not yet implemented.")
	default:
		return fmt.Errorf("database not available.")
	}

	return nil
}