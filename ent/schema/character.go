package schema

import (
	"github.com/google/uuid"
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	//"entgo.io/ent/schema/edge"
)

// Character holds the schema definition for the Character entity.
type Character struct {
	ent.Schema
}

// Fields of the Character.
func (Character) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Immutable().
			Default(uuid.New),
		field.Uint64("steamid").
			Positive(),
		field.Int("slot").
			Positive(),
		field.String("name").
			NotEmpty(),
		field.Int("gender").
			Positive(),
		field.Int("race").
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
		field.Int("kills").
			Positive(),
		field.Int("gold").
			Positive(),
		field.String("skills").
			NotEmpty().
			Default("{}"),
		field.String("pets").
			NotEmpty().
			Default("{}"),
		field.Int("health").
			Positive(),
		field.Int("mana").
			Positive(),
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

func (Character) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id").
			Unique(),
		index.Fields("steamid", "slot"),
	}
}
