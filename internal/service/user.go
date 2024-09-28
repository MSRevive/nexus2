package service

import (
	"github.com/msrevive/nexus2/internal/bitmask"
)

func (s *Service) GetUserFlags(steamid string) (bitmask.Bitmask, error) {
	flags, err := s.db.GetUserFlags(steamid)
	if err != nil {
		return 0, err
	}

	return bitmask.Bitmask(flags), nil
}

func (s *Service) AddUserFlag(steamid string, flag bitmask.Bitmask) (error) {
	rawFlags, err := s.db.GetUserFlags(steamid)
	if err != nil {
		err
	}

	flags := bitmask.Bitmask(rawFlags)
	flags.AddFlag(flag)

	if err := s.db.SetUserFlags(steamid, flags); err != nil {
		return err
	}

	return nil
}

func (s *Service) RemoveUserFlag(steamid string, flag bitmask.Bitmask) (error) {
	rawFlags, err := s.db.GetUserFlags(steamid)
	if err != nil {
		err
	}

	flags := bitmask.Bitmask(rawFlags)
	flags.ClearFlag(flag)

	if err := s.db.SetUserFlags(steamid, flags); err != nil {
		return err
	}

	return nil
}