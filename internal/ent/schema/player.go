package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Player holds the schema definition for the Player entity.
type Player struct {
	ent.Schema
}

func (Player) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimestampMixin{},
	}
}

// Fields of the Player.
func (Player) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Immutable().
			Default(uuid.New),
		field.String("steamid"),
	}
}

// Edges of the Player.
func (Player) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("characters", Character.Type),
	}
}

func (Player) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id").Unique(),
		index.Fields("steamid").Unique(),
	}
}
