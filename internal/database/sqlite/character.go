package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/msrevive/nexus2/internal/database"
	"github.com/msrevive/nexus2/pkg/database/schema"

	"github.com/google/uuid"
	"github.com/titpetric/oida"
)

// NewCharacter creates the user row (if missing) and the character row in a
// single transaction so they are always consistent.
func (d *sqliteDB) NewCharacter(ctx context.Context, steamid string, slot int, size int, data string) (uuid.UUID, error) {
	ctx, span := oida.Start(ctx, "INSERT character", oida.KindDatabase)
	defer span.End()
	span.SetAttribute("steamid", steamid)
	span.SetAttribute("slot", slot)
	span.SetAttribute("size", size)

	charID := uuid.New()
	now := time.Now().UTC()

	err := d.exec(ctx, func(ctx context.Context, tx *sql.Tx) error {
		// Upsert the user — mirrors the pebble logic that creates a new user
		// document when one doesn't exist yet.
		_, err := tx.ExecContext(ctx, 
			`INSERT INTO users (id) VALUES (?) ON CONFLICT(id) DO NOTHING`,
			steamid,
		)
		if err != nil {
			return fmt.Errorf("upsert user: %w", err)
		}

		_, err = tx.ExecContext(ctx, `
			INSERT INTO characters
				(id, steam_id, slot, created_at, data_created_at, data_size, data_payload)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			charID.String(), steamid, slot, now, now, size, data,
		)
		if err != nil {
			return fmt.Errorf("insert character: %w", err)
		}

		return nil
	})
	if err != nil {
		return uuid.Nil, err
	}
	return charID, nil
}

// UpdateCharacter does NOT touch the database immediately. It stores the latest
// state for this character ID in the coalescing map and returns. The next
// flushWorker tick will commit all coalesced updates in a single transaction.
//
// This means 100 calls to UpdateCharacter for the same character within the
// flush window result in exactly 1 database write — the one with the final state.
func (d *sqliteDB) UpdateCharacter(ctx context.Context, id uuid.UUID, size int, data string, backupMax int, backupTime time.Duration) error {
	_, span := oida.Start(ctx, "QUEUE character update", oida.KindDatabase)
	defer span.End()
	span.SetAttribute("uuid", id.String())
	span.SetAttribute("size", size)

	d.coalesceMu.Lock()
	d.pendingUpdates[id] = pendingUpdate{
		size:       size,
		data:       data,
		backupMax:  backupMax,
		backupTime: backupTime,
	}
	pending := len(d.pendingUpdates)
	d.coalesceMu.Unlock()

	span.SetAttribute("pending", pending)
	return nil
}

// applyCharacterUpdate is called inside the flush transaction. It performs the
// read-modify-write cycle for one character, applying the version/backup logic
// that mirrors the pebble implementation exactly.
//
// Called only from within a transaction on the write goroutine.
func applyCharacterUpdate(ctx context.Context, tx *sql.Tx, id uuid.UUID, upd pendingUpdate) error {
	// Read the current character data so we can snapshot it as a version.
	var (
		dataCreatedAt time.Time
		dataSize      int
		dataPayload   string
	)
	err := tx.QueryRowContext(ctx, `
		SELECT data_created_at, data_size, data_payload
		FROM characters WHERE id = ?`,
		id.String(),
	).Scan(&dataCreatedAt, &dataSize, &dataPayload)

	if err == sql.ErrNoRows {
		return database.ErrNoDocument
	}
	if err != nil {
		return err
	}

	// ------------------------------------------------------------------
	// Version / backup logic — mirrors the pebble UpdateCharacter exactly.
	// ------------------------------------------------------------------
	if upd.backupMax > 0 {
		var versionCount int
		if err := tx.QueryRowContext(ctx, 
			`SELECT COUNT(*) FROM character_versions WHERE character_id = ?`,
			id.String(),
		).Scan(&versionCount); err != nil {
			return err
		}

		// If we are at the cap, delete the oldest entry (lowest autoincrement id).
		if versionCount >= upd.backupMax {
			if _, err := tx.ExecContext(ctx, `
				DELETE FROM character_versions WHERE id = (
					SELECT id FROM character_versions
					WHERE character_id = ? ORDER BY id ASC LIMIT 1
				)`, id.String(),
			); err != nil {
				return err
			}
			versionCount--
		}

		if versionCount > 0 {
			// Only snapshot the current data if enough time has passed since
			// the newest existing backup. This prevents rapid-fire updates from
			// flooding the version table.
			var newestCreatedAt time.Time
			err := tx.QueryRowContext(ctx, `
				SELECT created_at FROM character_versions
				WHERE character_id = ? ORDER BY id DESC LIMIT 1`,
				id.String(),
			).Scan(&newestCreatedAt)
			if err != nil {
				return err
			}

			if dataCreatedAt.After(newestCreatedAt.Add(upd.backupTime)) {
				if _, err := tx.ExecContext(ctx, `
					INSERT INTO character_versions (character_id, created_at, size, data_payload)
					VALUES (?, ?, ?, ?)`,
					id.String(), dataCreatedAt, dataSize, dataPayload,
				); err != nil {
					return err
				}
			}
		} else {
			// No versions yet — always snapshot the current data on the first update.
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO character_versions (character_id, created_at, size, data_payload)
				VALUES (?, ?, ?, ?)`,
				id.String(), dataCreatedAt, dataSize, dataPayload,
			); err != nil {
				return err
			}
		}
	}

	// Write the new current character data.
	_, err = tx.ExecContext(ctx, `
		UPDATE characters
		SET data_created_at = ?, data_size = ?, data_payload = ?
		WHERE id = ?`,
		time.Now().UTC(), upd.size, upd.data, id.String(),
	)
	return err
}

