package service

import (
	"github.com/msrevive/nexus2/pkg/database/schema"
	"github.com/msrevive/nexus2/internal/bitmask"
)

func (s *Service) GetAllUsers() ([]*schema.User, error) {
	return s.db.GetAllUsers()
}

func (s *Service) GetUser(steamid string) (*schema.User, error) {
	return s.db.GetUser(steamid)
}

func (s *Service) GetUserFlags(steamid string) (bitmask.Bitmask, error) {
	flags, err := s.db.GetUserFlags(steamid)
	if err != nil {
		return 0, err
	}

	return flags, nil
}

func (s *Service) AddUserFlag(steamid string, flag bitmask.Bitmask) (error) {
	flags, err := s.db.GetUserFlags(steamid)
	if err != nil {
		return err
	}

	flags.AddFlag(flag)

	if err := s.db.SetUserFlags(steamid, flags); err != nil {
		return err
	}

	return nil
}

func (s *Service) RemoveUserFlag(steamid string, flag bitmask.Bitmask) (error) {
	flags, err := s.db.GetUserFlags(steamid)
	if err != nil {
		return err
	}

	flags.ClearFlag(flag)

	if err := s.db.SetUserFlags(steamid, flags); err != nil {
		return err
	}

	return nil
}