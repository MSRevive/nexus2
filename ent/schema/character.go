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
		field.Int("size").
			Default(0),
		field.String("data").
			SchemaType(map[string]string{
				dialect.SQLite: "text",
			}),
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
