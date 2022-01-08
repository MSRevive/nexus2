// Code generated by entc, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/msrevive/nexus2/ent/character"
)

// CharacterCreate is the builder for creating a Character entity.
type CharacterCreate struct {
	config
	mutation *CharacterMutation
	hooks    []Hook
}

// SetSteamid sets the "steamid" field.
func (cc *CharacterCreate) SetSteamid(s string) *CharacterCreate {
	cc.mutation.SetSteamid(s)
	return cc
}

// SetSlot sets the "slot" field.
func (cc *CharacterCreate) SetSlot(i int) *CharacterCreate {
	cc.mutation.SetSlot(i)
	return cc
}

// SetName sets the "name" field.
func (cc *CharacterCreate) SetName(s string) *CharacterCreate {
	cc.mutation.SetName(s)
	return cc
}

// SetGender sets the "gender" field.
func (cc *CharacterCreate) SetGender(i int) *CharacterCreate {
	cc.mutation.SetGender(i)
	return cc
}

// SetRace sets the "race" field.
func (cc *CharacterCreate) SetRace(i int) *CharacterCreate {
	cc.mutation.SetRace(i)
	return cc
}

// SetFlags sets the "flags" field.
func (cc *CharacterCreate) SetFlags(s string) *CharacterCreate {
	cc.mutation.SetFlags(s)
	return cc
}

// SetNillableFlags sets the "flags" field if the given value is not nil.
func (cc *CharacterCreate) SetNillableFlags(s *string) *CharacterCreate {
	if s != nil {
		cc.SetFlags(*s)
	}
	return cc
}

// SetQuickslots sets the "quickslots" field.
func (cc *CharacterCreate) SetQuickslots(s string) *CharacterCreate {
	cc.mutation.SetQuickslots(s)
	return cc
}

// SetNillableQuickslots sets the "quickslots" field if the given value is not nil.
func (cc *CharacterCreate) SetNillableQuickslots(s *string) *CharacterCreate {
	if s != nil {
		cc.SetQuickslots(*s)
	}
	return cc
}

// SetQuests sets the "quests" field.
func (cc *CharacterCreate) SetQuests(s string) *CharacterCreate {
	cc.mutation.SetQuests(s)
	return cc
}

// SetNillableQuests sets the "quests" field if the given value is not nil.
func (cc *CharacterCreate) SetNillableQuests(s *string) *CharacterCreate {
	if s != nil {
		cc.SetQuests(*s)
	}
	return cc
}

// SetGuild sets the "guild" field.
func (cc *CharacterCreate) SetGuild(s string) *CharacterCreate {
	cc.mutation.SetGuild(s)
	return cc
}

// SetKills sets the "kills" field.
func (cc *CharacterCreate) SetKills(i int) *CharacterCreate {
	cc.mutation.SetKills(i)
	return cc
}

// SetGold sets the "gold" field.
func (cc *CharacterCreate) SetGold(i int) *CharacterCreate {
	cc.mutation.SetGold(i)
	return cc
}

// SetSkills sets the "skills" field.
func (cc *CharacterCreate) SetSkills(s string) *CharacterCreate {
	cc.mutation.SetSkills(s)
	return cc
}

// SetNillableSkills sets the "skills" field if the given value is not nil.
func (cc *CharacterCreate) SetNillableSkills(s *string) *CharacterCreate {
	if s != nil {
		cc.SetSkills(*s)
	}
	return cc
}

// SetPets sets the "pets" field.
func (cc *CharacterCreate) SetPets(s string) *CharacterCreate {
	cc.mutation.SetPets(s)
	return cc
}

// SetNillablePets sets the "pets" field if the given value is not nil.
func (cc *CharacterCreate) SetNillablePets(s *string) *CharacterCreate {
	if s != nil {
		cc.SetPets(*s)
	}
	return cc
}

// SetHealth sets the "health" field.
func (cc *CharacterCreate) SetHealth(i int) *CharacterCreate {
	cc.mutation.SetHealth(i)
	return cc
}

