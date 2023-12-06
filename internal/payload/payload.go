package payload

// We have payload structs because the game doesn't need some of the values that are
// inside the database like the character version or w/e

import (

)

type Character struct {
	SteamID string
	Slot int
	Size int
	Data string
}