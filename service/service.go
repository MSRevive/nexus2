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
  SetData("data").
  Save(s.ctx)
  if err != nil {
    return err
  }
  
  return nil
}