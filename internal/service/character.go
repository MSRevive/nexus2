package service

import (
	"errors"

	"github.com/msrevive/nexus2/internal/payload"
	"github.com/msrevive/nexus2/internal/database/schema"

	"github.com/google/uuid"
)

func (s *Service) NewCharacter(char payload.Character) (*uuid.UUID, error) {
	uuid, err := s.db.NewCharacter(char.SteamID, char.Slot, char.Size, char.Data); 
	if err != nil {
		return nil, err
	}

	return uuid, nil
}

func (s *Service) UpdateCharacter(uuid uuid.UUID, char payload.Character) error {
	if err := s.db.UpdateCharacter(uuid, char.Size, char.Data, s.config.Char.MaxBackups, s.config.Char.BackupTime); err != nil {
		return err
	}

	return nil
}

func (s *Service) GetCharacterByID(uuid uuid.UUID) (*schema.CharacterData, error) {
	char, err := s.db.GetCharacter(uuid); 
	if err != nil {
		return nil, err
	}

	charData := char.Versions[0]
	if len(char.Versions) == 0 {
		return nil, errors.New("missing character data for 0")
	}

	return &charData, nil
}

func (s *Service) GetCharacter(steamid string, slot int) (*schema.CharacterData, error) {
	uid, err := s.db.LookUpCharacterID(steamid, slot)
	if err != nil {
		return nil, err
	}

	char, err := s.db.GetCharacter(*uid); 
	if err != nil {
		return nil, err
	}

	charData := char.Versions[0]
	if len(char.Versions) == 0 {
		return nil, errors.New("missing character data for 0")
	}

	return &charData, nil
}

func (s *Service) DeleteCharacter(uuid uuid.UUID) error {
	if err := s.db.SoftDeleteCharacter(uuid); err != nil {
		return err
	}

	return nil
}