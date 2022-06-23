// Code generated by entc, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/msrevive/nexus2/ent/deprecatedcharacter"
	"github.com/msrevive/nexus2/ent/predicate"
)

// DeprecatedCharacterUpdate is the builder for updating DeprecatedCharacter entities.
type DeprecatedCharacterUpdate struct {
	config
	hooks    []Hook
	mutation *DeprecatedCharacterMutation
}

// Where appends a list predicates to the DeprecatedCharacterUpdate builder.
func (dcu *DeprecatedCharacterUpdate) Where(ps ...predicate.DeprecatedCharacter) *DeprecatedCharacterUpdate {
	dcu.mutation.Where(ps...)
	return dcu
}

// SetSteamid sets the "steamid" field.
func (dcu *DeprecatedCharacterUpdate) SetSteamid(s string) *DeprecatedCharacterUpdate {
	dcu.mutation.SetSteamid(s)
	return dcu
}

// SetSlot sets the "slot" field.
func (dcu *DeprecatedCharacterUpdate) SetSlot(i int) *DeprecatedCharacterUpdate {
	dcu.mutation.ResetSlot()
	dcu.mutation.SetSlot(i)
	return dcu
}

// SetNillableSlot sets the "slot" field if the given value is not nil.
func (dcu *DeprecatedCharacterUpdate) SetNillableSlot(i *int) *DeprecatedCharacterUpdate {
	if i != nil {
		dcu.SetSlot(*i)
	}
	return dcu
}

// AddSlot adds i to the "slot" field.
func (dcu *DeprecatedCharacterUpdate) AddSlot(i int) *DeprecatedCharacterUpdate {
	dcu.mutation.AddSlot(i)
	return dcu
}

// SetSize sets the "size" field.
func (dcu *DeprecatedCharacterUpdate) SetSize(i int) *DeprecatedCharacterUpdate {
	dcu.mutation.ResetSize()
	dcu.mutation.SetSize(i)
	return dcu
}

// SetNillableSize sets the "size" field if the given value is not nil.
func (dcu *DeprecatedCharacterUpdate) SetNillableSize(i *int) *DeprecatedCharacterUpdate {
	if i != nil {
		dcu.SetSize(*i)
	}
	return dcu
}

// AddSize adds i to the "size" field.
func (dcu *DeprecatedCharacterUpdate) AddSize(i int) *DeprecatedCharacterUpdate {
	dcu.mutation.AddSize(i)
	return dcu
}

// SetData sets the "data" field.
func (dcu *DeprecatedCharacterUpdate) SetData(s string) *DeprecatedCharacterUpdate {
	dcu.mutation.SetData(s)
	return dcu
}

// Mutation returns the DeprecatedCharacterMutation object of the builder.
func (dcu *DeprecatedCharacterUpdate) Mutation() *DeprecatedCharacterMutation {
	return dcu.mutation
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (dcu *DeprecatedCharacterUpdate) Save(ctx context.Context) (int, error) {
	var (
		err      error
		affected int
	)
	if len(dcu.hooks) == 0 {
		if err = dcu.check(); err != nil {
			return 0, err
		}
		affected, err = dcu.sqlSave(ctx)
	} else {
		var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
			mutation, ok := m.(*DeprecatedCharacterMutation)
			if !ok {
				return nil, fmt.Errorf("unexpected mutation type %T", m)
			}
			if err = dcu.check(); err != nil {
				return 0, err
			}
			dcu.mutation = mutation
			affected, err = dcu.sqlSave(ctx)
			mutation.done = true
			return affected, err
		})
		for i := len(dcu.hooks) - 1; i >= 0; i-- {
			if dcu.hooks[i] == nil {
				return 0, fmt.Errorf("ent: uninitialized hook (forgotten import ent/runtime?)")
			}
			mut = dcu.hooks[i](mut)
		}
		if _, err := mut.Mutate(ctx, dcu.mutation); err != nil {
			return 0, err
		}
	}
	return affected, err
}

