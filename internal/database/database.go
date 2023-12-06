package database

import (
	"github.com/msrevive/nexus2/internal/database/schema"

	//"github.com/google/uuid"
)

type Database interface {
	Connect(conn string) error
	Disconnect() error

	NewCharacter(steamid string, slot int, size int, data string) error
	GetCharacters(steamid string) (map[int]schema.Character, error)
	GetUser(steamid string) (*schema.User, error)
	//UpdateCharacter(steamid string, slot int)
}