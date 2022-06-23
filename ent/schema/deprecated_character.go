package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// DeprecatedCharacter holds the schema definition for the DeprecatedCharacter entity.
type DeprecatedCharacter struct {
	ent.Schema
}

// Fields of the DeprecatedCharacter.
func (DeprecatedCharacter) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Immutable().
			Default(uuid.New),
		field.String("steamid"),
		field.Int("slot").
			Min(0).
			Default(0).
			StructTag(`json:"slot"`),
		field.Int("size").
			Default(0),
		field.String("data").
			SchemaType(map[string]string{
				dialect.SQLite: "text",
			}),
	}
}

func (DeprecatedCharacter) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "old_characters"},
	}
}

// Edges of the DeprecatedCharacter.
func (DeprecatedCharacter) Edges() []ent.Edge {
	return nil
}

func (DeprecatedCharacter) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id").
			Unique(),
		index.Fields("steamid", "slot"),
	}
}