func (d *sqliteDB) GetCharacter(ctx context.Context, id uuid.UUID) (*schema.Character, error) {
	ctx, span := oida.Start(ctx, "SELECT character", oida.KindDatabase)
	defer span.End()
	span.SetAttribute("uuid", id.String())

	c := &schema.Character{ID: id}

	var (
		steamID sql.NullString
		slot sql.NullInt32
		deletedAt sql.NullTime
	)

	err := d.db.QueryRowContext(ctx, `
		SELECT steam_id, slot, created_at, deleted_at,
		    data_created_at, data_size, data_payload
		FROM characters WHERE id = ?`,
		id.String(),
	).Scan(
		&steamID, &slot, &c.CreatedAt, &deletedAt,
		&c.Data.CreatedAt, &c.Data.Size, &c.Data.Data,
	)
	if err == sql.ErrNoRows {
		return nil, database.ErrNoDocument
	}
	if err != nil {
		return nil, err
	}
	c.SteamID = steamID.String
	c.Slot = int(slot.Int32)
	if deletedAt.Valid {
		c.DeletedAt = &deletedAt.Time
	}

	// Load the versions slice (Versions []CharacterData).
	rows, err := d.db.QueryContext(ctx, `
		SELECT created_at, size, data_payload
		FROM character_versions
		WHERE character_id = ? ORDER BY id ASC`,
		id.String(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var v schema.CharacterData
		if err := rows.Scan(&v.CreatedAt, &v.Size, &v.Data); err != nil {
			return nil, err
		}
		c.Versions = append(c.Versions, v)
	}
	return c, rows.Err()
}

func (d *sqliteDB) GetCharacters(ctx context.Context, steamid string) (map[int]schema.Character, error) {
	ctx, span := oida.Start(ctx, "SELECT characters", oida.KindDatabase)
	defer span.End()
	span.SetAttribute("steamid", steamid)

	rows, err := d.db.QueryContext(ctx, `
		SELECT id, slot, created_at, deleted_at, data_created_at, data_size, data_payload
		FROM characters
		WHERE steam_id = ? AND deleted_at IS NULL`,
		steamid,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	chars := make(map[int]schema.Character)
	for rows.Next() {
		var (
			c         schema.Character
			idStr     string
			deletedAt sql.NullTime
		)
		err := rows.Scan(
			&idStr, &c.Slot, &c.CreatedAt, &deletedAt,
			&c.Data.CreatedAt, &c.Data.Size, &c.Data.Data,
		)
		if err != nil {
			return nil, err
		}
		c.ID, _ = uuid.Parse(idStr)
		c.SteamID = steamid
		if deletedAt.Valid {
			c.DeletedAt = &deletedAt.Time
		}
		chars[c.Slot] = c
	}
	return chars, rows.Err()
}

func (d *sqliteDB) LookUpCharacterID(ctx context.Context, steamid string, slot int) (uuid.UUID, error) {
	ctx, span := oida.Start(ctx, "SELECT character id", oida.KindDatabase)
	defer span.End()
	span.SetAttribute("steamid", steamid)
	span.SetAttribute("slot", slot)

	var idStr string
	err := d.db.QueryRowContext(ctx, `
		SELECT id FROM characters
		WHERE steam_id = ? AND slot = ? AND deleted_at IS NULL`,
		steamid, slot,
	).Scan(&idStr)
	if err == sql.ErrNoRows {
		return uuid.Nil, database.ErrNoDocument
	}
	if err != nil {
		return uuid.Nil, err
	}
	return uuid.Parse(idStr)
}

// SoftDeleteCharacter sets deleted_at + expires_at on the character and records
// the slot in deleted_characters so it can be restored or GC'd later.
func (d *sqliteDB) SoftDeleteCharacter(ctx context.Context, id uuid.UUID, expiration time.Duration) error {
	ctx, span := oida.Start(ctx, "UPDATE character deleted", oida.KindDatabase)
	defer span.End()
	span.SetAttribute("uuid", id.String())

	now := time.Now().UTC()
	expiresAt := now.Add(expiration)

	return d.exec(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var steamID string
		var slot int
		err := tx.QueryRowContext(ctx, 
			`SELECT steam_id, slot FROM characters WHERE id = ?`, id.String(),
		).Scan(&steamID, &slot)
		if err == sql.ErrNoRows {
			return database.ErrNoDocument
		}
		if err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, `
			UPDATE characters SET deleted_at = ?, expires_at = ?, steam_id = NULL, slot = NULL WHERE id = ?`,
			now, expiresAt, id.String(),
		); err != nil {
			return err
		}

		// Upsert into deleted_characters to preserve the slot → id mapping
		// (mirrors user.DeletedCharacters in the pebble implementation).
		_, err = tx.ExecContext(ctx, `
			INSERT INTO deleted_characters (steam_id, slot, character_id, deleted_at)
			VALUES (?, ?, ?, ?)
			ON CONFLICT (steam_id, slot) DO UPDATE
			SET character_id = excluded.character_id,
			    deleted_at   = excluded.deleted_at`,
			steamID, slot, id.String(), now,
		)
		return err
	})
}

