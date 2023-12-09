package service

import (
	"errors"

	"github.com/msrevive/nexus2/internal/payload"
	"github.com/msrevive/nexus2/internal/database/schema"

	"github.com/google/uuid"
)

func (s *Service) NewCharacter(char payload.Character) (uuid.UUID, error) {
	uid, err := s.db.NewCharacter(char.SteamID, char.Slot, char.Size, char.Data); 
	if err != nil {
		return uuid.Nil, err
	}

	return uid, nil
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

	if len(char.Versions) == 0 {
		return nil, errors.New("malformed character data")
	}

	return char, nil
}

func (s *Service) GetCharacter(steamid string, slot int) (*schema.Character, error) {
	uid, err := s.db.LookUpCharacterID(steamid, slot)
	if err != nil {
		return nil, err
	}

	char, err := s.db.GetCharacter(uid); 
	if err != nil {
		return nil, err
	}

	if len(char.Versions) == 0 {
		return nil, errors.New("malformed character data")
	}

	return char, nil
}

func (s *Service) GetCharacters(steamid string) ([]schema.Character, error) {
	chars, err := s.db.GetCharacters(steamid)
	if err != nil {
		return nil, err
	}

	return chars, nil
}

func (s *Service) GetDeletedCharacters(steamid string) (map[int]schema.DeletedCharacter, error) {
	user, err := s.db.GetUser(steamid)
	if err != nil {
		return nil, err
	}

	return user.DeletedCharacters, nil
}

func (s *Service) SoftDeleteCharacter(uid uuid.UUID) error {
	if _,err := s.db.SoftDeleteCharacter(uid); err != nil {
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