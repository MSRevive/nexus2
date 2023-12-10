package service

import (
	"errors"

	"github.com/msrevive/nexus2/internal/database/schema"

	"github.com/google/uuid"
)

func (s *Service) GetCharacterVersions(uid uuid.UUID) ([]schema.CharacterData, error) {
	char, err := s.db.GetCharacter(uid)
	if err != nil {
		return nil, err
	}

	dataLen := len(char.Versions)-1
	if dataLen > 0 {
		return char.Versions[1:], nil
	}
	
	return nil, errors.New("no character versions exist")
}