// SaveX is like Save, but panics if an error occurs.
func (dcu *DeprecatedCharacterUpdate) SaveX(ctx context.Context) int {
	affected, err := dcu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (dcu *DeprecatedCharacterUpdate) Exec(ctx context.Context) error {
	_, err := dcu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (dcu *DeprecatedCharacterUpdate) ExecX(ctx context.Context) {
	if err := dcu.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (dcu *DeprecatedCharacterUpdate) check() error {
	if v, ok := dcu.mutation.Slot(); ok {
		if err := deprecatedcharacter.SlotValidator(v); err != nil {
			return &ValidationError{Name: "slot", err: fmt.Errorf(`ent: validator failed for field "DeprecatedCharacter.slot": %w`, err)}
		}
	}
	return nil
}

func (dcu *DeprecatedCharacterUpdate) sqlSave(ctx context.Context) (n int, err error) {
	_spec := &sqlgraph.UpdateSpec{
		Node: &sqlgraph.NodeSpec{
			Table:   deprecatedcharacter.Table,
			Columns: deprecatedcharacter.Columns,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeUUID,
				Column: deprecatedcharacter.FieldID,
			},
		},
	}
	if ps := dcu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := dcu.mutation.Steamid(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: deprecatedcharacter.FieldSteamid,
		})
	}
	if value, ok := dcu.mutation.Slot(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeInt,
			Value:  value,
			Column: deprecatedcharacter.FieldSlot,
		})
	}
	if value, ok := dcu.mutation.AddedSlot(); ok {
		_spec.Fields.Add = append(_spec.Fields.Add, &sqlgraph.FieldSpec{
			Type:   field.TypeInt,
			Value:  value,
			Column: deprecatedcharacter.FieldSlot,
		})
	}
	if value, ok := dcu.mutation.Size(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeInt,
			Value:  value,
			Column: deprecatedcharacter.FieldSize,
		})
	}
	if value, ok := dcu.mutation.AddedSize(); ok {
		_spec.Fields.Add = append(_spec.Fields.Add, &sqlgraph.FieldSpec{
			Type:   field.TypeInt,
			Value:  value,
			Column: deprecatedcharacter.FieldSize,
		})
	}
	if value, ok := dcu.mutation.Data(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: deprecatedcharacter.FieldData,
		})
	}
	if n, err = sqlgraph.UpdateNodes(ctx, dcu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{deprecatedcharacter.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{err.Error(), err}
		}
		return 0, err
	}
	return n, nil
}

// DeprecatedCharacterUpdateOne is the builder for updating a single DeprecatedCharacter entity.
type DeprecatedCharacterUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *DeprecatedCharacterMutation
}

// SetSteamid sets the "steamid" field.
func (dcuo *DeprecatedCharacterUpdateOne) SetSteamid(s string) *DeprecatedCharacterUpdateOne {
	dcuo.mutation.SetSteamid(s)
	return dcuo
}

// SetSlot sets the "slot" field.
func (dcuo *DeprecatedCharacterUpdateOne) SetSlot(i int) *DeprecatedCharacterUpdateOne {
	dcuo.mutation.ResetSlot()
	dcuo.mutation.SetSlot(i)
	return dcuo
}

// SetNillableSlot sets the "slot" field if the given value is not nil.
func (dcuo *DeprecatedCharacterUpdateOne) SetNillableSlot(i *int) *DeprecatedCharacterUpdateOne {
	if i != nil {
		dcuo.SetSlot(*i)
	}
	return dcuo
}

// AddSlot adds i to the "slot" field.
func (dcuo *DeprecatedCharacterUpdateOne) AddSlot(i int) *DeprecatedCharacterUpdateOne {
	dcuo.mutation.AddSlot(i)
	return dcuo
}

// SetSize sets the "size" field.
func (dcuo *DeprecatedCharacterUpdateOne) SetSize(i int) *DeprecatedCharacterUpdateOne {
	dcuo.mutation.ResetSize()
	dcuo.mutation.SetSize(i)
	return dcuo
}

// SetNillableSize sets the "size" field if the given value is not nil.
func (dcuo *DeprecatedCharacterUpdateOne) SetNillableSize(i *int) *DeprecatedCharacterUpdateOne {
	if i != nil {
		dcuo.SetSize(*i)
	}
	return dcuo
}

// AddSize adds i to the "size" field.
func (dcuo *DeprecatedCharacterUpdateOne) AddSize(i int) *DeprecatedCharacterUpdateOne {
	dcuo.mutation.AddSize(i)
	return dcuo
}

// SetData sets the "data" field.
func (dcuo *DeprecatedCharacterUpdateOne) SetData(s string) *DeprecatedCharacterUpdateOne {
	dcuo.mutation.SetData(s)
	return dcuo
}

