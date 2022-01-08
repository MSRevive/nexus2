// Code generated by entc, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/msrevive/nexus2/ent/character"
)

// Character is the model entity for the Character schema.
type Character struct {
	config `json:"-"`
	// ID of the ent.
	ID uuid.UUID `json:"id,omitempty"`
	// Steamid holds the value of the "steamid" field.
	Steamid string `json:"steamid,omitempty"`
	// Slot holds the value of the "slot" field.
	Slot int `json:"slot,omitempty"`
	// Name holds the value of the "name" field.
	Name string `json:"name,omitempty"`
	// Gender holds the value of the "gender" field.
	Gender int `json:"gender,omitempty"`
	// Race holds the value of the "race" field.
	Race int `json:"race,omitempty"`
	// Flags holds the value of the "flags" field.
	Flags string `json:"flags,omitempty"`
	// Quickslots holds the value of the "quickslots" field.
	Quickslots string `json:"quickslots,omitempty"`
	// Quests holds the value of the "quests" field.
	Quests string `json:"quests,omitempty"`
	// Guild holds the value of the "guild" field.
	Guild string `json:"guild,omitempty"`
	// Kills holds the value of the "kills" field.
	Kills int `json:"kills,omitempty"`
	// Gold holds the value of the "gold" field.
	Gold int `json:"gold,omitempty"`
	// Skills holds the value of the "skills" field.
	Skills string `json:"skills,omitempty"`
	// Pets holds the value of the "pets" field.
	Pets string `json:"pets,omitempty"`
	// Health holds the value of the "health" field.
	Health int `json:"health,omitempty"`
	// Mana holds the value of the "mana" field.
	Mana int `json:"mana,omitempty"`
	// Equipped holds the value of the "equipped" field.
	Equipped string `json:"equipped,omitempty"`
	// Lefthand holds the value of the "lefthand" field.
	Lefthand string `json:"lefthand,omitempty"`
	// Righthand holds the value of the "righthand" field.
	Righthand string `json:"righthand,omitempty"`
	// Spells holds the value of the "spells" field.
	Spells string `json:"spells,omitempty"`
	// Spellbook holds the value of the "spellbook" field.
	Spellbook string `json:"spellbook,omitempty"`
	// Bags holds the value of the "bags" field.
	Bags string `json:"bags,omitempty"`
	// Sheaths holds the value of the "sheaths" field.
	Sheaths string `json:"sheaths,omitempty"`
}

