// Code generated by entc, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/msrevive/nexus2/ent/character"
	"github.com/msrevive/nexus2/ent/player"
)

// Character is the model entity for the Character schema.
type Character struct {
	config `json:"-"`
	// ID of the ent.
	ID uuid.UUID `json:"id,omitempty"`
	// CreatedAt holds the value of the "created_at" field.
	CreatedAt time.Time `json:"created_at,omitempty"`
	// UpdatedAt holds the value of the "updated_at" field.
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	// PlayerID holds the value of the "player_id" field.
	PlayerID uuid.UUID `json:"player_id,omitempty"`
	// Version holds the value of the "version" field.
	Version int `json:"version"`
	// Slot holds the value of the "slot" field.
	Slot int `json:"slot"`
	// Size holds the value of the "size" field.
	Size int `json:"size,omitempty"`
	// Data holds the value of the "data" field.
	Data string `json:"data,omitempty"`
	// DeletedAt holds the value of the "deleted_at" field.
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the CharacterQuery when eager-loading is set.
	Edges CharacterEdges `json:"edges"`
}

// CharacterEdges holds the relations/edges for other nodes in the graph.
type CharacterEdges struct {
	// Player holds the value of the player edge.
	Player *Player `json:"player,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [1]bool
}

// PlayerOrErr returns the Player value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e CharacterEdges) PlayerOrErr() (*Player, error) {
	if e.loadedTypes[0] {
		if e.Player == nil {
			// The edge player was loaded in eager-loading,
			// but was not found.
			return nil, &NotFoundError{label: player.Label}
		}
		return e.Player, nil
	}
	return nil, &NotLoadedError{edge: "player"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*Character) scanValues(columns []string) ([]interface{}, error) {
	values := make([]interface{}, len(columns))
	for i := range columns {
		switch columns[i] {
		case character.FieldVersion, character.FieldSlot, character.FieldSize:
			values[i] = new(sql.NullInt64)
		case character.FieldData:
			values[i] = new(sql.NullString)
		case character.FieldCreatedAt, character.FieldUpdatedAt, character.FieldDeletedAt:
			values[i] = new(sql.NullTime)
		case character.FieldID, character.FieldPlayerID:
			values[i] = new(uuid.UUID)
		default:
			return nil, fmt.Errorf("unexpected column %q for type Character", columns[i])
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the Character fields.
func (c *Character) assignValues(columns []string, values []interface{}) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case character.FieldID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value != nil {
				c.ID = *value
			}
		case character.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				c.CreatedAt = value.Time
			}
		case character.FieldUpdatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field updated_at", values[i])
			} else if value.Valid {
				c.UpdatedAt = value.Time
			}
		case character.FieldPlayerID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field player_id", values[i])
			} else if value != nil {
				c.PlayerID = *value
			}
		case character.FieldVersion:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field version", values[i])
			} else if value.Valid {
				c.Version = int(value.Int64)
			}
		case character.FieldSlot:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field slot", values[i])
			} else if value.Valid {
				c.Slot = int(value.Int64)
			}
		case character.FieldSize:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field size", values[i])
			} else if value.Valid {
				c.Size = int(value.Int64)
			}
		case character.FieldData:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field data", values[i])
			} else if value.Valid {
				c.Data = value.String
			}
		case character.FieldDeletedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field deleted_at", values[i])
			} else if value.Valid {
				c.DeletedAt = new(time.Time)
				*c.DeletedAt = value.Time
			}
		}
	}
	return nil
}

// QueryPlayer queries the "player" edge of the Character entity.
func (c *Character) QueryPlayer() *PlayerQuery {
	return (&CharacterClient{config: c.config}).QueryPlayer(c)
}

// Update returns a builder for updating this Character.
// Note that you need to call Character.Unwrap() before calling this method if this Character
// was returned from a transaction, and the transaction was committed or rolled back.
func (c *Character) Update() *CharacterUpdateOne {
	return (&CharacterClient{config: c.config}).UpdateOne(c)
}

// Unwrap unwraps the Character entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (c *Character) Unwrap() *Character {
	tx, ok := c.config.driver.(*txDriver)
	if !ok {
		panic("ent: Character is not a transactional entity")
	}
	c.config.driver = tx.drv
	return c
}

// String implements the fmt.Stringer.
func (c *Character) String() string {
	var builder strings.Builder
	builder.WriteString("Character(")
	builder.WriteString(fmt.Sprintf("id=%v", c.ID))
	builder.WriteString(", created_at=")
	builder.WriteString(c.CreatedAt.Format(time.ANSIC))
	builder.WriteString(", updated_at=")
	builder.WriteString(c.UpdatedAt.Format(time.ANSIC))
	builder.WriteString(", player_id=")
	builder.WriteString(fmt.Sprintf("%v", c.PlayerID))
	builder.WriteString(", version=")
	builder.WriteString(fmt.Sprintf("%v", c.Version))
	builder.WriteString(", slot=")
	builder.WriteString(fmt.Sprintf("%v", c.Slot))
	builder.WriteString(", size=")
	builder.WriteString(fmt.Sprintf("%v", c.Size))
	builder.WriteString(", data=")
	builder.WriteString(c.Data)
	if v := c.DeletedAt; v != nil {
		builder.WriteString(", deleted_at=")
		builder.WriteString(v.Format(time.ANSIC))
	}
	builder.WriteByte(')')
	return builder.String()
}

// Characters is a parsable slice of Character.
type Characters []*Character

func (c Characters) config(cfg config) {
	for _i := range c {
		c[_i].config = cfg
	}
}
