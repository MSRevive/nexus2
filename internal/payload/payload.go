package payload

// We have payload structs because the game doesn't need some of the values that are
// inside the database like the character version or w/e

import (

)

type Character struct {
	SteamID string `json:"steamid"`
	Slot int `json:"slot"`
	Size int `json:"size"`
	Data string `json:"data"`
}