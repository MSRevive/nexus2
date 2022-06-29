package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/msrevive/nexus2/ent"
	"github.com/msrevive/nexus2/ent/character"
	"github.com/msrevive/nexus2/ent/player"
)

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

func (s *service) CharacterUpdate(uid uuid.UUID, updateChar ent.DeprecatedCharacter) (*ent.DeprecatedCharacter, error) {
	var char *ent.Character
	err := txn(s.ctx, s.client, func(tx *ent.Tx) error {
		// Get the current character
		current, err := s.client.Character.Get(s.ctx, uid)
		if err != nil {
			return err
		}

		// Get the latest backup version
		latest, err := s.client.Character.Query().
			Select(character.FieldVersion).
			Where(character.ID(uid)).
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

func (s *service) CharacterVersions(sid string, slot int) ([]*ent.Character, error) {
	chars, err := s.client.Character.Query().
		Where(
			character.And(
				character.HasPlayerWith(player.Steamid(sid)),
				character.Slot(slot),
			),
		).
		All(s.ctx)
	if err != nil {
		return nil, err
	}

	return chars, nil
}

func (s *service) CharacterRollback(sid string, slot, version int) (*ent.DeprecatedCharacter, error) {
	var char *ent.Character
	err := txn(s.ctx, s.client, func(tx *ent.Tx) error {
		// Get the current character
		targeted, err := s.client.Character.Query().
			Where(
				character.And(
					character.HasPlayerWith(player.Steamid(sid)),
					character.Slot(slot),
					character.Version(version),
				),
			).
			First(s.ctx)
		if err != nil {
			return err
		}

		// Get the current character
		current, err := s.client.Character.Query().
			Where(
				character.And(
					character.HasPlayerWith(player.Steamid(sid)),
					character.Version(1),
				),
			).
			First(s.ctx)
		if err != nil {
			return err
		}

		// Get the latest backup version
		latest, err := s.client.Character.Query().
			Select(character.FieldVersion).
			Where(character.HasPlayerWith(player.Steamid(sid))).
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
