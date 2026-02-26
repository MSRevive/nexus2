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

	rows, err := d.db.Query(ctx, `
		SELECT
			u.id, u.revision, u.flags,
			c.slot           AS char_slot,
			c.id             AS char_id,
			dc.slot          AS del_slot,
			dc.character_id  AS del_char_id
		FROM users u
		LEFT JOIN characters c
			ON c.steam_id = u.id AND c.deleted_at IS NULL
		LEFT JOIN deleted_characters dc
			ON dc.steam_id = u.id
		ORDER BY u.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	userMap := make(map[string]*schema.User)
	var order []string

	for rows.Next() {
		var (
			id string
			revision int
			flags uint32
			charSlot *int
			charID *uuid.UUID
			delSlot *int
			delID *uuid.UUID
		)
		if err := rows.Scan(&id, &revision, &flags, &charSlot, &charID, &delSlot, &delID); err != nil {
			return nil, err
		}

		u, exists := userMap[id]
		if !exists {
			u = &schema.User{
				ID: id,
				Revision: revision,
				Flags: flags,
				Characters: make(map[int]uuid.UUID),
				DeletedCharacters: make(map[int]uuid.UUID),
			}
			userMap[id] = u
			order = append(order, id)
		}

		if charSlot != nil && charID != nil {
			u.Characters[*charSlot] = *charID
		}
		if delSlot != nil && delID != nil {
			u.DeletedCharacters[*delSlot] = *delID
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	users := make([]*schema.User, 0, len(order))
	for _, id := range order {
		users = append(users, userMap[id])
	}
	return users, nil
}

func (d *postgresDB) GetUser(steamid string) (*schema.User, error) {
	ctx := context.Background()

	rows, err := d.db.Query(ctx, `
		SELECT
			u.id, u.revision, u.flags,
			c.slot          AS char_slot,
			c.id            AS char_id,
			dc.slot         AS del_slot,
			dc.character_id AS del_char_id
		FROM users u
		LEFT JOIN characters c
			ON c.steam_id = u.id AND c.deleted_at IS NULL
		LEFT JOIN deleted_characters dc
			ON dc.steam_id = u.id
		WHERE u.id = $1`,
		steamid,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var u *schema.User
	for rows.Next() {
		var (
			id string
			revision int
			flags uint32
			charSlot *int
			charID *uuid.UUID
			delSlot *int
			delID *uuid.UUID
		)
		if err := rows.Scan(&id, &revision, &flags, &charSlot, &charID, &delSlot, &delID); err != nil {
			return nil, err
		}

		if u == nil {
			u = &schema.User{
				ID: id,
				Revision: revision,
				Flags: flags,
				Characters: make(map[int]uuid.UUID),
				DeletedCharacters: make(map[int]uuid.UUID),
			}
		}

		if charSlot != nil && charID != nil {
			u.Characters[*charSlot] = *charID
		}
		if delSlot != nil && delID != nil {
			u.DeletedCharacters[*delSlot] = *delID
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if u == nil {
		return nil, database.ErrNoDocument
	}
	return u, nil
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