// DeleteCharacter permanently removes the character and all associated data.
// cascade on character_versions handles version cleanup automatically.
func (d *sqliteDB) DeleteCharacter(ctx context.Context, id uuid.UUID) error {
	ctx, span := oida.Start(ctx, "DELETE character", oida.KindDatabase)
	defer span.End()
	span.SetAttribute("uuid", id.String())

	return d.exec(ctx, func(ctx context.Context, tx *sql.Tx) error {
		// character_versions are deleted by ON DELETE CASCADE.
		_, err := tx.ExecContext(ctx, `DELETE FROM characters WHERE id = ?`, id.String())
		return err
	})
}

// DeleteCharacterReference removes the active slot→character mapping for a user,
// leaving the character row intact but unowned (steam_id = NULL).
// This is called by MoveCharacter to clear the character's old slot before
// reassigning it, mirroring delete(user.Characters, slot) in the pebble version.
func (d *sqliteDB) DeleteCharacterReference(ctx context.Context, steamid string, slot int) error {
	ctx, span := oida.Start(ctx, "UPDATE character reference", oida.KindDatabase)
	defer span.End()
	span.SetAttribute("steamid", steamid)
	span.SetAttribute("slot", slot)

	return d.exec(ctx, func(ctx context.Context, tx *sql.Tx) error {
		// Nullify steam_id/slot so the character no longer occupies the slot
		// on the old owner. The UNIQUE(steam_id, slot) constraint allows NULLs
		// on both columns, so this is safe.
		_, err := tx.ExecContext(ctx, `
			UPDATE characters SET steam_id = NULL, slot = NULL
			WHERE steam_id = ? AND slot = ? AND deleted_at IS NULL`,
			steamid, slot,
		)
		return err
	})
}

