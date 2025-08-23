package service

import (
	"fmt"

	"github.com/msrevive/nexus2/internal/bitmask"
	"github.com/msrevive/nexus2/internal/payload"
	"github.com/msrevive/nexus2/pkg/database/schema"
	"github.com/msrevive/nexus2/pkg/utils"

	"github.com/bwmarrin/snowflake"
)

func (s *Service) NewCharacter(char payload.Character) (snowflake.ID, bitmask.Bitmask, error) {
	uid, err := s.db.NewCharacter(char.SteamID, char.Slot, char.Size, char.Data); 
	if err != nil {
		return 0, 0, err
	}

	flags, err := s.db.GetUserFlags(char.SteamID)
	if err != nil {
		return 0, 0, err
	}

	return uid, flags, nil
}

func (s *Service) UpdateCharacter(uuid snowflake.ID, char payload.Character) error {
	if err := s.db.UpdateCharacter(uuid, char.Size, char.Data, s.config.Char.MaxBackups, s.config.Char.BackupTime); err != nil {
		return err
	}

	return nil
}

func (s *Service) GetCharacterByID(uuid snowflake.ID) (*schema.Character, error) {
	char, err := s.db.GetCharacter(uuid); 
	if err != nil {
		return nil, err
	}

	if (schema.CharacterData{}) == char.Data {
		return nil, fmt.Errorf("malformed character data")
	}

	return char, nil
}

func (s *Service) GetCharacter(steamid string, slot int) (*schema.Character, bitmask.Bitmask, error) {
	user, err := s.db.GetUser(steamid)
	if err != nil {
		return nil, 0, err
	}

	charID, _ := user.Characters[slot]

	char, err := s.GetCharacterByID(charID)
	if err != nil {
		return nil, 0, err
	}

	return char, bitmask.Bitmask(user.Flags), err
}

func (s *Service) GetCharacters(steamid string) (map[int]schema.Character, bitmask.Bitmask, error) {
	chars, err := s.db.GetCharacters(steamid)
	if err != nil {
		return nil, 0, err
	}

	flags, err := s.db.GetUserFlags(steamid)
	if err != nil {
		return nil, 0, err
	}

	return chars, flags, nil
}

func (s *Service) GetDeletedCharacters(steamid string) (map[int]snowflake.ID, error) {
	user, err := s.db.GetUser(steamid)
	if err != nil {
		return nil, err
	}

	return user.DeletedCharacters, nil
}

func (s *Service) SoftDeleteCharacter(uid snowflake.ID, expiration string) error {
	expire, err := utils.ParseDuration(expiration)
	if err != nil {
		return err
	}

	if err := s.db.SoftDeleteCharacter(uid, expire); err != nil {
		return err
	}

	return nil
}

func (s *Service) LookUpCharacterID(steamid string, slot int) (snowflake.ID, error) {
	uid, err := s.db.LookUpCharacterID(steamid, slot)
	if err != nil {
		return 0, err
	}

	return uid, nil
}

func (s *Service) MoveCharacter(uid snowflake.ID, steamid string, slot int) (snowflake.ID, error) {
	if err := s.db.MoveCharacter(uid, steamid, slot); err != nil {
		return 0, err
	}

	return uid, nil
}

func (s *Service) CopyCharacter(uid snowflake.ID, steamid string, slot int) (snowflake.ID, error) {
	newUID, err := s.db.CopyCharacter(uid, steamid, slot); 
	if err != nil {
		return 0, err
	}

	return newUID, nil
}

func (s *Service) HardDeleteCharacter(uid snowflake.ID) error {
	char, err := s.db.GetCharacter(uid); 
	if err != nil {
		return err
	}

	if err := s.db.DeleteCharacterReference(char.SteamID, char.Slot); err != nil {
		return err
	}

	if err := s.db.DeleteCharacter(uid); err != nil {
		return err
	}

	return nil
}

func (s *Service) RestoreCharacter(uid snowflake.ID) error {
	if err := s.db.RestoreCharacter(uid); err != nil {
		return err
	}

	return nil
}