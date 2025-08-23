package database

import (
	"time"
	"log"
	"errors"

	"github.com/msrevive/nexus2/internal/bitmask"
	"github.com/msrevive/nexus2/pkg/database/schema"

	"github.com/bwmarrin/snowflake"
)

var (
	ErrNoDocument = errors.New("no document")
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

	NewCharacter(steamid string, slot int, size int, data string) (snowflake.ID, error)
	UpdateCharacter(id snowflake.ID, size int, data string, backupMax int, backupTime time.Duration) error
	GetCharacter(id snowflake.ID) (*schema.Character, error)
	GetCharacters(steamid string) (map[int]schema.Character, error) //Gotta be a map cause JSON
	LookUpCharacterID(steamid string, slot int) (snowflake.ID, error)
	SoftDeleteCharacter(id snowflake.ID, expiration time.Duration) error
	DeleteCharacter(id snowflake.ID) error
	DeleteCharacterReference(steamid string, slot int) error
	MoveCharacter(id snowflake.ID, steamid string, slot int) error
	CopyCharacter(id snowflake.ID, steamid string, slot int) (snowflake.ID, error)
	RestoreCharacter(id snowflake.ID) error
	RollbackCharacter(id snowflake.ID, ver int) error
	RollbackCharacterToLatest(id snowflake.ID) error
	DeleteCharacterVersions(id snowflake.ID) error

	SyncToDisk() error
	RunGC() error
}