// SetMana sets the "mana" field.
func (cc *CharacterCreate) SetMana(i int) *CharacterCreate {
	cc.mutation.SetMana(i)
	return cc
}

// SetEquipped sets the "equipped" field.
func (cc *CharacterCreate) SetEquipped(s string) *CharacterCreate {
	cc.mutation.SetEquipped(s)
	return cc
}

// SetNillableEquipped sets the "equipped" field if the given value is not nil.
func (cc *CharacterCreate) SetNillableEquipped(s *string) *CharacterCreate {
	if s != nil {
		cc.SetEquipped(*s)
	}
	return cc
}

// SetLefthand sets the "lefthand" field.
func (cc *CharacterCreate) SetLefthand(s string) *CharacterCreate {
	cc.mutation.SetLefthand(s)
	return cc
}

// SetRighthand sets the "righthand" field.
func (cc *CharacterCreate) SetRighthand(s string) *CharacterCreate {
	cc.mutation.SetRighthand(s)
	return cc
}

// SetSpells sets the "spells" field.
func (cc *CharacterCreate) SetSpells(s string) *CharacterCreate {
	cc.mutation.SetSpells(s)
	return cc
}

// SetNillableSpells sets the "spells" field if the given value is not nil.
func (cc *CharacterCreate) SetNillableSpells(s *string) *CharacterCreate {
	if s != nil {
		cc.SetSpells(*s)
	}
	return cc
}

// SetSpellbook sets the "spellbook" field.
func (cc *CharacterCreate) SetSpellbook(s string) *CharacterCreate {
	cc.mutation.SetSpellbook(s)
	return cc
}

// SetNillableSpellbook sets the "spellbook" field if the given value is not nil.
func (cc *CharacterCreate) SetNillableSpellbook(s *string) *CharacterCreate {
	if s != nil {
		cc.SetSpellbook(*s)
	}
	return cc
}

// SetBags sets the "bags" field.
func (cc *CharacterCreate) SetBags(s string) *CharacterCreate {
	cc.mutation.SetBags(s)
	return cc
}

// SetNillableBags sets the "bags" field if the given value is not nil.
func (cc *CharacterCreate) SetNillableBags(s *string) *CharacterCreate {
	if s != nil {
		cc.SetBags(*s)
	}
	return cc
}

// SetSheaths sets the "sheaths" field.
func (cc *CharacterCreate) SetSheaths(s string) *CharacterCreate {
	cc.mutation.SetSheaths(s)
	return cc
}

// SetNillableSheaths sets the "sheaths" field if the given value is not nil.
func (cc *CharacterCreate) SetNillableSheaths(s *string) *CharacterCreate {
	if s != nil {
		cc.SetSheaths(*s)
	}
	return cc
}

// SetID sets the "id" field.
func (cc *CharacterCreate) SetID(u uuid.UUID) *CharacterCreate {
	cc.mutation.SetID(u)
	return cc
}

// Mutation returns the CharacterMutation object of the builder.
func (cc *CharacterCreate) Mutation() *CharacterMutation {
	return cc.mutation
}

// Save creates the Character in the database.
func (cc *CharacterCreate) Save(ctx context.Context) (*Character, error) {
	var (
		err  error
		node *Character
	)
	cc.defaults()
	if len(cc.hooks) == 0 {
		if err = cc.check(); err != nil {
			return nil, err
		}
		node, err = cc.sqlSave(ctx)
	} else {
		var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
			mutation, ok := m.(*CharacterMutation)
			if !ok {
				return nil, fmt.Errorf("unexpected mutation type %T", m)
			}
			if err = cc.check(); err != nil {
				return nil, err
			}
			cc.mutation = mutation
			if node, err = cc.sqlSave(ctx); err != nil {
				return nil, err
			}
			mutation.id = &node.ID
			mutation.done = true
			return node, err
		})
		for i := len(cc.hooks) - 1; i >= 0; i-- {
			if cc.hooks[i] == nil {
				return nil, fmt.Errorf("ent: uninitialized hook (forgotten import ent/runtime?)")
			}
			mut = cc.hooks[i](mut)
		}
		if _, err := mut.Mutate(ctx, cc.mutation); err != nil {
			return nil, err
		}
	}
	return node, err
}

