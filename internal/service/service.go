package service

import (
	"context"

	"github.com/msrevive/nexus2/cmd/app"
	"github.com/msrevive/nexus2/ent"
)

type service struct {
	ctx 	context.Context
	apps 	*app.App
	client 	*ent.Client
}

func New(ctx context.Context, apps *app.App) *service {
	return &service{
		ctx:    ctx,
		apps: 	apps,
		client: apps.Client,
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
