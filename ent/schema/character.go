package schema

import (
	"github.com/google/uuid"
	"entgo.io/ent"
)

// Character holds the schema definition for the Character entity.
type Character struct {
	ent.Schema
}

// Fields of the Character.
func (Character) Fields() []ent.Field {
	[]ent.Field{
		field.UUID("id", uuid.UUID{}).
			Immutable().
			Default(uuid.New),
		field.Uint64("steamid"),
		field.Byte("slot").
			Positive(),
		field.String("name").
			NotEmpty(),
		field.Byte("gender").
			Positive(),
		field.Byte("race").
			Positive(),
		field.String("flags").
			NotEmpty().
			Default("{}"),
		field.String("quickslots").
			NotEmpty().
			Default("{}"),
		field.String("quests").
			NotEmpty().
			Default("{}"),
		field.String("guild").
			NotEmpty(),
		field.Int16("kills").
			Positive(),
		field.Uint32("gold").
			Positive(),
		field.String("skills").
			NotEmpty().
			Default("{}"),
		field.String("pets").
			NotEmpty().
			Default("{}"),
		field.Int("health"),
		field.Int("mana"),
		field.String("equipped").
			NotEmpty().
			Default("{}"),
		field.String("lefthand").
			NotEmpty(),
		field.String("righthand").
			NotEmpty(),
		field.String("spells").
			NotEmpty().
			Default("{}"),
		field.String("spellbook").
			NotEmpty().
			Default("{}"),
		field.String("bags").
			NotEmpty().
			Default("{}"),
		field.String("sheaths").
			NotEmpty().
			Default("{}"),
	}
}

// Edges of the Character.
func (Character) Edges() []ent.Edge {
	return nil
}
