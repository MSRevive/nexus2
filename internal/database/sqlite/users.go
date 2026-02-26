package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/msrevive/nexus2/internal/bitmask"
	"github.com/msrevive/nexus2/internal/database"
	"github.com/msrevive/nexus2/pkg/database/schema"

	"github.com/google/uuid"
)

func (d *sqliteDB) GetAllUsers() ([]*schema.User, error) {
	rows, err := d.db.Query(`
		SELECT
			u.id, u.revision, u.flags,
			c.slot         AS char_slot,
			c.id           AS char_id,
			dc.slot        AS del_slot,
			dc.character_id AS del_char_id
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
			charID *string
			delSlot *int
			delID *string
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
			parsed, err := uuid.Parse(*charID)
			if err != nil {
				return nil, fmt.Errorf("bad character uuid %q: %w", *charID, err)
			}
			u.Characters[*charSlot] = parsed
		}
		if delSlot != nil && delID != nil {
			parsed, err := uuid.Parse(*delID)
			if err != nil {
				return nil, fmt.Errorf("bad deleted character uuid %q: %w", *delID, err)
			}
			u.DeletedCharacters[*delSlot] = parsed
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

func (d *sqliteDB) GetUser(steamid string) (*schema.User, error) {
	rows, err := d.db.Query(`
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
		WHERE u.id = ?`,
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
			charID *string
			delSlot *int
			delID *string
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
			parsed, err := uuid.Parse(*charID)
			if err != nil {
				return nil, fmt.Errorf("bad character uuid %q: %w", *charID, err)
			}
			u.Characters[*charSlot] = parsed
		}
		if delSlot != nil && delID != nil {
			parsed, err := uuid.Parse(*delID)
			if err != nil {
				return nil, fmt.Errorf("bad deleted character uuid %q: %w", *delID, err)
			}
			u.DeletedCharacters[*delSlot] = parsed
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

func (d *sqliteDB) SetUserFlags(steamid string, flags bitmask.Bitmask) error {
	return d.exec(func(tx *sql.Tx) error {
		res, err := tx.Exec(`UPDATE users SET flags = ? WHERE id = ?`, uint32(flags), steamid)
		if err != nil {
			return err
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			return database.ErrNoDocument
		}
		return nil
	})
}

func (d *sqliteDB) GetUserFlags(steamid string) (bitmask.Bitmask, error) {
	var flags uint32
	err := d.db.QueryRow(`SELECT flags FROM users WHERE id = ?`, steamid).Scan(&flags)
	if err == sql.ErrNoRows {
		return 0, database.ErrNoDocument
	}
	return bitmask.Bitmask(flags), err
}
