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
  SetName(newChar.Name).
  SetGender(newChar.Gender).
  SetFlags(newChar.Flags).
  SetQuickslots(newChar.Quickslots).
  SetQuests(newChar.Quests).
  SetGuild(newChar.Guild).
  SetKills(newChar.Kills).
  SetGold(newChar.Gold).
  SetSkills(newChar.Skills).
  SetPets(newChar.Pets).
  SetHealth(newChar.Health).
  SetMana(newChar.Mana).
  SetEquipped(newChar.Equipped).
  SetLefthand(newChar.Lefthand).
  SetRighthand(newChar.Righthand).
  SetSpells(newChar.Spells).
  SetSpellbook(newChar.Spellbook).
  SetBags(newChar.Bags).
  Save(s.ctx)
  if err != nil {
    return nil, err
  }
  
  return char, nil
}

func (s *service) CharacterUpdate(uid uuid.UUID, updateChar ent.Character) (*ent.Character, error) {
  // char, err := s.client.Character.UpdateOneID(uid).
  // SetSteamid(updateChar.Steamid).
  // SetSlot(updateChar.Slot).
  // SetName(updateChar.Name).
  // SetGender(updateChar.Gender).
  // SetRace(updateChar.Race).
  // SetNillableFlags(updateChar.Flags).
  // SetNillableQuickslots(updateChar.Quickslots).
  // SetNillableQuests(updateChar.Quests).
  // SetGuild(updateChar.Guild).
  // SetKills(updateChar.Kills).
  // SetGold(updateChar.Gold).
  // SetNillableSkills(updateChar.Skills).
  // SetNillablePets(updateChar.Pets).
  // SetHealth(updateChar.Health).
  // SetMana(updateChar.Mana).
  // SetNillableEquipped(updateChar.Equipped).
  // SetLefthand(updateChar.Lefthand).
  // SetRighthand(updateChar.Righthand).
  // SetNillableSpells(updateChar.Spells).
  // SetNillableSpellbook(updateChar.Spellbook).
  // SetNillableBags(updateChar.Bags).
  // SetNillableSheaths(updateChar.Sheaths).
  // UpdateNewValues().Save(s.ctx)
  char, err := s.client.Character.UpdateOneID(uid).
  SetSteamid(updateChar.Steamid).
  SetSlot(updateChar.Slot).
  SetName(updateChar.Name).
  SetGender(updateChar.Gender).
  SetFlags(updateChar.Flags).
  SetQuickslots(updateChar.Quickslots).
  SetQuests(updateChar.Quests).
  SetGuild(updateChar.Guild).
  SetKills(updateChar.Kills).
  SetGold(updateChar.Gold).
  SetSkills(updateChar.Skills).
  SetPets(updateChar.Pets).
  SetHealth(updateChar.Health).
  SetMana(updateChar.Mana).
  SetEquipped(updateChar.Equipped).
  SetLefthand(updateChar.Lefthand).
  SetRighthand(updateChar.Righthand).
  SetSpells(updateChar.Spells).
  SetSpellbook(updateChar.Spellbook).
  SetBags(updateChar.Bags).
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