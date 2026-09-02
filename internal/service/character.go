package service

import (
	"context"

	"github.com/msrevive/nexus2/internal/bitmask"
	"github.com/msrevive/nexus2/internal/payload"
	"github.com/msrevive/nexus2/internal/static"
	"github.com/msrevive/nexus2/pkg/database/schema"
	"github.com/msrevive/nexus2/pkg/utils"

	"github.com/google/uuid"
	"github.com/titpetric/oida"
)

func (s *Service) NewCharacter(ctx context.Context, char payload.Character) (uuid.UUID, bitmask.Bitmask, error) {
	ctx, span := oida.Start(ctx, "svc NewCharacter")
	defer span.End()
	span.SetAttribute("steamid", char.SteamID)
	span.SetAttribute("slot", char.Slot)

	uid, err := s.db.NewCharacter(ctx, char.SteamID, char.Slot, char.Size, char.Data);
	if err != nil {
		span.RecordError(err)
		return uuid.Nil, 0, err
	}

	flags, err := s.db.GetUserFlags(ctx, char.SteamID)
	if err != nil {
		span.RecordError(err)
		return uuid.Nil, 0, err
	}

	return uid, flags, nil
}

func (s *Service) UpdateCharacter(ctx context.Context, uuid uuid.UUID, char payload.Character) error {
	if s.readonly {
		return nil
	}

	if err := s.db.UpdateCharacter(ctx, uuid, char.Size, char.Data, s.config.Char.MaxBackups, s.config.Char.BackupTime); err != nil {
		return err
	}

	return nil
}

func (s *Service) GetCharacterByID(ctx context.Context, uuid uuid.UUID) (*schema.Character, error) {
	char, err := s.db.GetCharacter(ctx, uuid);
	if err != nil {
		return nil, err
	}

	if (schema.CharacterData{}) == char.Data {
		return nil, static.ErrBadCharacterData
	}

	return char, nil
}

func (s *Service) GetCharacter(ctx context.Context, steamid string, slot int) (*schema.Character, bitmask.Bitmask, error) {
	ctx, span := oida.Start(ctx, "svc GetCharacter")
	defer span.End()
	span.SetAttribute("steamid", steamid)
	span.SetAttribute("slot", slot)

	user, err := s.db.GetUser(ctx, steamid)
	if err != nil {
		span.RecordError(err)
		return nil, 0, err
	}

	charID, _ := user.Characters[slot]

	char, err := s.GetCharacterByID(ctx, charID)
	if err != nil {
		span.RecordError(err)
		return nil, 0, err
	}

	return char, bitmask.Bitmask(user.Flags), err
}

func (s *Service) GetCharacters(ctx context.Context, steamid string) (map[int]schema.Character, bitmask.Bitmask, error) {
	ctx, span := oida.Start(ctx, "svc GetCharacters")
	defer span.End()
	span.SetAttribute("steamid", steamid)

	chars, err := s.db.GetCharacters(ctx, steamid)
	if err != nil {
		span.RecordError(err)
		return nil, 0, err
	}

	flags, err := s.db.GetUserFlags(ctx, steamid)
	if err != nil {
		span.RecordError(err)
		return nil, 0, err
	}

	return chars, flags, nil
}

func (s *Service) GetDeletedCharacters(ctx context.Context, steamid string) (map[int]uuid.UUID, error) {
	user, err := s.db.GetUser(ctx, steamid)
	if err != nil {
		return nil, err
	}

	return user.DeletedCharacters, nil
}

func (s *Service) SoftDeleteCharacter(ctx context.Context, uid uuid.UUID, expiration string) error {
	expire, err := utils.ParseDuration(expiration)
	if err != nil {
		return err
	}

	if err := s.db.SoftDeleteCharacter(ctx, uid, expire); err != nil {
		return err
	}

	return nil
}

func (s *Service) LookUpCharacterID(ctx context.Context, steamid string, slot int) (uuid.UUID, error) {
	uid, err := s.db.LookUpCharacterID(ctx, steamid, slot)
	if err != nil {
		return uuid.Nil, err
	}

	return uid, nil
}

func (s *Service) MoveCharacter(ctx context.Context, uid uuid.UUID, steamid string, slot int) (uuid.UUID, error) {
	if s.readonly {
		return uuid.Nil, nil
	}

	if err := s.db.MoveCharacter(ctx, uid, steamid, slot); err != nil {
		return uuid.Nil, err
	}

	return uid, nil
}

func (s *Service) CopyCharacter(ctx context.Context, uid uuid.UUID, steamid string, slot int) (uuid.UUID, error) {
	if s.readonly {
		return uuid.Nil, nil
	}

	newUID, err := s.db.CopyCharacter(ctx, uid, steamid, slot);
	if err != nil {
		return uuid.Nil, err
	}

	return newUID, nil
}

func (s *Service) HardDeleteCharacter(ctx context.Context, uid uuid.UUID) error {
	if s.readonly {
		return nil
	}

	ctx, span := oida.Start(ctx, "svc HardDeleteCharacter")
	defer span.End()
	span.SetAttribute("uuid", uid.String())

	// make sure character exists
	_, err := s.db.GetCharacter(ctx, uid);
	if err != nil {
		span.RecordError(err)
		return err
	}

	// if err := s.db.DeleteCharacterReference(ctx, char.SteamID, char.Slot); err != nil {
	// 	return err
	// }

	if err := s.db.DeleteCharacter(ctx, uid); err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}

func (s *Service) RestoreCharacter(ctx context.Context, uid uuid.UUID) error {
	if err := s.db.RestoreCharacter(ctx, uid); err != nil {
		return err
	}

	return nil
}
