package schema

import (
	"time"

	"github.com/google/uuid"
)

type CharacterData struct {
	CreatedAt time.Time `bson:"created_at"`
	Size int `bson:"size"`
	Data string `bson:"data"`
}

//Characters collection
type Character struct {
	ID uuid.UUID `bson:"_id"`
	SteamID string `bson:"steamid"`
	Slot int `bson:"slot"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
	DeletedAt *time.Time `bson:"deleted_at,omitempty"`
	Versions []CharacterData `bson:"versions"` //Version => character data
}

type User struct {
	ID string `bson:"_id"` //this is the SteamID64
	Characters map[int]uuid.UUID `bson:"characters"` //Slot => reference Character by ID
	DeletedCharacters map[int]uuid.UUID `bson:"deleted_characters"`
}