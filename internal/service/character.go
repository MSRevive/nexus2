package service

import (
	"fmt"

	"github.com/msrevive/nexus2/internal/payload"
	"github.com/msrevive/nexus2/internal/database/schema"
)

func (s *Service) NewCharacter(char payload.Character) error {
	if err := s.db.NewCharacter(char.SteamID, char.Slot, char.Size, char.Data); err != nil {
		return err
	}

	return nil
}

func (s *Service) GetCharacter(steamid string, slot int) (*schema.CharacterData, error) {
	chars, err := s.db.GetCharacters(steamid); 
	if err != nil {
		return nil, err
	}

	charData, ok := chars[slot].Versions[0]
	if !ok {
		return nil, fmt.Errorf("no character data for version 0 for %s at slot %d", steamid, slot)
	}

	return &charData, nil
}