// MoveCharacter transfers a character to a different user/slot atomically.
func (d *sqliteDB) MoveCharacter(ctx context.Context, id uuid.UUID, steamid string, slot int) error {
	ctx, span := oida.Start(ctx, "UPDATE character owner", oida.KindDatabase)
	defer span.End()
	span.SetAttribute("uuid", id.String())
	span.SetAttribute("steamid", steamid)
	span.SetAttribute("slot", slot)

	return d.exec(ctx, func(ctx context.Context, tx *sql.Tx) error {
		// Fetch the character's current owner so we can clear that slot.
		var oldSteamID string
		var oldSlot int
		err := tx.QueryRowContext(ctx, 
			`SELECT steam_id, slot FROM characters WHERE id = ?`, id.String(),
		).Scan(&oldSteamID, &oldSlot)
		if err == sql.ErrNoRows {
			return database.ErrNoDocument
		}
		if err != nil {
			return err
		}

		// Ensure the target user exists.
		var exists int
		if err := tx.QueryRowContext(ctx, 
			`SELECT COUNT(*) FROM users WHERE id = ?`, steamid,
		).Scan(&exists); err != nil {
			return err
		}
		if exists == 0 {
			return database.ErrNoDocument
		}

		// Clear the old owner's reference by nullifying steam_id/slot so the
		// UNIQUE constraint on (steam_id, slot) doesn't block the reassignment.
		if _, err := tx.ExecContext(ctx, `
			UPDATE characters SET steam_id = NULL, slot = NULL
			WHERE steam_id = ? AND slot = ? AND deleted_at IS NULL`,
			oldSteamID, oldSlot,
		); err != nil {
			return err
		}

		// Now assign the character to the new owner.
		_, err = tx.ExecContext(ctx, `
			UPDATE characters SET steam_id = ?, slot = ?, deleted_at = NULL
			WHERE id = ?`,
			steamid, slot, id.String(),
		)
		return err
	})
}

