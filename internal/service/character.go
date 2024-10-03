package service

import (
	"fmt"
	"time"

	"github.com/msrevive/nexus2/internal/bitmask"
	"github.com/msrevive/nexus2/internal/payload"
	"github.com/msrevive/nexus2/pkg/database/schema"

	"github.com/google/uuid"
)

func (s *Service) NewCharacter(char payload.Character) (uuid.UUID, bitmask.Bitmask, error) {
	uid, err := s.db.NewCharacter(char.SteamID, char.Slot, char.Size, char.Data); 
	if err != nil {
		return uuid.Nil, 0, err
	}

	flags, err := s.db.GetUserFlags(char.SteamID)
	if err != nil {
		return uuid.Nil, 0, err
	}

	return uid, flags, nil
}

func (s *Service) UpdateCharacter(uuid uuid.UUID, char payload.Character) error {
	if err := s.db.UpdateCharacter(uuid, char.Size, char.Data, s.config.Char.MaxBackups, s.config.Char.BackupTime); err != nil {
		return err
	}

	return nil
}

func (s *Service) GetCharacterByID(uuid uuid.UUID) (*schema.Character, error) {
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

func (s *Service) GetDeletedCharacters(steamid string) (map[int]uuid.UUID, error) {
	user, err := s.db.GetUser(steamid)
	if err != nil {
		return nil, err
	}

	return user.DeletedCharacters, nil
}

func (s *Service) SoftDeleteCharacter(uid uuid.UUID, expiration time.Duration) error {
	if err := s.db.SoftDeleteCharacter(uid, expiration); err != nil {
		return err
	}

	return nil
}

func (s *Service) LookUpCharacterID(steamid string, slot int) (uuid.UUID, error) {
	uid, err := s.db.LookUpCharacterID(steamid, slot)
	if err != nil {
		return uuid.Nil, err
	}

	return uid, nil
}

func (s *Service) MoveCharacter(uid uuid.UUID, steamid string, slot int) (uuid.UUID, error) {
	if err := s.db.MoveCharacter(uid, steamid, slot); err != nil {
		return uuid.Nil, err
	}

	return uid, nil
}

func (s *Service) CopyCharacter(uid uuid.UUID, steamid string, slot int) (uuid.UUID, error) {
	newUID, err := s.db.CopyCharacter(uid, steamid, slot); 
	if err != nil {
		return uuid.Nil, err
	}

	return newUID, nil
}

func (s *Service) HardDeleteCharacter(uid uuid.UUID) error {
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

func (s *Service) RestoreCharacter(uid uuid.UUID) error {
	if err := s.db.RestoreCharacter(uid); err != nil {
		return err
	}

	return nil
}