// Mutation returns the DeprecatedCharacterMutation object of the builder.
func (dcuo *DeprecatedCharacterUpdateOne) Mutation() *DeprecatedCharacterMutation {
	return dcuo.mutation
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (dcuo *DeprecatedCharacterUpdateOne) Select(field string, fields ...string) *DeprecatedCharacterUpdateOne {
	dcuo.fields = append([]string{field}, fields...)
	return dcuo
}

// Save executes the query and returns the updated DeprecatedCharacter entity.
func (dcuo *DeprecatedCharacterUpdateOne) Save(ctx context.Context) (*DeprecatedCharacter, error) {
	var (
		err  error
		node *DeprecatedCharacter
	)
	if len(dcuo.hooks) == 0 {
		if err = dcuo.check(); err != nil {
			return nil, err
		}
		node, err = dcuo.sqlSave(ctx)
	} else {
		var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
			mutation, ok := m.(*DeprecatedCharacterMutation)
			if !ok {
				return nil, fmt.Errorf("unexpected mutation type %T", m)
			}
			if err = dcuo.check(); err != nil {
				return nil, err
			}
			dcuo.mutation = mutation
			node, err = dcuo.sqlSave(ctx)
			mutation.done = true
			return node, err
		})
		for i := len(dcuo.hooks) - 1; i >= 0; i-- {
			if dcuo.hooks[i] == nil {
				return nil, fmt.Errorf("ent: uninitialized hook (forgotten import ent/runtime?)")
			}
			mut = dcuo.hooks[i](mut)
		}
		if _, err := mut.Mutate(ctx, dcuo.mutation); err != nil {
			return nil, err
		}
	}
	return node, err
}

// SaveX is like Save, but panics if an error occurs.
func (dcuo *DeprecatedCharacterUpdateOne) SaveX(ctx context.Context) *DeprecatedCharacter {
	node, err := dcuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (dcuo *DeprecatedCharacterUpdateOne) Exec(ctx context.Context) error {
	_, err := dcuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (dcuo *DeprecatedCharacterUpdateOne) ExecX(ctx context.Context) {
	if err := dcuo.Exec(ctx); err != nil {
		panic(err)
	}
}

// check runs all checks and user-defined validators on the builder.
func (dcuo *DeprecatedCharacterUpdateOne) check() error {
	if v, ok := dcuo.mutation.Slot(); ok {
		if err := deprecatedcharacter.SlotValidator(v); err != nil {
			return &ValidationError{Name: "slot", err: fmt.Errorf(`ent: validator failed for field "DeprecatedCharacter.slot": %w`, err)}
		}
	}
	return nil
}

func (dcuo *DeprecatedCharacterUpdateOne) sqlSave(ctx context.Context) (_node *DeprecatedCharacter, err error) {
	_spec := &sqlgraph.UpdateSpec{
		Node: &sqlgraph.NodeSpec{
			Table:   deprecatedcharacter.Table,
			Columns: deprecatedcharacter.Columns,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeUUID,
				Column: deprecatedcharacter.FieldID,
			},
		},
	}
	id, ok := dcuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "DeprecatedCharacter.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := dcuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, deprecatedcharacter.FieldID)
		for _, f := range fields {
			if !deprecatedcharacter.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != deprecatedcharacter.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := dcuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := dcuo.mutation.Steamid(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: deprecatedcharacter.FieldSteamid,
		})
	}
	if value, ok := dcuo.mutation.Slot(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeInt,
			Value:  value,
			Column: deprecatedcharacter.FieldSlot,
		})
	}
	if value, ok := dcuo.mutation.AddedSlot(); ok {
		_spec.Fields.Add = append(_spec.Fields.Add, &sqlgraph.FieldSpec{
			Type:   field.TypeInt,
			Value:  value,
			Column: deprecatedcharacter.FieldSlot,
		})
	}
	if value, ok := dcuo.mutation.Size(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeInt,
			Value:  value,
			Column: deprecatedcharacter.FieldSize,
		})
	}
	if value, ok := dcuo.mutation.AddedSize(); ok {
		_spec.Fields.Add = append(_spec.Fields.Add, &sqlgraph.FieldSpec{
			Type:   field.TypeInt,
			Value:  value,
			Column: deprecatedcharacter.FieldSize,
		})
	}
	if value, ok := dcuo.mutation.Data(); ok {
		_spec.Fields.Set = append(_spec.Fields.Set, &sqlgraph.FieldSpec{
			Type:   field.TypeString,
			Value:  value,
			Column: deprecatedcharacter.FieldData,
		})
	}
	_node = &DeprecatedCharacter{config: dcuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, dcuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{deprecatedcharacter.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{err.Error(), err}
		}
		return nil, err
	}
	return _node, nil
}
