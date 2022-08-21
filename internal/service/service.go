package service

import (
	"context"
	"database/sql"
)

type service struct {
	ctx	context.Context
	db *sql.DB
}

func New(ctx context.Context, db *sql.DB) *service {
	return &service{
		ctx: ctx,
		db: db,
	}
}

func (s *service) CreateTables() error {
	stmt := `
	CREATE TABLE IF NOT EXISTS characters
	CREATE TABLE IF NOT EXISTS nerds
	`
	if _, err = s.db.ExecContext(s.ctx, stmt); err != nil {
		return err
	}

	return nil
}

func (s *service) Debug() error {
	// _, err := s.CharacterCreate(ent.DeprecatedCharacter{
	// 	Steamid: "76561198092541763",
	// 	Slot:    1,
	// 	Size:    0,
	// 	Data:    "data",
	// })
	// if err != nil {
	// 	return err
	// }

	return nil
}

// func txn(ctx context.Context, client *ent.Client, fn func(tx *ent.Tx) error) error {
// 	tx, err := client.BeginTx(ctx, nil)
// 	if err != nil {
// 		return err
// 	}

// 	if err := fn(tx); err != nil {
// 		_ = tx.Rollback()
// 		return err
// 	}

// 	if err := tx.Commit(); err != nil {
// 		_ = tx.Rollback()
// 		return err
// 	}

// 	return nil
// }
