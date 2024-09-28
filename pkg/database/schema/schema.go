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
	DeletedAt *time.Time `bson:"deleted_at,omitempty"`
	Data CharacterData `bson:"data,omitempty"`
	Versions []CharacterData `bson:"versions"` //Version => character data
}

type User struct {
	ID string `bson:"_id"` //this is the SteamID64
	Revision int `bson:"revison"` //we store what revision the user is on so in the future we can automagically update users.
	Flags uint32 `bson:"flags"` //account flags
	Characters map[int]uuid.UUID `bson:"characters"` //Slot => reference Character by ID
	DeletedCharacters map[int]uuid.UUID `bson:"deleted_characters"`
}