// CopyCharacter duplicates a character's current data under a new UUID
// assigned to the target user/slot.
func (d *sqliteDB) CopyCharacter(ctx context.Context, id uuid.UUID, steamid string, slot int) (uuid.UUID, error) {
	ctx, span := oida.Start(ctx, "INSERT character copy", oida.KindDatabase)
	defer span.End()
	span.SetAttribute("uuid", id.String())
	span.SetAttribute("steamid", steamid)
	span.SetAttribute("slot", slot)

	newID := uuid.New()
	now := time.Now().UTC()

	err := d.exec(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var dataCreatedAt time.Time
		var dataSize int
		var dataPayload string
		err := tx.QueryRowContext(ctx, `
			SELECT data_created_at, data_size, data_payload
			FROM characters WHERE id = ?`,
			id.String(),
		).Scan(&dataCreatedAt, &dataSize, &dataPayload)
		if err == sql.ErrNoRows {
			return database.ErrNoDocument
		}
		if err != nil {
			return err
		}

		// Ensure the target user exists.
		if _, err := tx.ExecContext(ctx, 
			`INSERT INTO users (id) VALUES (?) ON CONFLICT(id) DO NOTHING`, steamid,
		); err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, `
			INSERT INTO characters
				(id, steam_id, slot, created_at, data_created_at, data_size, data_payload)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			newID.String(), steamid, slot, now, dataCreatedAt, dataSize, dataPayload,
		)
		return err
	})
	if err != nil {
		return uuid.Nil, err
	}
	return newID, nil
}

// RestoreCharacter clears the soft-delete markers and removes the entry from
// deleted_characters, making the character active again.
func (d *sqliteDB) RestoreCharacter(ctx context.Context, id uuid.UUID) error {
	ctx, span := oida.Start(ctx, "UPDATE character restored", oida.KindDatabase)
	defer span.End()
	span.SetAttribute("uuid", id.String())

	return d.exec(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var steamID string
		var slot int
		err := tx.QueryRowContext(ctx, 
			`SELECT steam_id, slot FROM deleted_characters WHERE character_id = ?`, id.String(),
		).Scan(&steamID, &slot)
		if err == sql.ErrNoRows {
			return database.ErrNoDocument
		}
		if err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, `
			UPDATE characters SET deleted_at = NULL, expires_at = NULL, steam_id = ?, slot = ? WHERE id = ?`,
			steamID, slot, id.String(),
		); err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, 
			`DELETE FROM deleted_characters WHERE character_id = ?`,
			id,
		)
		return err
	})
}

// RollbackCharacter replaces the current character data with the version at
// index ver (0-based, ordered oldest → newest). Mirrors the pebble implementation.
func (d *sqliteDB) RollbackCharacter(ctx context.Context, id uuid.UUID, ver int) error {
	ctx, span := oida.Start(ctx, "UPDATE character rollback", oida.KindDatabase)
	defer span.End()
	span.SetAttribute("uuid", id.String())
	span.SetAttribute("version", ver)

	return d.exec(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var createdAt time.Time
		var size int
		var payload string
		err := tx.QueryRowContext(ctx, `
			SELECT created_at, size, data_payload
			FROM character_versions
			WHERE character_id = ?
			ORDER BY id ASC
			LIMIT 1 OFFSET ?`,
			id.String(), ver,
		).Scan(&createdAt, &size, &payload)
		if err == sql.ErrNoRows {
			return fmt.Errorf("no character version at index %d", ver)
		}
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, `
			UPDATE characters
			SET data_created_at = ?, data_size = ?, data_payload = ?
			WHERE id = ?`,
			createdAt, size, payload, id.String(),
		)
		return err
	})
}

// RollbackCharacterToLatest replaces the current data with the most recent version.
func (d *sqliteDB) RollbackCharacterToLatest(ctx context.Context, id uuid.UUID) error {
	ctx, span := oida.Start(ctx, "UPDATE character rollback latest", oida.KindDatabase)
	defer span.End()
	span.SetAttribute("uuid", id.String())

	return d.exec(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var createdAt time.Time
		var size int
		var payload string
		err := tx.QueryRowContext(ctx, `
			SELECT created_at, size, data_payload
			FROM character_versions
			WHERE character_id = ?
			ORDER BY id DESC LIMIT 1`,
			id.String(),
		).Scan(&createdAt, &size, &payload)
		if err == sql.ErrNoRows {
			return fmt.Errorf("no character backups exist")
		}
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, `
			UPDATE characters
			SET data_created_at = ?, data_size = ?, data_payload = ?
			WHERE id = ?`,
			createdAt, size, payload, id.String(),
		)
		return err
	})
}

// DeleteCharacterVersions wipes all version history for a character.
func (d *sqliteDB) DeleteCharacterVersions(ctx context.Context, id uuid.UUID) error {
	ctx, span := oida.Start(ctx, "DELETE character versions", oida.KindDatabase)
	defer span.End()
	span.SetAttribute("uuid", id.String())

	return d.exec(ctx, func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, 
			`DELETE FROM character_versions WHERE character_id = ?`, id.String(),
		)
		return err
	})
}

func (d *sqliteDB) GetRollbackVersionsTimestamp(ctx context.Context, id uuid.UUID) (map[int]string, error) {
	ctx, span := oida.Start(ctx, "SELECT character version timestamps", oida.KindDatabase)
	defer span.End()
	span.SetAttribute("uuid", id.String())

	rows, err := d.db.QueryContext(ctx, `
		SELECT created_at
		FROM character_versions
		WHERE character_id = ?
		ORDER BY id ASC`,
		id.String(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	versions := make(map[int]string)
	idx := 0
	for rows.Next() {
		var createdAt time.Time
		if err := rows.Scan(&createdAt); err != nil {
			return nil, err
		}
		versions[idx] = createdAt.UTC().Format(time.RFC3339)
		idx++
	}
	return versions, rows.Err()
}
