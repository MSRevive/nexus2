package service

import (
	"time"

	"github.com/msrevive/nexus2/ent"
	"github.com/msrevive/nexus2/ent/character"
	"github.com/msrevive/nexus2/ent/player"

	"github.com/google/uuid"
)

// CharactersGetAll returns all latest, active Characters
func (s *service) CharactersGetAll() ([]*ent.DeprecatedCharacter, error) {
	chars, err := s.client.Character.Query().
		WithPlayer().
		Where(
			character.And(
				character.Version(1),
				character.DeletedAtIsNil(),
			),
		).
		All(s.ctx)
	if err != nil {
		return nil, err
	}

	return charsToDepChars(chars), nil
}

// CharactersGetBySteamid returns all Characters associated to a Player with the provided Steam ID
func (s *service) CharactersGetBySteamid(sid string) ([]*ent.DeprecatedCharacter, error) {
	chars, err := s.client.Character.Query().
		WithPlayer().
		Where(
			character.And(
				character.HasPlayerWith(player.Steamid(sid)),
				character.Version(1),
				character.DeletedAtIsNil(),
			),
		).
		All(s.ctx)
	if err != nil {
		return nil, err
	}

	return charsToDepChars(chars), nil
}

// CharacterGetBySteamidSlot returns the latest Character for the provided slot associated
// to a Player with the provided Steam ID
func (s *service) CharacterGetBySteamidSlot(sid string, slt int) (*ent.DeprecatedCharacter, error) {
	char, err := s.client.Character.Query().Where(
		character.And(
			character.HasPlayerWith(player.Steamid(sid)),
			character.Slot(slt),
			character.Version(1),
			character.DeletedAtIsNil(),
		),
	).Only(s.ctx)
	if err != nil {
		return nil, err
	}

	return charToDepChar(sid, char), nil
}