// SaveX calls Save and panics if Save returns an error.
func (cc *CharacterCreate) SaveX(ctx context.Context) *Character {
	v, err := cc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (cc *CharacterCreate) Exec(ctx context.Context) error {
	_, err := cc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (cc *CharacterCreate) ExecX(ctx context.Context) {
	if err := cc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (cc *CharacterCreate) defaults() {
	if _, ok := cc.mutation.Flags(); !ok {
		v := character.DefaultFlags
		cc.mutation.SetFlags(v)
	}
	if _, ok := cc.mutation.Quickslots(); !ok {
		v := character.DefaultQuickslots
		cc.mutation.SetQuickslots(v)
	}
	if _, ok := cc.mutation.Quests(); !ok {
		v := character.DefaultQuests
		cc.mutation.SetQuests(v)
	}
	if _, ok := cc.mutation.Skills(); !ok {
		v := character.DefaultSkills
		cc.mutation.SetSkills(v)
	}
	if _, ok := cc.mutation.Pets(); !ok {
		v := character.DefaultPets
		cc.mutation.SetPets(v)
	}
	if _, ok := cc.mutation.Equipped(); !ok {
		v := character.DefaultEquipped
		cc.mutation.SetEquipped(v)
	}
	if _, ok := cc.mutation.Spells(); !ok {
		v := character.DefaultSpells
		cc.mutation.SetSpells(v)
	}
	if _, ok := cc.mutation.Spellbook(); !ok {
		v := character.DefaultSpellbook
		cc.mutation.SetSpellbook(v)
	}
	if _, ok := cc.mutation.Bags(); !ok {
		v := character.DefaultBags
		cc.mutation.SetBags(v)
	}
	if _, ok := cc.mutation.Sheaths(); !ok {
		v := character.DefaultSheaths
		cc.mutation.SetSheaths(v)
	}
	if _, ok := cc.mutation.ID(); !ok {
		v := character.DefaultID()
		cc.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (cc *CharacterCreate) check() error {
	if _, ok := cc.mutation.Steamid(); !ok {
		return &ValidationError{Name: "steamid", err: errors.New(`ent: missing required field "Character.steamid"`)}
	}
	if _, ok := cc.mutation.Slot(); !ok {
		return &ValidationError{Name: "slot", err: errors.New(`ent: missing required field "Character.slot"`)}
	}
	if v, ok := cc.mutation.Slot(); ok {
		if err := character.SlotValidator(v); err != nil {
			return &ValidationError{Name: "slot", err: fmt.Errorf(`ent: validator failed for field "Character.slot": %w`, err)}
		}
	}
	if _, ok := cc.mutation.Name(); !ok {
		return &ValidationError{Name: "name", err: errors.New(`ent: missing required field "Character.name"`)}
	}
	if v, ok := cc.mutation.Name(); ok {
		if err := character.NameValidator(v); err != nil {
			return &ValidationError{Name: "name", err: fmt.Errorf(`ent: validator failed for field "Character.name": %w`, err)}
		}
	}
	if _, ok := cc.mutation.Gender(); !ok {
		return &ValidationError{Name: "gender", err: errors.New(`ent: missing required field "Character.gender"`)}
	}
	if v, ok := cc.mutation.Gender(); ok {
		if err := character.GenderValidator(v); err != nil {
			return &ValidationError{Name: "gender", err: fmt.Errorf(`ent: validator failed for field "Character.gender": %w`, err)}
		}
	}
	if _, ok := cc.mutation.Race(); !ok {
		return &ValidationError{Name: "race", err: errors.New(`ent: missing required field "Character.race"`)}
	}
	if v, ok := cc.mutation.Race(); ok {
		if err := character.RaceValidator(v); err != nil {
			return &ValidationError{Name: "race", err: fmt.Errorf(`ent: validator failed for field "Character.race": %w`, err)}
		}
	}
	if _, ok := cc.mutation.Flags(); !ok {
		return &ValidationError{Name: "flags", err: errors.New(`ent: missing required field "Character.flags"`)}
	}
	if v, ok := cc.mutation.Flags(); ok {
		if err := character.FlagsValidator(v); err != nil {
			return &ValidationError{Name: "flags", err: fmt.Errorf(`ent: validator failed for field "Character.flags": %w`, err)}
		}
	}
	if _, ok := cc.mutation.Quickslots(); !ok {
		return &ValidationError{Name: "quickslots", err: errors.New(`ent: missing required field "Character.quickslots"`)}
	}
	if v, ok := cc.mutation.Quickslots(); ok {
		if err := character.QuickslotsValidator(v); err != nil {
			return &ValidationError{Name: "quickslots", err: fmt.Errorf(`ent: validator failed for field "Character.quickslots": %w`, err)}
		}
	}
	if _, ok := cc.mutation.Quests(); !ok {
		return &ValidationError{Name: "quests", err: errors.New(`ent: missing required field "Character.quests"`)}
	}
	if v, ok := cc.mutation.Quests(); ok {
		if err := character.QuestsValidator(v); err != nil {
			return &ValidationError{Name: "quests", err: fmt.Errorf(`ent: validator failed for field "Character.quests": %w`, err)}
		}
	}
	if _, ok := cc.mutation.Guild(); !ok {
		return &ValidationError{Name: "guild", err: errors.New(`ent: missing required field "Character.guild"`)}
	}
	if _, ok := cc.mutation.Kills(); !ok {
		return &ValidationError{Name: "kills", err: errors.New(`ent: missing required field "Character.kills"`)}
	}
	if v, ok := cc.mutation.Kills(); ok {
		if err := character.KillsValidator(v); err != nil {
			return &ValidationError{Name: "kills", err: fmt.Errorf(`ent: validator failed for field "Character.kills": %w`, err)}
		}
	}
	if _, ok := cc.mutation.Gold(); !ok {
		return &ValidationError{Name: "gold", err: errors.New(`ent: missing required field "Character.gold"`)}
	}
	if v, ok := cc.mutation.Gold(); ok {
		if err := character.GoldValidator(v); err != nil {
			return &ValidationError{Name: "gold", err: fmt.Errorf(`ent: validator failed for field "Character.gold": %w`, err)}
		}
	}
	if _, ok := cc.mutation.Skills(); !ok {
		return &ValidationError{Name: "skills", err: errors.New(`ent: missing required field "Character.skills"`)}
	}
	if v, ok := cc.mutation.Skills(); ok {
		if err := character.SkillsValidator(v); err != nil {
			return &ValidationError{Name: "skills", err: fmt.Errorf(`ent: validator failed for field "Character.skills": %w`, err)}
		}
	}
	if _, ok := cc.mutation.Pets(); !ok {
		return &ValidationError{Name: "pets", err: errors.New(`ent: missing required field "Character.pets"`)}
	}
	if v, ok := cc.mutation.Pets(); ok {
		if err := character.PetsValidator(v); err != nil {
			return &ValidationError{Name: "pets", err: fmt.Errorf(`ent: validator failed for field "Character.pets": %w`, err)}
		}
	}
	if _, ok := cc.mutation.Health(); !ok {
		return &ValidationError{Name: "health", err: errors.New(`ent: missing required field "Character.health"`)}
	}
	if v, ok := cc.mutation.Health(); ok {
		if err := character.HealthValidator(v); err != nil {
			return &ValidationError{Name: "health", err: fmt.Errorf(`ent: validator failed for field "Character.health": %w`, err)}
		}
	}
	if _, ok := cc.mutation.Mana(); !ok {
		return &ValidationError{Name: "mana", err: errors.New(`ent: missing required field "Character.mana"`)}
	}
	if v, ok := cc.mutation.Mana(); ok {
		if err := character.ManaValidator(v); err != nil {
			return &ValidationError{Name: "mana", err: fmt.Errorf(`ent: validator failed for field "Character.mana": %w`, err)}
		}
	}
	if _, ok := cc.mutation.Equipped(); !ok {
		return &ValidationError{Name: "equipped", err: errors.New(`ent: missing required field "Character.equipped"`)}
	}
	if v, ok := cc.mutation.Equipped(); ok {
		if err := character.EquippedValidator(v); err != nil {
			return &ValidationError{Name: "equipped", err: fmt.Errorf(`ent: validator failed for field "Character.equipped": %w`, err)}
		}
	}
	if _, ok := cc.mutation.Lefthand(); !ok {
		return &ValidationError{Name: "lefthand", err: errors.New(`ent: missing required field "Character.lefthand"`)}
	}
	if _, ok := cc.mutation.Righthand(); !ok {
		return &ValidationError{Name: "righthand", err: errors.New(`ent: missing required field "Character.righthand"`)}
	}
	if _, ok := cc.mutation.Spells(); !ok {
		return &ValidationError{Name: "spells", err: errors.New(`ent: missing required field "Character.spells"`)}
	}
	if v, ok := cc.mutation.Spells(); ok {
		if err := character.SpellsValidator(v); err != nil {
			return &ValidationError{Name: "spells", err: fmt.Errorf(`ent: validator failed for field "Character.spells": %w`, err)}
		}
	}
	if _, ok := cc.mutation.Spellbook(); !ok {
		return &ValidationError{Name: "spellbook", err: errors.New(`ent: missing required field "Character.spellbook"`)}
	}
	if v, ok := cc.mutation.Spellbook(); ok {
		if err := character.SpellbookValidator(v); err != nil {
			return &ValidationError{Name: "spellbook", err: fmt.Errorf(`ent: validator failed for field "Character.spellbook": %w`, err)}
		}
	}
	if _, ok := cc.mutation.Bags(); !ok {
		return &ValidationError{Name: "bags", err: errors.New(`ent: missing required field "Character.bags"`)}
	}
	if v, ok := cc.mutation.Bags(); ok {
		if err := character.BagsValidator(v); err != nil {
			return &ValidationError{Name: "bags", err: fmt.Errorf(`ent: validator failed for field "Character.bags": %w`, err)}
		}
	}
	if _, ok := cc.mutation.Sheaths(); !ok {
		return &ValidationError{Name: "sheaths", err: errors.New(`ent: missing required field "Character.sheaths"`)}
	}
	if v, ok := cc.mutation.Sheaths(); ok {
		if err := character.SheathsValidator(v); err != nil {
			return &ValidationError{Name: "sheaths", err: fmt.Errorf(`ent: validator failed for field "Character.sheaths": %w`, err)}
		}
	}
	return nil
}

func (cc *CharacterCreate) sqlSave(ctx context.Context) (*Character, error) {
	_node, _spec := cc.createSpec()
	if err := sqlgraph.CreateNode(ctx, cc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{err.Error(), err}
		}
		return nil, err
	}
	if _spec.ID.Value != nil {
		if id, ok := _spec.ID.Value.(*uuid.UUID); ok {
			_node.ID = *id
		} else if err := _node.ID.Scan(_spec.ID.Value); err != nil {
			return nil, err
		}
	}
	return _node, nil
}

func (cc *CharacterCreate) createSpec() (*Character, *sqlgraph.CreateSpec) {
	var (
		_node = &Character{config: cc.config}
		_spec = &sqlgraph.CreateSpec{
			Table: character.Table,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeUUID,
				Column: character.FieldID,
			},
		}
	)
	if id, ok := cc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := cc.mutation.Steamid(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: character.FieldSteamid,
		})
		_node.Steamid = value
	}
	if value, ok := cc.mutation.Slot(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeInt,
			Value:  value,
			Column: character.FieldSlot,
		})
		_node.Slot = value
	}
	if value, ok := cc.mutation.Name(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: character.FieldName,
		})
		_node.Name = value
	}
	if value, ok := cc.mutation.Gender(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeInt,
			Value:  value,
			Column: character.FieldGender,
		})
		_node.Gender = value
	}
	if value, ok := cc.mutation.Race(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeInt,
			Value:  value,
			Column: character.FieldRace,
		})
		_node.Race = value
	}
	if value, ok := cc.mutation.Flags(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: character.FieldFlags,
		})
		_node.Flags = value
	}
	if value, ok := cc.mutation.Quickslots(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: character.FieldQuickslots,
		})
		_node.Quickslots = value
	}
	if value, ok := cc.mutation.Quests(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: character.FieldQuests,
		})
		_node.Quests = value
	}
	if value, ok := cc.mutation.Guild(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: character.FieldGuild,
		})
		_node.Guild = value
	}
	if value, ok := cc.mutation.Kills(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeInt,
			Value:  value,
			Column: character.FieldKills,
		})
		_node.Kills = value
	}
	if value, ok := cc.mutation.Gold(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeInt,
			Value:  value,
			Column: character.FieldGold,
		})
		_node.Gold = value
	}
	if value, ok := cc.mutation.Skills(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: character.FieldSkills,
		})
		_node.Skills = value
	}
	if value, ok := cc.mutation.Pets(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: character.FieldPets,
		})
		_node.Pets = value
	}
	if value, ok := cc.mutation.Health(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeInt,
			Value:  value,
			Column: character.FieldHealth,
		})
		_node.Health = value
	}
	if value, ok := cc.mutation.Mana(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeInt,
			Value:  value,
			Column: character.FieldMana,
		})
		_node.Mana = value
	}
	if value, ok := cc.mutation.Equipped(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: character.FieldEquipped,
		})
		_node.Equipped = value
	}
	if value, ok := cc.mutation.Lefthand(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: character.FieldLefthand,
		})
		_node.Lefthand = value
	}
	if value, ok := cc.mutation.Righthand(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: character.FieldRighthand,
		})
		_node.Righthand = value
	}
	if value, ok := cc.mutation.Spells(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: character.FieldSpells,
		})
		_node.Spells = value
	}
	if value, ok := cc.mutation.Spellbook(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: character.FieldSpellbook,
		})
		_node.Spellbook = value
	}
	if value, ok := cc.mutation.Bags(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: character.FieldBags,
		})
		_node.Bags = value
	}
	if value, ok := cc.mutation.Sheaths(); ok {
		_spec.Fields = append(_spec.Fields, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: character.FieldSheaths,
		})
		_node.Sheaths = value
	}
	return _node, _spec
}

