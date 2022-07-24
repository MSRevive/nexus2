package service

import (
	"context"

	"github.com/msrevive/nexus2/internal/ent"
	"github.com/msrevive/nexus2/internal/system"
)

type service struct {
	ctx    context.Context
	client *ent.Client
}

func New(ctx context.Context) *service {
	return &service{
		ctx:    ctx,
		client: system.Client,
	}
}

func (s *service) Debug() error {
	_, err := s.CharacterCreate(ent.DeprecatedCharacter{
		Steamid: "76561198092541763",
		Slot:    1,
		Size:    0,
		Data:    "data",
	})
	if err != nil {
		return err
	}

	return nil
}

func txn(ctx context.Context, client *ent.Client, fn func(tx *ent.Tx) error) error {
	tx, err := client.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return err
	}

	return nil
}
