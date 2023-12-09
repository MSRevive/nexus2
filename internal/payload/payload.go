package payload

// We have payload structs because the game doesn't need some of the values that are
// inside the database like the character version or w/e

import (
	"github.com/google/uuid"
)

type Character struct {
	ID uuid.UUID `json:"id"`
	SteamID string `json:"steamid"`
	Slot int `json:"slot"`
	Size int `json:"size"`
	Data string `json:"data"`
}