package schema

import (
	"time"
)

type CharacterData struct {
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	Size int `bson:"size" json:"size"`
	Data string `bson:"data" json:"data"`
}