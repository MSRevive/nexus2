package database

import (
	"github.com/msrevive/nexus2/internal/database/schema"

	"github.com/google/uuid"
)

type Database interface {
	Connect(conn string) error
	Disconnect() error

	NewCharacter(steamid string, slot int, size int, data string) (uuid.UUID, error)
	UpdateCharacter(id uuid.UUID, size int, data string, backupMax int, backupTime string) error
	GetUser(steamid string) (*schema.User, error)
	GetCharacter(id uuid.UUID) (*schema.Character, error)
	GetCharacters(steamid string) ([]schema.Character, error)
	LookUpCharacterID(steamid string, slot int) (uuid.UUID, error)
	SoftDeleteCharacter(id uuid.UUID) (uuid.UUID, error)
}