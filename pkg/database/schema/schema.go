package schema

import (
	"time"

	"github.com/google/uuid"
)

//Characters collection
type Character struct {
	ID uuid.UUID `bson:"_id" json:"_id"`
	SteamID string `bson:"steamid" json:"steamid"`
	Slot int `bson:"slot" json:"slot"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
	Data CharacterData `bson:"data,omitempty" json:"data,omitempty"`
	Versions []CharacterData `bson:"versions" json:"versions"` //Version => character data
}

type User struct {
	ID string `bson:"_id" json:"_id"` //this is the SteamID64
	Flags uint32 `bson:"flags" json:"flags"` //account flags
	Characters map[int]uuid.UUID `bson:"characters" json:"characters"` //Slot => reference Character by ID
	DeletedCharacters map[int]uuid.UUID `bson:"deleted_characters" json:"deleted_characters"`
}