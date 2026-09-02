package service

import (
	"context"

	"github.com/msrevive/nexus2/pkg/database/schema"
	"github.com/msrevive/nexus2/internal/static"

	"github.com/google/uuid"
)

func (s *Service) GetCharacterVersions(ctx context.Context, uid uuid.UUID) (map[int]schema.CharacterData, error) {
	char, err := s.db.GetCharacter(ctx, uid)
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

	return nil, static.ErrNoCharacterVersions
}

func (s *Service) RollbackCharacter(ctx context.Context, uid uuid.UUID, ver int) error {
	if s.readonly {
		return nil
	}

	err := s.db.RollbackCharacter(ctx, uid, ver)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) RollbackCharacterToLatest(ctx context.Context, uid uuid.UUID) error {
	if s.readonly {
		return nil
	}

	err := s.db.RollbackCharacterToLatest(ctx, uid)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) DeleteCharacterVersions(ctx context.Context, uid uuid.UUID) error {
	if s.readonly {
		return nil
	}

	err := s.db.DeleteCharacterVersions(ctx, uid)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) GetCharacterVersionsTimestamp(ctx context.Context, uid uuid.UUID) (data map[int]string, err error) {
	data, err = s.db.GetRollbackVersionsTimestamp(ctx, uid)
	if len(data) == 0 {
		return data, static.ErrNoCharacterVersions
	}

	return
}
