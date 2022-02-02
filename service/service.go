package service

import (
  "context"
  
  "github.com/msrevive/nexus2/system"
  "github.com/msrevive/nexus2/ent"
)

type service struct {
  ctx context.Context
  client *ent.Client
}

func New(ctx context.Context) *service {
  return &service{
    ctx: ctx,
    client: system.Client,
  }
}

func (s *service) Debug() error {
  _, err := s.client.Character.Create().
  SetSteamid("76561198092541763").
  SetSlot(1).
  SetName("Test Char").
  SetGender(1).
  SetRace("hooman").
  SetFlags("{}").
  SetQuickslots("{}").
  SetQuests("{}").
  SetGuild("").
  SetKills(0).
  SetGold(0).
  SetSkills("{}").
  SetPets("{}").
  SetHealth(5).
  SetMana(10).
  SetEquipped("{}").
  SetLefthand("").
  SetRighthand("").
  SetSpells("{}").
  SetSpellbook("{}").
  SetBags("{}").
  Save(s.ctx)
  if err != nil {
    return err
  }
  
  return nil
}