// Code generated by entc, DO NOT EDIT.

package ent

import (
	"time"

	"github.com/google/uuid"
	"github.com/msrevive/nexus2/ent/character"
	"github.com/msrevive/nexus2/ent/deprecatedcharacter"
	"github.com/msrevive/nexus2/ent/player"
	"github.com/msrevive/nexus2/ent/schema"
)

// The init function reads all schema descriptors with runtime code
// (default values, validators, hooks and policies) and stitches it
// to their package variables.
func init() {
	characterMixin := schema.Character{}.Mixin()
	characterMixinFields0 := characterMixin[0].Fields()
	_ = characterMixinFields0
	characterFields := schema.Character{}.Fields()
	_ = characterFields
	// characterDescCreatedAt is the schema descriptor for created_at field.
	characterDescCreatedAt := characterMixinFields0[0].Descriptor()
	// character.DefaultCreatedAt holds the default value on creation for the created_at field.
	character.DefaultCreatedAt = characterDescCreatedAt.Default.(func() time.Time)
	// characterDescUpdatedAt is the schema descriptor for updated_at field.
	characterDescUpdatedAt := characterMixinFields0[1].Descriptor()
	// character.DefaultUpdatedAt holds the default value on creation for the updated_at field.
	character.DefaultUpdatedAt = characterDescUpdatedAt.Default.(func() time.Time)
	// character.UpdateDefaultUpdatedAt holds the default value on update for the updated_at field.
	character.UpdateDefaultUpdatedAt = characterDescUpdatedAt.UpdateDefault.(func() time.Time)
	// characterDescVersion is the schema descriptor for version field.
	characterDescVersion := characterFields[2].Descriptor()
	// character.VersionValidator is a validator for the "version" field. It is called by the builders before save.
	character.VersionValidator = characterDescVersion.Validators[0].(func(int) error)
	// characterDescSlot is the schema descriptor for slot field.
	characterDescSlot := characterFields[3].Descriptor()
	// character.DefaultSlot holds the default value on creation for the slot field.
	character.DefaultSlot = characterDescSlot.Default.(int)
	// character.SlotValidator is a validator for the "slot" field. It is called by the builders before save.
	character.SlotValidator = characterDescSlot.Validators[0].(func(int) error)
	// characterDescSize is the schema descriptor for size field.
	characterDescSize := characterFields[4].Descriptor()
	// character.DefaultSize holds the default value on creation for the size field.
	character.DefaultSize = characterDescSize.Default.(int)
	// characterDescID is the schema descriptor for id field.
	characterDescID := characterFields[0].Descriptor()
	// character.DefaultID holds the default value on creation for the id field.
	character.DefaultID = characterDescID.Default.(func() uuid.UUID)
	deprecatedcharacterFields := schema.DeprecatedCharacter{}.Fields()
	_ = deprecatedcharacterFields
	// deprecatedcharacterDescSlot is the schema descriptor for slot field.
	deprecatedcharacterDescSlot := deprecatedcharacterFields[2].Descriptor()
	// deprecatedcharacter.DefaultSlot holds the default value on creation for the slot field.
	deprecatedcharacter.DefaultSlot = deprecatedcharacterDescSlot.Default.(int)
	// deprecatedcharacter.SlotValidator is a validator for the "slot" field. It is called by the builders before save.
	deprecatedcharacter.SlotValidator = deprecatedcharacterDescSlot.Validators[0].(func(int) error)
	// deprecatedcharacterDescSize is the schema descriptor for size field.
	deprecatedcharacterDescSize := deprecatedcharacterFields[3].Descriptor()
	// deprecatedcharacter.DefaultSize holds the default value on creation for the size field.
	deprecatedcharacter.DefaultSize = deprecatedcharacterDescSize.Default.(int)
	// deprecatedcharacterDescID is the schema descriptor for id field.
	deprecatedcharacterDescID := deprecatedcharacterFields[0].Descriptor()
	// deprecatedcharacter.DefaultID holds the default value on creation for the id field.
	deprecatedcharacter.DefaultID = deprecatedcharacterDescID.Default.(func() uuid.UUID)
	playerMixin := schema.Player{}.Mixin()
	playerMixinFields0 := playerMixin[0].Fields()
	_ = playerMixinFields0
	playerFields := schema.Player{}.Fields()
	_ = playerFields
	// playerDescCreatedAt is the schema descriptor for created_at field.
	playerDescCreatedAt := playerMixinFields0[0].Descriptor()
	// player.DefaultCreatedAt holds the default value on creation for the created_at field.
	player.DefaultCreatedAt = playerDescCreatedAt.Default.(func() time.Time)
	// playerDescUpdatedAt is the schema descriptor for updated_at field.
	playerDescUpdatedAt := playerMixinFields0[1].Descriptor()
	// player.DefaultUpdatedAt holds the default value on creation for the updated_at field.
	player.DefaultUpdatedAt = playerDescUpdatedAt.Default.(func() time.Time)
	// player.UpdateDefaultUpdatedAt holds the default value on update for the updated_at field.
	player.UpdateDefaultUpdatedAt = playerDescUpdatedAt.UpdateDefault.(func() time.Time)
	// playerDescID is the schema descriptor for id field.
	playerDescID := playerFields[0].Descriptor()
	// player.DefaultID holds the default value on creation for the id field.
	player.DefaultID = playerDescID.Default.(func() uuid.UUID)
}
