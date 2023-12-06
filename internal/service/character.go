package service

import (
	"errors"

	"github.com/msrevive/nexus2/internal/payload"
	"github.com/msrevive/nexus2/internal/database/schema"
)

func (s *Service) NewCharacter(char payload.Character) error {
	if err := s.db.NewCharacter(char.SteamID, char.Slot, char.Size, char.Data); err != nil {
		return err
	}

	return nil
}

func (s *Service) UpdateCharacter(char payload.Character) error {
	if err := s.db.UpdateCharacter(char.SteamID, char.Slot, char.Size, char.Data, s.config.Char.MaxBackups ,s.config.Char.BackupTime); err != nil {
		return err
	}

	return nil
}

func (s *Service) GetCharacter(steamid string, slot int) (*schema.CharacterData, error) {
	chars, err := s.db.GetCharacters(steamid); 
	if err != nil {
		return nil, err
	}

	char, ok := chars[slot]
	if !ok {
		return nil, errors.New("character doesn't exists")
	}

	charData := char.Versions[0]
	if len(char.Versions) == 0 {
		return nil, errors.New("missing character data for 0")
	}

	return &charData, nil
}