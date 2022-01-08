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
		field.String("steamid"),
		field.Int("slot").
			Min(0),
		field.String("name").
			NotEmpty(),
		field.Int("gender").
			Min(0),
		field.Int("race").
			Min(0),
		field.String("flags").
			NotEmpty().
			Default("{}"),
		field.String("quickslots").
			NotEmpty().
			Default("{}"),
		field.String("quests").
			NotEmpty().
			Default("{}"),
		field.String("guild"),
		field.Int("kills").
			Min(0),
		field.Int("gold").
			Min(0),
		field.String("skills").
			NotEmpty().
			Default("{}"),
		field.String("pets").
			NotEmpty().
			Default("{}"),
		field.Int("health").
			Min(0),
		field.Int("mana").
			Min(0),
		field.String("equipped").
			NotEmpty().
			Default("{}"),
		field.String("lefthand"),
		field.String("righthand"),
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
