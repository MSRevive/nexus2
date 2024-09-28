package service

import (
	"github.com/msrevive/nexus2/internal/database"
	"github.com/msrevive/nexus2/internal/config"
	"github.com/msrevive/nexus2/internal/bitmask"
)

type Service struct {
	db database.Database
	config *config.Config
}

func New(db database.Database, cfg *config.Config) *Service {
	return &Service{
		db: db,
		config: cfg,
	}
}

func (s *Service) GetUserFlags(steamid string) (bitmask.Bitmask, error) {
	flags, err := s.db.GetUserFlags(steamid)
	if err != nil {
		return 0, err
	}

	return bitmask.Bitmask(flags), nil
}