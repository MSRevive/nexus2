package service

import (
	"fmt"

	"github.com/msrevive/nexus2/pkg/database/schema"

	"github.com/bwmarrin/snowflake"
)

func (s *Service) GetCharacterVersions(uid snowflake.ID) (map[int]schema.CharacterData, error) {
	char, err := s.db.GetCharacter(uid)
	if err != nil {
		return nil, err
	}

	backupLen := len(char.Versions)
	if backupLen > 0 {
		datas := make(map[int]schema.CharacterData, backupLen)

		for k,v := range char.Versions {
			datas[k] = v
		}

		return datas, nil
	}
	
	return nil, fmt.Errorf("no character versions exist")
}

func (s *Service) RollbackCharacter(uid snowflake.ID, ver int) error {
	err := s.db.RollbackCharacter(uid, ver)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) RollbackCharacterToLatest(uid snowflake.ID) error {
	err := s.db.RollbackCharacterToLatest(uid)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteCharacterVersions(uid snowflake.ID) error {
	err := s.db.DeleteCharacterVersions(uid)
	if err != nil {
		return err
	}

	return nil
}