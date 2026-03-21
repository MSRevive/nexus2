package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/msrevive/nexus2/internal/database"
	"github.com/msrevive/nexus2/pkg/database/schema"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// NewCharacter creates the user row (if missing) and the character row in a
// single transaction so they are always consistent.
func (d *postgresDB) NewCharacter(steamid string, slot int, size int, data string) (uuid.UUID, error) {
	charID := uuid.New()
	now := time.Now().UTC()
	ctx := context.Background()

	err := d.execTx(ctx, func(tx pgx.Tx) error {
		// Upsert the user.
		_, err := tx.Exec(ctx,
			`INSERT INTO users (id) VALUES ($1) ON CONFLICT(id) DO NOTHING`,
			steamid,
		)
		if err != nil {
			return fmt.Errorf("upsert user: %w", err)
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO characters
				(id, steam_id, slot, created_at, data_created_at, data_size, data_payload)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			charID, steamid, slot, now, now, size, data,
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

// UpdateCharacter stores the latest state in the coalescing map. The next
// flushWorker tick will commit all coalesced updates in a single transaction.
func (d *postgresDB) UpdateCharacter(id uuid.UUID, size int, data string, backupMax int, backupTime time.Duration) error {
	d.coalesceMu.Lock()
	d.pendingUpdates[id] = pendingUpdate{
		size:       size,
		data:       data,
		backupMax:  backupMax,
		backupTime: backupTime,
	}
	d.coalesceMu.Unlock()
	return nil
}

// applyCharacterUpdate is called inside the flush transaction. It performs the
// read-modify-write cycle for one character, applying version/backup logic.
func applyCharacterUpdate(ctx context.Context, tx pgx.Tx, id uuid.UUID, upd pendingUpdate) error {
	var (
		dataCreatedAt time.Time
		dataSize      int
		dataPayload   string
	)
	err := tx.QueryRow(ctx, `
		SELECT data_created_at, data_size, data_payload
		FROM characters WHERE id = $1
		FOR UPDATE`, // we do FOR UPDATE to let postgres know to lock it ahead of time for updating.
		id,
	).Scan(&dataCreatedAt, &dataSize, &dataPayload)

	if err == pgx.ErrNoRows {
		return database.ErrNoDocument
	}
	if err != nil {
		return err
	}

	// ------------------------------------------------------------------
	// Version / backup logic
	// ------------------------------------------------------------------
	if upd.backupMax > 0 {
		var versionCount int
		if err := tx.QueryRow(ctx,
			`SELECT COUNT(*) FROM character_versions WHERE character_id = $1`,
			id,
		).Scan(&versionCount); err != nil {
			return err
		}

		// If at the cap, delete the oldest entry.
		if versionCount >= upd.backupMax {
			if _, err := tx.Exec(ctx, `
				DELETE FROM character_versions WHERE id = (
					SELECT id FROM character_versions
					WHERE character_id = $1 ORDER BY id ASC LIMIT 1
				)`, id,
			); err != nil {
				return err
			}
			versionCount--
		}

		if versionCount > 0 {
			var newestCreatedAt time.Time
			err := tx.QueryRow(ctx, `
				SELECT created_at FROM character_versions
				WHERE character_id = $1 ORDER BY id DESC LIMIT 1`,
				id,
			).Scan(&newestCreatedAt)
			if err != nil {
				return err
			}

			if dataCreatedAt.After(newestCreatedAt.Add(upd.backupTime)) {
				if _, err := tx.Exec(ctx, `
					INSERT INTO character_versions (character_id, created_at, size, data_payload)
					VALUES ($1, $2, $3, $4)`,
					id, dataCreatedAt, dataSize, dataPayload,
				); err != nil {
					return err
				}
			}
		} else {
			// No versions yet — always snapshot.
			if _, err := tx.Exec(ctx, `
				INSERT INTO character_versions (character_id, created_at, size, data_payload)
				VALUES ($1, $2, $3, $4)`,
				id, dataCreatedAt, dataSize, dataPayload,
			); err != nil {
				return err
			}
		}
	}

	// Write the new current character data.
	_, err = tx.Exec(ctx, `
		UPDATE characters
		SET data_created_at = $1, data_size = $2, data_payload = $3
		WHERE id = $4`,
		time.Now().UTC(), upd.size, upd.data, id,
	)
	return err
}

func (d *postgresDB) GetCharacter(id uuid.UUID) (*schema.Character, error) {
	ctx := context.Background()
	c := &schema.Character{ID: id}

	var (
		steamID pgtype.Text
		slot pgtype.Int4
		deletedAt pgtype.Timestamptz
	)

	err := d.db.QueryRow(ctx, `
		SELECT steam_id, slot, created_at, deleted_at,
			data_created_at, data_size, data_payload
		FROM characters WHERE id = $1`,
		id,
	).Scan(
		&steamID, &slot, &c.CreatedAt, &deletedAt,
		&c.Data.CreatedAt, &c.Data.Size, &c.Data.Data,
	)
	if err == pgx.ErrNoRows {
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

	// Overlay any pending update that hasn't been flushed yet.
	d.coalesceMu.RLock()
	if upd, ok := d.pendingUpdates[id]; ok {
		c.Data.Size = upd.size
		c.Data.Data = upd.data
	}
	d.coalesceMu.RUnlock()

	// Load version history.
	rows, err := d.db.Query(ctx, `
		SELECT created_at, size, data_payload
		FROM character_versions
		WHERE character_id = $1 ORDER BY id ASC`,
		id,
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

func (d *postgresDB) GetCharacters(steamid string) (map[int]schema.Character, error) {
	ctx := context.Background()
	rows, err := d.db.Query(ctx, `
		SELECT id, slot, created_at, deleted_at, data_created_at, data_size, data_payload
		FROM characters
		WHERE steam_id = $1 AND deleted_at IS NULL`,
		steamid,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	chars := make(map[int]schema.Character)
	for rows.Next() {
		var (
			c schema.Character
			deletedAt *time.Time
		)
		err := rows.Scan(
			&c.ID, &c.Slot, &c.CreatedAt, &deletedAt,
			&c.Data.CreatedAt, &c.Data.Size, &c.Data.Data,
		)
		if err != nil {
			return nil, err
		}
		c.SteamID = steamid
		c.DeletedAt = deletedAt

		// Overlay any pending update that hasn't been flushed yet.
		d.coalesceMu.RLock()
		if upd, ok := d.pendingUpdates[c.ID]; ok {
			c.Data.Size = upd.size
			c.Data.Data = upd.data
		}
		d.coalesceMu.RUnlock()

		chars[c.Slot] = c
	}
	return chars, rows.Err()
}

func (d *postgresDB) LookUpCharacterID(steamid string, slot int) (uuid.UUID, error) {
	ctx := context.Background()
	var id uuid.UUID
	err := d.db.QueryRow(ctx, `
		SELECT id FROM characters
		WHERE steam_id = $1 AND slot = $2 AND deleted_at IS NULL`,
		steamid, slot,
	).Scan(&id)
	if err == pgx.ErrNoRows {
		return uuid.Nil, database.ErrNoDocument
	}
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

// SoftDeleteCharacter sets deleted_at + expires_at on the character and records
// the slot in deleted_characters so it can be restored or GC'd later.
func (d *postgresDB) SoftDeleteCharacter(id uuid.UUID, expiration time.Duration) error {
	now := time.Now().UTC()
	expiresAt := now.Add(expiration)
	ctx := context.Background()

	return d.execTx(ctx, func(tx pgx.Tx) error {
		var steamID string
		var slot int
		err := tx.QueryRow(ctx,
			`SELECT steam_id, slot FROM characters WHERE id = $1`, id,
		).Scan(&steamID, &slot)
		if err == pgx.ErrNoRows {
			return database.ErrNoDocument
		}
		if err != nil {
			return err
		}

		// we orphan the character data here.
		if _, err := tx.Exec(ctx, `
			UPDATE characters SET deleted_at = $1, expires_at = $2, steam_id = NULL, slot = NULL WHERE id = $3`,
			now, expiresAt, id,
		); err != nil {
			return err
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO deleted_characters (steam_id, slot, character_id, deleted_at)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (steam_id, slot) DO UPDATE
			SET character_id = EXCLUDED.character_id,
			    deleted_at   = EXCLUDED.deleted_at`,
			steamID, slot, id, now,
		)
		return err
	})
}

// DeleteCharacter permanently removes the character and all associated data.
func (d *postgresDB) DeleteCharacter(id uuid.UUID) error {
	ctx := context.Background()
	return d.execTx(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `DELETE FROM characters WHERE id = $1`, id)
		return err
	})
}

// DeleteCharacterReference removes the active slot→character mapping for a user.
func (d *postgresDB) DeleteCharacterReference(steamid string, slot int) error {
	ctx := context.Background()
	return d.execTx(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			UPDATE characters SET steam_id = NULL, slot = NULL
			WHERE steam_id = $1 AND slot = $2 AND deleted_at IS NULL`,
			steamid, slot,
		)
		return err
	})
}

// MoveCharacter transfers a character to a different user/slot atomically.
func (d *postgresDB) MoveCharacter(id uuid.UUID, steamid string, slot int) error {
	ctx := context.Background()
	return d.execTx(ctx, func(tx pgx.Tx) error {
		var oldSteamID string
		var oldSlot int
		err := tx.QueryRow(ctx,
			`SELECT steam_id, slot FROM characters WHERE id = $1`, id,
		).Scan(&oldSteamID, &oldSlot)
		if err == pgx.ErrNoRows {
			return database.ErrNoDocument
		}
		if err != nil {
			return err
		}

		// Ensure target user exists.
		var exists int
		if err := tx.QueryRow(ctx,
			`SELECT COUNT(*) FROM users WHERE id = $1`, steamid,
		).Scan(&exists); err != nil {
			return err
		}
		if exists == 0 {
			return database.ErrNoDocument
		}

		// Clear old slot.
		if _, err := tx.Exec(ctx, `
			UPDATE characters SET steam_id = NULL, slot = NULL
			WHERE steam_id = $1 AND slot = $2 AND deleted_at IS NULL`,
			oldSteamID, oldSlot,
		); err != nil {
			return err
		}

		// Assign to new owner.
		_, err = tx.Exec(ctx, `
			UPDATE characters SET steam_id = $1, slot = $2, deleted_at = NULL
			WHERE id = $3`,
			steamid, slot, id,
		)
		return err
	})
}

// CopyCharacter duplicates a character's current data under a new UUID.
func (d *postgresDB) CopyCharacter(id uuid.UUID, steamid string, slot int) (uuid.UUID, error) {
	newID := uuid.New()
	now := time.Now().UTC()
	ctx := context.Background()

	err := d.execTx(ctx, func(tx pgx.Tx) error {
		var dataCreatedAt time.Time
		var dataSize int
		var dataPayload string
		err := tx.QueryRow(ctx, `
			SELECT data_created_at, data_size, data_payload
			FROM characters WHERE id = $1`,
			id,
		).Scan(&dataCreatedAt, &dataSize, &dataPayload)
		if err == pgx.ErrNoRows {
			return database.ErrNoDocument
		}
		if err != nil {
			return err
		}

		if _, err := tx.Exec(ctx,
			`INSERT INTO users (id) VALUES ($1) ON CONFLICT(id) DO NOTHING`, steamid,
		); err != nil {
			return err
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO characters
				(id, steam_id, slot, created_at, data_created_at, data_size, data_payload)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			newID, steamid, slot, now, dataCreatedAt, dataSize, dataPayload,
		)
		return err
	})
	if err != nil {
		return uuid.Nil, err
	}
	return newID, nil
}

// RestoreCharacter clears the soft-delete markers and makes the character active again.
func (d *postgresDB) RestoreCharacter(id uuid.UUID) error {
	ctx := context.Background()
	return d.execTx(ctx, func(tx pgx.Tx) error {
		var steamID string
		var slot int
		err := tx.QueryRow(ctx,
			`SELECT steam_id, slot FROM deleted_characters WHERE character_id = $1`, id,
		).Scan(&steamID, &slot)
		if err == pgx.ErrNoRows {
			return database.ErrNoDocument
		}
		if err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, `
			UPDATE characters SET deleted_at = NULL, expires_at = NULL, steam_id = $1, slot = $2 WHERE id = $3`,
			steamID, slot, id,
		); err != nil {
			return err
		}

		_, err = tx.Exec(ctx,
			`DELETE FROM deleted_characters WHERE character_id = $1`,
			id,
		)
		return err
	})
}

// RollbackCharacter replaces the current character data with the version at
// index ver (0-based, ordered oldest → newest).
func (d *postgresDB) RollbackCharacter(id uuid.UUID, ver int) error {
	ctx := context.Background()
	return d.execTx(ctx, func(tx pgx.Tx) error {
		var createdAt time.Time
		var size int
		var payload string
		err := tx.QueryRow(ctx, `
			SELECT created_at, size, data_payload
			FROM character_versions
			WHERE character_id = $1
			ORDER BY id ASC
			LIMIT 1 OFFSET $2`,
			id, ver,
		).Scan(&createdAt, &size, &payload)
		if err == pgx.ErrNoRows {
			return fmt.Errorf("no character version at index %d", ver)
		}
		if err != nil {
			return err
		}

		_, err = tx.Exec(ctx, `
			UPDATE characters
			SET data_created_at = $1, data_size = $2, data_payload = $3
			WHERE id = $4`,
			createdAt, size, payload, id,
		)
		return err
	})
}

// RollbackCharacterToLatest replaces the current data with the most recent version.
func (d *postgresDB) RollbackCharacterToLatest(id uuid.UUID) error {
	ctx := context.Background()
	return d.execTx(ctx, func(tx pgx.Tx) error {
		var createdAt time.Time
		var size int
		var payload string
		err := tx.QueryRow(ctx, `
			SELECT created_at, size, data_payload
			FROM character_versions
			WHERE character_id = $1
			ORDER BY id DESC LIMIT 1`,
			id,
		).Scan(&createdAt, &size, &payload)
		if err == pgx.ErrNoRows {
			return fmt.Errorf("no character backups exist")
		}
		if err != nil {
			return err
		}

		_, err = tx.Exec(ctx, `
			UPDATE characters
			SET data_created_at = $1, data_size = $2, data_payload = $3
			WHERE id = $4`,
			createdAt, size, payload, id,
		)
		return err
	})
}

// DeleteCharacterVersions wipes all version history for a character.
func (d *postgresDB) DeleteCharacterVersions(id uuid.UUID) error {
	ctx := context.Background()
	return d.execTx(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`DELETE FROM character_versions WHERE character_id = $1`, id,
		)
		return err
	})
}

func (d *postgresDB) GetRollbackVersionsTimestamp(id uuid.UUID) (map[int]string, error) {
	ctx := context.Background()
	rows, err := d.db.Query(ctx, `
		SELECT created_at
		FROM character_versions
		WHERE character_id = $1
		ORDER BY id ASC`,
		id,
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