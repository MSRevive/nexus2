package service

import (
	"github.com/msrevive/nexus2/internal/database"
	"github.com/msrevive/nexus2/internal/config"
)

type Service struct {
	db database.Database
	config config.Config
}

func New(db database.Database, cfg config.Config) *Service {
	return &Service{
		db: db,
		config: cfg,
	}
}