package schema

import (
	"github.com/google/uuid"
	"entgo.io/ent"
	"entgo.io/ent/dialect"
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
		field.String("flags").
			SchemaType(map[string]string{
				dialect.SQLite: "text",
			}).
			NotEmpty().
			Default("{}"),
		field.String("quickslots").
			SchemaType(map[string]string{
				dialect.SQLite: "text",
			}).
			NotEmpty().
			Default("{}"),
		field.String("quests").
			SchemaType(map[string]string{
				dialect.SQLite: "text",
			}).
			NotEmpty().
			Default("{}"),
		field.String("guild"),
		field.Int("kills").
			Min(0),
		field.Int("gold").
			Min(0),
		field.String("skills").
			SchemaType(map[string]string{
				dialect.SQLite: "text",
			}).
			NotEmpty().
			Default("{}"),
		field.String("pets").
			SchemaType(map[string]string{
				dialect.SQLite: "text",
			}).
			NotEmpty().
			Default("{}"),
		field.Int("health").
			Min(0).
			Default(15),
		field.Int("mana").
			Min(0).
			Default(5),
		field.String("equipped").
			SchemaType(map[string]string{
				dialect.SQLite: "text",
			}).
			NotEmpty().
			Default("{}"),
		field.String("lefthand"),
		field.String("righthand"),
		field.String("spells").
			SchemaType(map[string]string{
				dialect.SQLite: "text",
			}).
			NotEmpty().
			Default("{}"),
		field.String("spellbook").
			SchemaType(map[string]string{
				dialect.SQLite: "text",
			}).
			NotEmpty().
			Default("{}"),
		field.Text("bags").
			SchemaType(map[string]string{
				dialect.SQLite: "text",
			}).
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
