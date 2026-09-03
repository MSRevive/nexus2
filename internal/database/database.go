package database

import (
	"context"
	"time"
	"log/slog"
	"errors"

	"github.com/msrevive/nexus2/internal/bitmask"
	"github.com/msrevive/nexus2/pkg/database/schema"

	"github.com/google/uuid"
	"github.com/titpetric/oida"
)

var (
	ErrNoDocument = errors.New("no document")
	ErrNotImplemented = errors.New("database not yet implemented")
	ErrNotAvailable = errors.New("database not available")
)

type Options struct {
	Logger *slog.Logger
	Tracer *oida.Tracer
}

type Database interface {
	Connect(cfg Config, opts Options) error
	Disconnect() error

	GetAllUsers(ctx context.Context) ([]*schema.User, error)
	GetUser(ctx context.Context, steamid string) (*schema.User, error)
	SetUserFlags(ctx context.Context, steamid string, flags bitmask.Bitmask) (error)
	GetUserFlags(ctx context.Context, steamid string) (bitmask.Bitmask, error)

	NewCharacter(ctx context.Context, steamid string, slot int, size int, data string) (uuid.UUID, error)
	UpdateCharacter(ctx context.Context, id uuid.UUID, size int, data string, backupMax int, backupTime time.Duration) error
	GetCharacter(ctx context.Context, id uuid.UUID) (*schema.Character, error)
	GetCharacters(ctx context.Context, steamid string) (map[int]schema.Character, error) //Gotta be a map cause JSON
	LookUpCharacterID(ctx context.Context, steamid string, slot int) (uuid.UUID, error)
	SoftDeleteCharacter(ctx context.Context, id uuid.UUID, expiration time.Duration) error
	DeleteCharacter(ctx context.Context, id uuid.UUID) error
	DeleteCharacterReference(ctx context.Context, steamid string, slot int) error
	MoveCharacter(ctx context.Context, id uuid.UUID, steamid string, slot int) error
	CopyCharacter(ctx context.Context, id uuid.UUID, steamid string, slot int) (uuid.UUID, error)
	RestoreCharacter(ctx context.Context, id uuid.UUID) error

	RollbackCharacter(ctx context.Context, id uuid.UUID, ver int) error
	RollbackCharacterToLatest(ctx context.Context, id uuid.UUID) error
	DeleteCharacterVersions(ctx context.Context, id uuid.UUID) error
	GetRollbackVersionsTimestamp(ctx context.Context, id uuid.UUID) (map[int]string, error)

	SyncToDisk(ctx context.Context) error
	RunGC(ctx context.Context) error
}