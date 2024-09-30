package database

import (
	"time"
	"log"

	"github.com/msrevive/nexus2/internal/bitmask"
	"github.com/msrevive/nexus2/pkg/database/schema"

	"github.com/google/uuid"
)

type Options struct {
	Logger *log.Logger
}

type Database interface {
	Connect(cfg Config, opts Options) error
	Disconnect() error

	GetAllUsers() ([]*schema.User, error)
	GetUser(steamid string) (*schema.User, error)
	SetUserFlags(steamid string, flags bitmask.Bitmask) (error)
	GetUserFlags(steamid string) (bitmask.Bitmask, error)

	NewCharacter(steamid string, slot int, size int, data string) (uuid.UUID, error)
	UpdateCharacter(id uuid.UUID, size int, data string, backupMax int, backupTime time.Duration) error
	GetCharacter(id uuid.UUID) (*schema.Character, error)
	GetCharacters(steamid string) (map[int]schema.Character, error) //Gotta be a map cause JSON
	LookUpCharacterID(steamid string, slot int) (uuid.UUID, error)
	SoftDeleteCharacter(id uuid.UUID, expiration time.Duration) error
	DeleteCharacter(id uuid.UUID) error
	DeleteCharacterReference(steamid string, slot int) error
	MoveCharacter(id uuid.UUID, steamid string, slot int) error
	CopyCharacter(id uuid.UUID, steamid string, slot int) (uuid.UUID, error)
	RestoreCharacter(id uuid.UUID) error
	RollbackCharacter(id uuid.UUID, ver int) error
	RollbackCharacterToLatest(id uuid.UUID) error
	DeleteCharacterVersions(id uuid.UUID) error

	SaveToDatabase() error
	ClearCache()
}