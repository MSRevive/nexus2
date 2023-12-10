package service

import (
	"errors"

	"github.com/msrevive/nexus2/internal/database/schema"

	"github.com/google/uuid"
)

func (s *Service) GetCharacterVersions(uid uuid.UUID) (map[int]schema.CharacterData, error) {
	char, err := s.db.GetCharacter(uid)
	if err != nil {
		return nil, err
	}

	dataLen := len(char.Versions)-1
	if dataLen > 0 {
		datas := make(map[int]schema.CharacterData, dataLen)

		for k,v := range char.Versions[1:] {
			datas[k] = v
		}

		return datas, nil
	}
	
	return nil, errors.New("no character versions exist")
}