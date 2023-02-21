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
