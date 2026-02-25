package postgres

import (
	"context"
	"fmt"

	"github.com/msrevive/nexus2/internal/bitmask"
	"github.com/msrevive/nexus2/internal/database"
	"github.com/msrevive/nexus2/pkg/database/schema"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (d *postgresDB) GetAllUsers() ([]*schema.User, error) {
	ctx := context.Background()
	ctx2 := context.Background()
	rows, err := d.db.Query(ctx, `SELECT id, revision, flags FROM users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*schema.User
	for rows.Next() {
		u := &schema.User{
			Characters:        make(map[int]uuid.UUID),
			DeletedCharacters: make(map[int]uuid.UUID),
		}
		if err := rows.Scan(&u.ID, &u.Revision, &u.Flags); err != nil {
			return nil, err
		}
		if err := d.loadUserCharacters(ctx2, u); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (d *postgresDB) GetUser(steamid string) (*schema.User, error) {
	ctx := context.Background()
	u := &schema.User{
		Characters:        make(map[int]uuid.UUID),
		DeletedCharacters: make(map[int]uuid.UUID),
	}

	err := d.db.QueryRow(ctx,
		`SELECT id, revision, flags FROM users WHERE id = $1`, steamid,
	).Scan(&u.ID, &u.Revision, &u.Flags)

	if err == pgx.ErrNoRows {
		return nil, database.ErrNoDocument
	}
	if err != nil {
		return nil, err
	}

	return u, d.loadUserCharacters(ctx, u)
}

func (d *postgresDB) loadUserCharacters(ctx context.Context, u *schema.User) error {
	// Active characters
	rows, err := d.db.Query(ctx,
		`SELECT slot, id FROM characters WHERE steam_id = $1 AND deleted_at IS NULL`,
		u.ID,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var slot int
		var id uuid.UUID
		if err := rows.Scan(&slot, &id); err != nil {
			return err
		}
		u.Characters[slot] = id
	}
	if err := rows.Err(); err != nil {
		return err
	}

	// Soft-deleted characters
	drows, err := d.db.Query(ctx,
		`SELECT slot, character_id FROM deleted_characters WHERE steam_id = $1`,
		u.ID,
	)
	if err != nil {
		return err
	}
	defer drows.Close()

	for drows.Next() {
		var slot int
		var id uuid.UUID
		if err := drows.Scan(&slot, &id); err != nil {
			return err
		}
		u.DeletedCharacters[slot] = id
	}
	return drows.Err()
}

func (d *postgresDB) SetUserFlags(steamid string, flags bitmask.Bitmask) error {
	ctx := context.Background()
	return d.execTx(ctx, func(tx pgx.Tx) error {
		ct, err := tx.Exec(ctx,
			`UPDATE users SET flags = $1 WHERE id = $2`, uint32(flags), steamid,
		)
		if err != nil {
			return err
		}
		if ct.RowsAffected() == 0 {
			return database.ErrNoDocument
		}
		return nil
	})
}

func (d *postgresDB) GetUserFlags(steamid string) (bitmask.Bitmask, error) {
	ctx := context.Background()
	var flags uint32
	err := d.db.QueryRow(ctx,
		`SELECT flags FROM users WHERE id = $1`, steamid,
	).Scan(&flags)
	if err == pgx.ErrNoRows {
		return 0, database.ErrNoDocument
	}
	if err != nil {
		return 0, fmt.Errorf("get user flags: %w", err)
	}
	return bitmask.Bitmask(flags), nil
}