// CharacterCreateBulk is the builder for creating many Character entities in bulk.
type CharacterCreateBulk struct {
	config
	builders []*CharacterCreate
}

// Save creates the Character entities in the database.
func (ccb *CharacterCreateBulk) Save(ctx context.Context) ([]*Character, error) {
	specs := make([]*sqlgraph.CreateSpec, len(ccb.builders))
	nodes := make([]*Character, len(ccb.builders))
	mutators := make([]Mutator, len(ccb.builders))
	for i := range ccb.builders {
		func(i int, root context.Context) {
			builder := ccb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*CharacterMutation)
				if !ok {
					return nil, fmt.Errorf("unexpected mutation type %T", m)
				}
				if err := builder.check(); err != nil {
					return nil, err
				}
				builder.mutation = mutation
				nodes[i], specs[i] = builder.createSpec()
				var err error
				if i < len(mutators)-1 {
					_, err = mutators[i+1].Mutate(root, ccb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, ccb.driver, spec); err != nil {
						if sqlgraph.IsConstraintError(err) {
							err = &ConstraintError{err.Error(), err}
						}
					}
				}
				if err != nil {
					return nil, err
				}
				mutation.id = &nodes[i].ID
				mutation.done = true
				return nodes[i], nil
			})
			for i := len(builder.hooks) - 1; i >= 0; i-- {
				mut = builder.hooks[i](mut)
			}
			mutators[i] = mut
		}(i, ctx)
	}
	if len(mutators) > 0 {
		if _, err := mutators[0].Mutate(ctx, ccb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (ccb *CharacterCreateBulk) SaveX(ctx context.Context) []*Character {
	v, err := ccb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (ccb *CharacterCreateBulk) Exec(ctx context.Context) error {
	_, err := ccb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ccb *CharacterCreateBulk) ExecX(ctx context.Context) {
	if err := ccb.Exec(ctx); err != nil {
		panic(err)
	}
}
