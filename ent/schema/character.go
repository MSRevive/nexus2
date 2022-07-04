package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Character holds the schema definition for the Character entity.
type Character struct {
	ent.Schema
}

func (Character) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimestampMixin{},
	}
}

// Fields of the Character.
func (Character) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Immutable().
			Default(uuid.New),
		field.UUID("player_id", uuid.UUID{}),
		field.Int("version").
			Min(1).
			StructTag(`json:"version"`),
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
		field.Time("deleted_at").
			Optional().
			Nillable(),
	}
}

// Edges of the Character.
func (Character) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("player", Player.Type).
			Ref("characters").
			Unique().
			Field("player_id").
			Required(),
	}
}

func (Character) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id").Unique(),
		index.Fields("player_id", "slot", "version").Unique(),
	}
}
