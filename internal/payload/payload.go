package payload

// We have payload structs because the game doesn't need some of the values that are
// inside the database like the character version or w/e

import (
	"github.com/msrevive/nexus2/internal/bitmask"

	"github.com/bwmarrin/snowflake"
)

type Character struct {
	ID snowflake.ID `json:"id"`
	SteamID string `json:"steamid"`
	Slot int `json:"slot"`
	Size int `json:"size"`
	Data string `json:"data"`
	Flags bitmask.Bitmask `json:"flags"`
}

type CharacterCreate struct {
	ID snowflake.ID `json:"id"`
	Flags bitmask.Bitmask `json:"flags"`
}