// CharacterGetByID returns the latest Character with the provided ID
func (s *service) CharacterGetByID(id uuid.UUID) (*ent.DeprecatedCharacter, error) {
	char, err := s.client.Character.Query().
		WithPlayer().
		Where(
			character.And(
				character.ID(id),
				character.DeletedAtIsNil(),
			),
		).
		Only(s.ctx)
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
func (s *service) CharacterCreate(newChar ent.DeprecatedCharacter) (*ent.DeprecatedCharacter, error) {
	var char *ent.Character
	err := txn(s.ctx, s.client, func(tx *ent.Tx) error {
		player, err := s.client.Player.Query().
			Where(
				player.Steamid(newChar.Steamid),
			).
			Only(s.ctx)
		if err != nil {
			if !ent.IsNotFound(err) {
				return err
			}

			// Create Player if one doesn't exist
			player, err = s.client.Player.Create().
				SetSteamid(newChar.Steamid).
				Save(s.ctx)
			if err != nil {
				return err
			}
		}

		// Hard delete characters taking the requested slot
		found, err := s.client.Character.Query().
			Where(
				character.And(
					character.PlayerID(player.ID),
					character.Slot(newChar.Slot),
				),
			).
			Exist(s.ctx)
		if err != nil {
			return err
		}

		if found {
			_, err = s.client.Character.Delete().
				Where(
					character.And(
						character.PlayerID(player.ID),
						character.Slot(newChar.Slot),
					),
				).
				Exec(s.ctx)
			if err != nil {
				return err
			}
		}

		// Create new character
		c, err := s.client.Character.Create().
			SetPlayer(player).
			SetSlot(newChar.Slot).
			SetSize(newChar.Size).
			SetData(newChar.Data).
			SetVersion(1).
			Save(s.ctx)
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
func (s *service) CharacterUpdate(uid uuid.UUID, updateChar ent.DeprecatedCharacter) (*ent.DeprecatedCharacter, error) {
	var char *ent.Character
	err := txn(s.ctx, s.client, func(tx *ent.Tx) error {
		// Get the current character
		current, err := s.client.Character.Get(s.ctx, uid)
		if err != nil {
			return err
		}

		// Get all backup characters
		all, err := s.client.Character.Query().
			Select(character.FieldUpdatedAt, character.FieldVersion, character.FieldID).
			Where(
				character.And(
					character.PlayerID(current.PlayerID),
					character.Slot(current.Slot),
					character.VersionNEQ(1),
				),
			).
			Order(ent.Desc(character.FieldVersion)).
			All(s.ctx)
		if err != nil {
			return err
		}

		// now we just use the original slice so we don't need to make uneeded queries
		latest := all[0]
		earliest := all[len(all)-1]

		// Backup the current version
		backupTime,err := time.ParseDuration(s.apps.Config.Char.BackupTime)
		if err != nil {
			return err
		}
		timeCheck := latest.UpdatedAt.Add(backupTime)
		if (current.UpdatedAt.After(timeCheck) || latest.Version == 0) {
			if len(all) > s.apps.Config.Char.MaxBackups-1 {
				if err := s.client.Character.DeleteOneID(earliest.ID).Exec(s.ctx); err != nil {
					return err
				}
			}

			_, err = s.client.Character.Create().
				SetPlayerID(current.PlayerID).
				SetVersion(latest.Version+1).
				SetSlot(current.Slot).
				SetSize(current.Size).
				SetData(current.Data).
				Save(s.ctx)
			if err != nil {
				return err
			}
		}
		
		// Clean up any excess character backups
		if len(all) > s.apps.Config.Char.MaxBackups {
			for _, old := range all[s.apps.Config.Char.MaxBackups:] {
				if err := s.client.Character.DeleteOneID(old.ID).Exec(s.ctx); err != nil {
					return err
				}
			}
		}

		// Update the character
		c, err := s.client.Character.UpdateOneID(uid).
			SetSize(updateChar.Size).
			SetData(updateChar.Data).
			Save(s.ctx)
		if err != nil {
			return err
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
func (s *service) CharacterDelete(uid uuid.UUID) error {
	return txn(s.ctx, s.client, func(tx *ent.Tx) error {
		// Get Current version
		char, err := s.client.Character.Get(s.ctx, uid)
		if err != nil {
			return err
		}

		// Soft delete
		_, err = char.Update().
			SetDeletedAt(time.Now()).
			Save(s.ctx)
		if err != nil {
			return err
		}

		// Hard delete backups
		_, err = s.client.Character.Delete().
			Where(
				character.And(
					character.PlayerID(char.PlayerID),
					character.Slot(char.Slot),
					character.VersionNEQ(char.Version),
				),
			).
			Exec(s.ctx)
		if err != nil {
			return err
		}

		return nil
	})
}

// CharacterRestore removes the deleted_at timestamp from the Character resource
func (s *service) CharacterRestore(uid uuid.UUID) (*ent.DeprecatedCharacter, error) {
	char, err := s.client.Character.UpdateOneID(uid).
		ClearDeletedAt().
		Save(s.ctx)
	if err != nil {
		return nil, err
	}

	player, err := char.QueryPlayer().Only(s.ctx)
	if err != nil {
		return nil, err
	}

	return charToDepChar(player.Steamid, char), nil
}

// Shortcut to restore a deleted character by SteamID and slot.
func (s *service) CharacterRestoreBySteamID(steamid string, slot int) (*ent.DeprecatedCharacter, error) {
	target, err := s.client.Character.Query().
		Select(character.FieldID).
		Where(
			character.And(
				character.HasPlayerWith(player.Steamid(steamid)),
				character.Slot(slot),
				character.Version(1),
			),
		).
		Only(s.ctx)
	if err != nil {
		return nil, err
	}

	return s.CharacterRestore(target.ID)
}

// CharacterVersions returns the latest Character version, and all of its backups,
// ordered by the updated_at timestamp in descending order (current version -> oldest version)
func (s *service) CharacterVersions(sid string, slot int) ([]*ent.Character, error) {
	chars, err := s.client.Character.Query().
		Where(
			character.And(
				character.HasPlayerWith(player.Steamid(sid)),
				character.Slot(slot),
			),
		).
		Order(ent.Asc(character.FieldVersion)).
		All(s.ctx)
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
// 	Character "abc": v1, v2, v3, v4, v5, v6, v7, v8, v9, v10
// 	1. Rollback to "v8"
// 		Previous: v1, v2, v3, v4, v5, v6, v7, v8, v9, v10
// 		Current:  v8, v3, v4, v5, v6, v7, v8, v9, v10, v1
// 	2. Rollback to "v1"
// 		Previous: v8, v3, v4, v5, v6, v7, v8, v9, v10, v1
// 		Current:  v1, v4, v5, v6, v7, v8, v9, v10, v1, v8
// 	3. ... continues 8 more times
// 		Current:  v1, v8, v1, v8, v1, v8, v1, v8, v1, v8
//
// Trying to revert to the current version (1) will NOT save an additional backup.
func (s *service) CharacterRollback(sid string, slot, version int) (*ent.DeprecatedCharacter, error) {
	var char *ent.Character
	err := txn(s.ctx, s.client, func(tx *ent.Tx) error {
		// Get the targeted character version
		targeted, err := s.client.Character.Query().
			Where(
				character.And(
					character.HasPlayerWith(player.Steamid(sid)),
					character.Slot(slot),
					character.Version(version),
				),
			).
			Only(s.ctx)
		if err != nil {
			return err
		}

		// No work needed if targeted version is current version
		if version == 1 {
			char = targeted
			return nil
		}

		// Get the current character
		current, err := s.client.Character.Query().
			Where(
				character.And(
					character.HasPlayerWith(player.Steamid(sid)),
					character.Slot(slot),
					character.Version(1),
				),
			).
			Only(s.ctx)
		if err != nil {
			return err
		}

		// Get the latest backup version
		latest, err := s.client.Character.Query().
			Select(character.FieldVersion).
			Where(
				character.And(
					character.HasPlayerWith(player.Steamid(sid)),
					character.Slot(slot),
				),
			).
			Order(ent.Desc(character.FieldVersion)).
			First(s.ctx)
		if err != nil {
			return err
		}

		// Backup the current version
		_, err = s.client.Character.Create().
			SetPlayerID(current.PlayerID).
			SetVersion(latest.Version + 1).
			SetSlot(current.Slot).
			SetSize(current.Size).
			SetData(current.Data).
			Save(s.ctx)
		if err != nil {
			return err
		}

		// Update the character
		c, err := current.Update().
			SetSize(targeted.Size).
			SetData(targeted.Data).
			Save(s.ctx)
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

func (s *service) CharacterRollbackLatest(sid string, slot int) (*ent.DeprecatedCharacter, error) {
	var char *ent.Character
	err := txn(s.ctx, s.client, func(tx *ent.Tx) error {
		// Get the current character
		current, err := s.client.Character.Query().
			Where(
				character.And(
					character.HasPlayerWith(player.Steamid(sid)),
					character.Slot(slot),
					character.Version(1),
				),
			).
			Only(s.ctx)
		if err != nil {
			return err
		}

		// Get the latest backup version
		latest, err := s.client.Character.Query().
			Where(
				character.And(
					character.HasPlayerWith(player.Steamid(sid)),
					character.Slot(slot),
				),
			).
			Order(ent.Desc(character.FieldVersion)).
			First(s.ctx)
		if err != nil {
			return err
		}

		// Backup the current version
		_, err = s.client.Character.Create().
			SetPlayerID(current.PlayerID).
			SetVersion(latest.Version + 1).
			SetSlot(current.Slot).
			SetSize(current.Size).
			SetData(current.Data).
			Save(s.ctx)
		if err != nil {
			return err
		}

		// Update the character
		c, err := current.Update().
			SetSize(latest.Size).
			SetData(latest.Data).
			Save(s.ctx)
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

func (s *service) CharacterDeleteRollbacks(sid string, slot int) error {
	return txn(s.ctx, s.client, func(tx *ent.Tx) error {
		char, err := s.client.Character.Query().
			Select(character.FieldPlayerID, character.FieldVersion).
			Where(
				character.And(
					character.HasPlayerWith(player.Steamid(sid)),
					character.Slot(slot),
					character.Version(1),
				),
			).
			Only(s.ctx)
		if err != nil {
			return err
		}

		// Hard delete backups
		_, err = s.client.Character.Delete().
			Where(
				character.And(
					character.PlayerID(char.PlayerID),
					character.Slot(slot),
					character.VersionNEQ(char.Version),
				),
			).
			Exec(s.ctx)
		if err != nil {
			return err
		}

		return nil
	})
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
