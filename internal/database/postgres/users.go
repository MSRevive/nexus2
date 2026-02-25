package postgres

import (
	"database/sql"
	"fmt"

	"github.com/msrevive/nexus2/internal/bitmask"
	"github.com/msrevive/nexus2/internal/database"
	"github.com/msrevive/nexus2/pkg/database/schema"

	"github.com/google/uuid"
)

func (d *sqliteDB) GetAllUsers() ([]*schema.User, error) {
	rows, err := d.db.Query(`SELECT id, revision, flags FROM users`)
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
		users = append(users, u)
	}
	return users, rows.Err()
}

func (d *sqliteDB) GetUser(steamid string) (*schema.User, error) {
	u := &schema.User{
		Characters:        make(map[int]uuid.UUID),
		DeletedCharacters: make(map[int]uuid.UUID),
	}

	err := d.db.QueryRow(
		`SELECT id, revision, flags FROM users WHERE id = ?`, steamid,
	).Scan(&u.ID, &u.Revision, &u.Flags)

	if err == sql.ErrNoRows {
		return nil, database.ErrNoDocument
	}
	if err != nil {
		return nil, err
	}

	return u, d.loadUserCharacters(u)
}

func (d *sqliteDB) loadUserCharacters(u *schema.User) error {
	// Active characters
	rows, err := d.db.Query(
		`SELECT slot, id FROM characters WHERE steam_id = ? AND deleted_at IS NULL`,
		u.ID,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var slot int
		var idStr string
		if err := rows.Scan(&slot, &idStr); err != nil {
			return err
		}
		id, err := uuid.Parse(idStr)
		if err != nil {
			return fmt.Errorf("loadUserCharacters: bad uuid %q: %w", idStr, err)
		}
		u.Characters[slot] = id
	}
	if err := rows.Err(); err != nil {
		return err
	}

	// Soft-deleted characters (the DeletedCharacters map)
	drows, err := d.db.Query(
		`SELECT slot, character_id FROM deleted_characters WHERE steam_id = ?`,
		u.ID,
	)
	if err != nil {
		return err
	}
	defer drows.Close()

	for drows.Next() {
		var slot int
		var idStr string
		if err := drows.Scan(&slot, &idStr); err != nil {
			return err
		}
		id, err := uuid.Parse(idStr)
		if err != nil {
			return fmt.Errorf("loadUserCharacters: bad deleted uuid %q: %w", idStr, err)
		}
		u.DeletedCharacters[slot] = id
	}
	return drows.Err()
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
