package schema

import (
	"time"

	"github.com/google/uuid"
)

type CharacterData struct {
	Version int `bson:"version"`
	CreatedAt time.Time `bson:"created_at"`
	Size int `bson:"size"`
	Data string `bson:"data"`
}

type Character struct {
	ID uuid.UUID `bson:"id"`
	Slot int `bson:"slot`
	UpdatedAt time.Time `bson:"updated_at"`
	DeletedAt time.Time `bson:"deleted_at"`
	Versions map[int]CharacterData `bson:"data"` //Version => character data
}

type User struct {
	ID string `bson:"_id"` //this is the SteamID64
	Characters map[uuid.UUID]Character `bson:"characters"` //Slot => character
}