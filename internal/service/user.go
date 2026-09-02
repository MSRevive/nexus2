package service

import (
	"context"

	"github.com/msrevive/nexus2/pkg/database/schema"
	"github.com/msrevive/nexus2/internal/bitmask"

	"github.com/titpetric/oida"
)

func (s *Service) GetAllUsers(ctx context.Context) ([]*schema.User, error) {
	return s.db.GetAllUsers(ctx)
}

func (s *Service) GetUser(ctx context.Context, steamid string) (*schema.User, error) {
	return s.db.GetUser(ctx, steamid)
}

func (s *Service) GetUserFlags(ctx context.Context, steamid string) (bitmask.Bitmask, error) {
	flags, err := s.db.GetUserFlags(ctx, steamid)
	if err != nil {
		return 0, err
	}

	return flags, nil
}

func (s *Service) AddUserFlag(ctx context.Context, steamid string, flag bitmask.Bitmask) (error) {
	if s.readonly {
		return nil
	}

	ctx, span := oida.Start(ctx, "svc AddUserFlag")
	defer span.End()
	span.SetAttribute("steamid", steamid)

	flags, err := s.db.GetUserFlags(ctx, steamid)
	if err != nil {
		span.RecordError(err)
		return err
	}

	flags.AddFlag(flag)

	if err := s.db.SetUserFlags(ctx, steamid, flags); err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}

func (s *Service) RemoveUserFlag(ctx context.Context, steamid string, flag bitmask.Bitmask) (error) {
	if s.readonly {
		return nil
	}

	ctx, span := oida.Start(ctx, "svc RemoveUserFlag")
	defer span.End()
	span.SetAttribute("steamid", steamid)

	flags, err := s.db.GetUserFlags(ctx, steamid)
	if err != nil {
		span.RecordError(err)
		return err
	}

	flags.ClearFlag(flag)

	if err := s.db.SetUserFlags(ctx, steamid, flags); err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}