// scanValues returns the types for scanning values from sql.Rows.
func (*Character) scanValues(columns []string) ([]interface{}, error) {
	values := make([]interface{}, len(columns))
	for i := range columns {
		switch columns[i] {
		case character.FieldSlot, character.FieldGender, character.FieldRace, character.FieldKills, character.FieldGold, character.FieldHealth, character.FieldMana:
			values[i] = new(sql.NullInt64)
		case character.FieldSteamid, character.FieldName, character.FieldFlags, character.FieldQuickslots, character.FieldQuests, character.FieldGuild, character.FieldSkills, character.FieldPets, character.FieldEquipped, character.FieldLefthand, character.FieldRighthand, character.FieldSpells, character.FieldSpellbook, character.FieldBags, character.FieldSheaths:
			values[i] = new(sql.NullString)
		case character.FieldID:
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
		case character.FieldSteamid:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field steamid", values[i])
			} else if value.Valid {
				c.Steamid = value.String
			}
		case character.FieldSlot:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field slot", values[i])
			} else if value.Valid {
				c.Slot = int(value.Int64)
			}
		case character.FieldName:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field name", values[i])
			} else if value.Valid {
				c.Name = value.String
			}
		case character.FieldGender:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field gender", values[i])
			} else if value.Valid {
				c.Gender = int(value.Int64)
			}
		case character.FieldRace:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field race", values[i])
			} else if value.Valid {
				c.Race = int(value.Int64)
			}
		case character.FieldFlags:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field flags", values[i])
			} else if value.Valid {
				c.Flags = value.String
			}
		case character.FieldQuickslots:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field quickslots", values[i])
			} else if value.Valid {
				c.Quickslots = value.String
			}
		case character.FieldQuests:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field quests", values[i])
			} else if value.Valid {
				c.Quests = value.String
			}
		case character.FieldGuild:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field guild", values[i])
			} else if value.Valid {
				c.Guild = value.String
			}
		case character.FieldKills:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field kills", values[i])
			} else if value.Valid {
				c.Kills = int(value.Int64)
			}
		case character.FieldGold:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field gold", values[i])
			} else if value.Valid {
				c.Gold = int(value.Int64)
			}
		case character.FieldSkills:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field skills", values[i])
			} else if value.Valid {
				c.Skills = value.String
			}
		case character.FieldPets:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field pets", values[i])
			} else if value.Valid {
				c.Pets = value.String
			}
		case character.FieldHealth:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field health", values[i])
			} else if value.Valid {
				c.Health = int(value.Int64)
			}
		case character.FieldMana:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field mana", values[i])
			} else if value.Valid {
				c.Mana = int(value.Int64)
			}
		case character.FieldEquipped:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field equipped", values[i])
			} else if value.Valid {
				c.Equipped = value.String
			}
		case character.FieldLefthand:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field lefthand", values[i])
			} else if value.Valid {
				c.Lefthand = value.String
			}
		case character.FieldRighthand:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field righthand", values[i])
			} else if value.Valid {
				c.Righthand = value.String
			}
		case character.FieldSpells:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field spells", values[i])
			} else if value.Valid {
				c.Spells = value.String
			}
		case character.FieldSpellbook:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field spellbook", values[i])
			} else if value.Valid {
				c.Spellbook = value.String
			}
		case character.FieldBags:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field bags", values[i])
			} else if value.Valid {
				c.Bags = value.String
			}
		case character.FieldSheaths:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field sheaths", values[i])
			} else if value.Valid {
				c.Sheaths = value.String
			}
		}
	}
	return nil
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
	builder.WriteString(", steamid=")
	builder.WriteString(c.Steamid)
	builder.WriteString(", slot=")
	builder.WriteString(fmt.Sprintf("%v", c.Slot))
	builder.WriteString(", name=")
	builder.WriteString(c.Name)
	builder.WriteString(", gender=")
	builder.WriteString(fmt.Sprintf("%v", c.Gender))
	builder.WriteString(", race=")
	builder.WriteString(fmt.Sprintf("%v", c.Race))
	builder.WriteString(", flags=")
	builder.WriteString(c.Flags)
	builder.WriteString(", quickslots=")
	builder.WriteString(c.Quickslots)
	builder.WriteString(", quests=")
	builder.WriteString(c.Quests)
	builder.WriteString(", guild=")
	builder.WriteString(c.Guild)
	builder.WriteString(", kills=")
	builder.WriteString(fmt.Sprintf("%v", c.Kills))
	builder.WriteString(", gold=")
	builder.WriteString(fmt.Sprintf("%v", c.Gold))
	builder.WriteString(", skills=")
	builder.WriteString(c.Skills)
	builder.WriteString(", pets=")
	builder.WriteString(c.Pets)
	builder.WriteString(", health=")
	builder.WriteString(fmt.Sprintf("%v", c.Health))
	builder.WriteString(", mana=")
	builder.WriteString(fmt.Sprintf("%v", c.Mana))
	builder.WriteString(", equipped=")
	builder.WriteString(c.Equipped)
	builder.WriteString(", lefthand=")
	builder.WriteString(c.Lefthand)
	builder.WriteString(", righthand=")
	builder.WriteString(c.Righthand)
	builder.WriteString(", spells=")
	builder.WriteString(c.Spells)
	builder.WriteString(", spellbook=")
	builder.WriteString(c.Spellbook)
	builder.WriteString(", bags=")
	builder.WriteString(c.Bags)
	builder.WriteString(", sheaths=")
	builder.WriteString(c.Sheaths)
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
