package schema

import (
	"time"

	"github.com/bwmarrin/snowflake"
)

//Characters collection
type Character struct {
	ID snowflake.ID `bson:"_id" json:"_id"`
	SteamID string `bson:"steamid" json:"steamid"`
	Slot int `bson:"slot" json:"slot"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
	Data CharacterData `bson:"data,omitempty" json:"data,omitempty"`
	Versions []CharacterData `bson:"versions" json:"versions"` //Version => character data
}

type User struct {
	ID string `bson:"_id" json:"_id"` //this is the SteamID64
	Revision int `bson:"revison" json:"revision"` //we store what revision the user is on so in the future we can automagically update users.
	Flags uint32 `bson:"flags" json:"flags"` //account flags
	Characters map[int]snowflake.ID `bson:"characters" json:"characters"` //Slot => reference Character by ID
	DeletedCharacters map[int]snowflake.ID `bson:"deleted_characters" json:"deleted_characters"`
}

//
//	Children Schema
//
type CharacterData struct {
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	Size int `bson:"size" json:"size"`
	Data string `bson:"data" json:"data"`
}