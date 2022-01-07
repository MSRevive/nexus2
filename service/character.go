package service

import (
  "github.com/msrevive/nexus2/ent"
)

func (s *service) CharGetAll() ([]*ent.Character, error) {
  chars, err := s.client.Character.Query().All(s.ctx)
  if err != nil {
    return nil, err
  }
  
  return chars, nil
}