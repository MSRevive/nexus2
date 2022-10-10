package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/msrevive/nexus2/internal/ent"
	"github.com/msrevive/nexus2/internal/ent/character"
	"github.com/msrevive/nexus2/internal/ent/player"
)

type characterRepository struct {
	db *sqlx.DB
}

func NewCharacterRepository(ctx context.Context, db *sql.DB) *characterRepository {
	return &characterRepository{
		db: db,
	}
}

func (r *characterRepository) Debug() error {
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

// CharactersGetAll returns all latest, active Characters
func (r *characterRepository) CharactersGetAll(ctx context.Context) ([]*ent.DeprecatedCharacter, error) {
	chars, err := r.client.Character.Query().
		WithPlayer().
		Where(
			character.And(
				character.Version(1),
				character.DeletedAtIsNil(),
			),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	return charsToDepChars(chars), nil
}

// CharactersGetBySteamid returns all Characters associated to a Player with the provided Steam ID
func (r *characterRepository) CharactersGetBySteamid(ctx context.Context, sid string) ([]*ent.DeprecatedCharacter, error) {
	chars, err := r.client.Character.Query().
		WithPlayer().
		Where(
			character.And(
				character.HasPlayerWith(player.Steamid(sid)),
				character.Version(1),
				character.DeletedAtIsNil(),
			),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	return charsToDepChars(chars), nil
}

// CharacterGetBySteamidSlot returns the latest Character for the provided slot associated
// to a Player with the provided Steam ID
func (r *characterRepository) CharacterGetBySteamidSlot(ctx context.Context, sid string, slt int) (*ent.DeprecatedCharacter, error) {
	char, err := r.client.Character.Query().Where(
		character.And(
			character.HasPlayerWith(player.Steamid(sid)),
			character.Slot(slt),
			character.Version(1),
			character.DeletedAtIsNil(),
		),
	).Only(ctx)
	if err != nil {
		return nil, err
	}

	return charToDepChar(sid, char), nil
}

// CharacterGetByID returns the latest Character with the provided ID
func (r *characterRepository) CharacterGetByID(ctx context.Context, id uuid.UUID) (*ent.DeprecatedCharacter, error) {
	char, err := r.client.Character.Query().
		WithPlayer().
		Where(
			character.And(
				character.ID(id),
				character.DeletedAtIsNil(),
			),
		).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	return charToDepChar(char.Edges.Player.Steamid, char), nil
}

// CharacterCreate creates and returns a Character for the provided slot and associated
// to the provided Player Steam ID
//
// This function makes the assumption that the newChar parameter has a valid Steam ID
// and slot number.
//
// If a Player resource doesn't exist, it is created
//
// If a Character already exists for the designated slot, it is deleted before saving the new Character
func (r *characterRepository) CharacterCreate(ctx context.Context, newChar ent.DeprecatedCharacter) (*ent.DeprecatedCharacter, error) {
	var char *ent.Character
	err := txn(ctx, r.client, func(tx *ent.Tx) error {
		player, err := r.client.Player.Query().
			Where(
				player.Steamid(newChar.Steamid),
			).
			Only(ctx)
		if err != nil {
			if !ent.IsNotFound(err) {
				return err
			}

			// Create Player if one doesn't exist
			player, err = r.client.Player.Create().
				SetSteamid(newChar.Steamid).
				Save(ctx)
			if err != nil {
				return err
			}
		}

		// Hard delete characters taking the requested slot
		found, err := r.client.Character.Query().
			Where(
				character.And(
					character.PlayerID(player.ID),
					character.Slot(newChar.Slot),
				),
			).
			Exist(ctx)
		if err != nil {
			return err
		}

		if found {
			_, err = r.client.Character.Delete().
				Where(
					character.And(
						character.PlayerID(player.ID),
						character.Slot(newChar.Slot),
					),
				).
				Exec(ctx)
			if err != nil {
				return err
			}
		}

		// Create new character
		c, err := r.client.Character.Create().
			SetPlayer(player).
			SetSlot(newChar.Slot).
			SetSize(newChar.Size).
			SetData(newChar.Data).
			SetVersion(1).
			Save(ctx)
		if err != nil {
			return err
		}

		char = c
		return nil
	})
	if err != nil {
		return nil, err
	}

	return charToDepChar(newChar.Steamid, char), nil
}

// CharacterUpdate updates and returns a Character using the provided data within the updateChar parameter
//
// The current Character version will be stored as a new backup before updating. If this backup goes over the
// maximum number of backups for this slot, the oldest backup version will be deleted.
func (r *characterRepository) CharacterUpdate(ctx context.Context, uid uuid.UUID, updateChar ent.DeprecatedCharacter) (*ent.DeprecatedCharacter, error) {
	var char *ent.Character
	err := txn(ctx, r.client, func(tx *ent.Tx) error {
		// Get the current character
		current, err := r.client.Character.Get(ctx, uid)
		if err != nil {
			return err
		}

		// Get the latest backup version
		latest, err := r.client.Character.Query().
			Select(character.FieldVersion).
			Where(character.Slot(current.Slot)).
			Order(ent.Desc(character.FieldVersion)).
			First(ctx)
		if err != nil {
			return err
		}

		// Backup the current version
		_, err = r.client.Character.Create().
			SetPlayerID(current.PlayerID).
			SetVersion(latest.Version + 1).
			SetSlot(current.Slot).
			SetSize(current.Size).
			SetData(current.Data).
			Save(ctx)
		if err != nil {
			return err
		}

		// Update the character
		c, err := r.client.Character.UpdateOneID(uid).
			SetSize(updateChar.Size).
			SetData(updateChar.Data).
			Save(ctx)
		if err != nil {
			return err
		}

		// Get all backup characters
		all, err := r.client.Character.Query().
			Where(
				character.And(
					character.PlayerID(c.PlayerID),
					character.Slot(c.Slot),
					character.VersionNEQ(1),
				),
			).
			Order(ent.Desc(character.FieldCreatedAt)).
			All(ctx)
		if err != nil {
			return err
		}

		// Delete all characters beyond 10 backups (version "1" not in current slice)
		if len(all) > 9 {
			for _, old := range all[9:] {
				if err := r.client.Character.DeleteOneID(old.ID).Exec(ctx); err != nil {
					return err
				}
			}
		}

		char = c
		return nil
	})
	if err != nil {
		return nil, err
	}

	return charToDepChar(updateChar.Steamid, char), nil
}

// CharacterDelete deletes a character
//
// The latest Character version will be soft deleted in an effort to preserve the right to
// restore the Character at a later time, i.e. A Player deletes a Character by accident and wants
// it restored. Backups will be hard deleted to free up space, since only the latest version is required.
func (r *characterRepository) CharacterDelete(ctx context.Context, uid uuid.UUID) error {
	return txn(ctx, r.client, func(tx *ent.Tx) error {
		// Get Current version
		char, err := r.client.Character.Get(ctx, uid)
		if err != nil {
			return err
		}

		// Soft delete
		_, err = char.Update().
			SetDeletedAt(time.Now()).
			Save(ctx)
		if err != nil {
			return err
		}

		// Hard delete backups
		_, err = r.client.Character.Delete().
			Where(
				character.And(
					character.PlayerID(char.PlayerID),
					character.Slot(char.Slot),
					character.VersionNEQ(char.Version),
				),
			).
			Exec(ctx)
		if err != nil {
			return err
		}

		return nil
	})
}

// CharacterRestore removes the deleted_at timestamp from the Character resource
func (r *characterRepository) CharacterRestore(ctx context.Context, uid uuid.UUID) (*ent.DeprecatedCharacter, error) {
	char, err := r.client.Character.UpdateOneID(uid).
		ClearDeletedAt().
		Save(ctx)
	if err != nil {
		return nil, err
	}

	player, err := char.QueryPlayer().Only(ctx)
	if err != nil {
		return nil, err
	}

	return charToDepChar(player.Steamid, char), nil
}

// CharacterVersions returns the latest Character version, and all of its backups,
// ordered by the updated_at timestamp in descending order (current version -> oldest version)
func (r *characterRepository) CharacterVersions(ctx context.Context, sid string, slot int) ([]*ent.Character, error) {
	chars, err := r.client.Character.Query().
		Where(
			character.And(
				character.HasPlayerWith(player.Steamid(sid)),
				character.Slot(slot),
			),
		).
		Order(ent.Desc(character.FieldUpdatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	return chars, nil
}

// CharacterRollback reverts the Character to a previous version
//
// Reverting to a previous version will first save the current version as a backup in an effort
// to keep a clean and "fast-forward only" history of the Character. There is a side affect from this.
// If a Character has 10 versions and the Player decides to "rollback" between two versions 10 times,
// the Character will end up having 2 versions saved 5 times each.
//
// Here is an illustration demonstrating this side effect
//
//	Character "abc": v1, v2, v3, v4, v5, v6, v7, v8, v9, v10
//	1. Rollback to "v8"
//		Previous: v1, v2, v3, v4, v5, v6, v7, v8, v9, v10
//		Current:  v8, v3, v4, v5, v6, v7, v8, v9, v10, v1
//	2. Rollback to "v1"
//		Previous: v8, v3, v4, v5, v6, v7, v8, v9, v10, v1
//		Current:  v1, v4, v5, v6, v7, v8, v9, v10, v1, v8
//	3. ... continues 8 more times
//		Current:  v1, v8, v1, v8, v1, v8, v1, v8, v1, v8
//
// Trying to revert to the current version (1) will NOT save an additional backup.
func (r *characterRepository) CharacterRollback(ctx context.Context, sid string, slot, version int) (*ent.DeprecatedCharacter, error) {
	var char *ent.Character
	err := txn(ctx, r.client, func(tx *ent.Tx) error {
		// Get the targeted character version
		targeted, err := r.client.Character.Query().
			Where(
				character.And(
					character.HasPlayerWith(player.Steamid(sid)),
					character.Slot(slot),
					character.Version(version),
				),
			).
			Only(ctx)
		if err != nil {
			return err
		}

		// No work needed if targeted version is current version
		if version == 1 {
			char = targeted
			return nil
		}

		// Get the current character
		current, err := r.client.Character.Query().
			Where(
				character.And(
					character.HasPlayerWith(player.Steamid(sid)),
					character.Slot(slot),
					character.Version(1),
				),
			).
			Only(ctx)
		if err != nil {
			return err
		}

		// Get the latest backup version
		latest, err := r.client.Character.Query().
			Select(character.FieldVersion).
			Where(
				character.And(
					character.HasPlayerWith(player.Steamid(sid)),
					character.Slot(slot),
				),
			).
			Order(ent.Desc(character.FieldVersion)).
			First(ctx)
		if err != nil {
			return err
		}

		// Backup the current version
		_, err = r.client.Character.Create().
			SetPlayerID(current.PlayerID).
			SetVersion(latest.Version + 1).
			SetSlot(current.Slot).
			SetSize(current.Size).
			SetData(current.Data).
			Save(ctx)
		if err != nil {
			return err
		}

		// Update the character
		c, err := current.Update().
			SetSize(targeted.Size).
			SetData(targeted.Data).
			Save(ctx)
		if err != nil {
			return err
		}

		char = c
		return nil
	})
	if err != nil {
		return nil, err
	}

	return charToDepChar(sid, char), nil
}

func charToDepChar(s string, c *ent.Character) *ent.DeprecatedCharacter {
	return &ent.DeprecatedCharacter{
		ID:      c.ID,
		Steamid: s,
		Slot:    c.Slot,
		Size:    c.Size,
		Data:    c.Data,
	}
}

func charsToDepChars(c []*ent.Character) []*ent.DeprecatedCharacter {
	deps := make([]*ent.DeprecatedCharacter, len(c))
	for i := range c {
		deps[i] = charToDepChar(c[i].Edges.Player.Steamid, c[i])
	}
	return deps
}
