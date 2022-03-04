package service

import (
  //"entgo.io/ent/dialect/sql"
  "github.com/google/uuid"
  
  "github.com/msrevive/nexus2/ent"
  "github.com/msrevive/nexus2/ent/character"
)

func (s *service) CharactersGetAll() ([]*ent.Character, error) {
  chars, err := s.client.Character.Query().All(s.ctx)
  if err != nil {
    return nil, err
  }
  
  return chars, nil
}

func (s *service) CharactersGetBySteamid(sid string) ([]*ent.Character, error) {
  chars, err := s.client.Character.Query().Where(
    character.Steamid(sid),
  ).All(s.ctx)
  if err != nil {
    return nil, err
  }
  
  return chars, nil
}

func (s *service) CharacterGetBySteamidSlot(sid string, slt int) ([]*ent.Character, error) {
  char, err := s.client.Character.Query().Where(
    character.And(
      character.Steamid(sid),
      character.Slot(slt),
    ),
  ).All(s.ctx)
  if err != nil {
    return nil, err
  }
  
  return char, nil
}

func (s *service) CharacterGetByID(id uuid.UUID) (*ent.Character, error) {
  char, err := s.client.Character.Get(s.ctx, id)
  if err != nil {
    return nil, err
  }
  
  return char, nil
}

func (s *service) CharacterCreate(newChar ent.Character) (*ent.Character, error) {
  char, err := s.client.Character.Create().
  SetSteamid(newChar.Steamid).
  SetSlot(newChar.Slot).
  SetSize(newChar.Size).
  SetData(newChar.Data).
  Save(s.ctx)
  if err != nil {
    return nil, err
  }
  
  return char, nil
}

func (s *service) CharacterUpdate(uid uuid.UUID, updateChar ent.Character) (*ent.Character, error) {
  char, err := s.client.Character.UpdateOneID(uid).
  SetSize(updateChar.Size).
  SetData(updateChar.Data).
  Save(s.ctx)
  if err != nil {
    return nil, err
  }
  
  return char, nil
}

func (s *service) CharacterDelete(uid uuid.UUID) (error) {
  err := s.client.Character.DeleteOneID(uid).Exec(s.ctx)
  if err != nil {
    return err
  }
  
  return nil
}