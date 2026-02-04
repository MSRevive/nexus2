package service

import (
	"github.com/msrevive/nexus2/internal/database"
	"github.com/msrevive/nexus2/internal/config"
)

type Service struct {
	db database.Database
	config *config.Config
	readonly bool
}

func New(db database.Database, cfg *config.Config, ro bool) *Service {
	return &Service{
		db: db,
		config: cfg,
		readonly: ro,
	}
}