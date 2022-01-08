package service

import (
  "context"
  
  "github.com/msrevive/nexus2/session"
  "github.com/msrevive/nexus2/ent"
)

type service struct {
  ctx context.Context
  client *ent.Client
}

func New(ctx context.Context) *service {
  return &service{
    ctx: ctx,
    client: session.Client,
  }
}