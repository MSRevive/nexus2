// Code generated by entc, DO NOT EDIT.

package ent

import (
	"context"
	"fmt"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/msrevive/nexus2/ent/deprecatedcharacter"
	"github.com/msrevive/nexus2/ent/predicate"
)

// DeprecatedCharacterDelete is the builder for deleting a DeprecatedCharacter entity.
type DeprecatedCharacterDelete struct {
	config
	hooks    []Hook
	mutation *DeprecatedCharacterMutation
}

// Where appends a list predicates to the DeprecatedCharacterDelete builder.
func (dcd *DeprecatedCharacterDelete) Where(ps ...predicate.DeprecatedCharacter) *DeprecatedCharacterDelete {
	dcd.mutation.Where(ps...)
	return dcd
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (dcd *DeprecatedCharacterDelete) Exec(ctx context.Context) (int, error) {
	var (
		err      error
		affected int
	)
	if len(dcd.hooks) == 0 {
		affected, err = dcd.sqlExec(ctx)
	} else {
		var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
			mutation, ok := m.(*DeprecatedCharacterMutation)
			if !ok {
				return nil, fmt.Errorf("unexpected mutation type %T", m)
			}
			dcd.mutation = mutation
			affected, err = dcd.sqlExec(ctx)
			mutation.done = true
			return affected, err
		})
		for i := len(dcd.hooks) - 1; i >= 0; i-- {
			if dcd.hooks[i] == nil {
				return 0, fmt.Errorf("ent: uninitialized hook (forgotten import ent/runtime?)")
			}
			mut = dcd.hooks[i](mut)
		}
		if _, err := mut.Mutate(ctx, dcd.mutation); err != nil {
			return 0, err
		}
	}
	return affected, err
}

// ExecX is like Exec, but panics if an error occurs.
func (dcd *DeprecatedCharacterDelete) ExecX(ctx context.Context) int {
	n, err := dcd.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (dcd *DeprecatedCharacterDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := &sqlgraph.DeleteSpec{
		Node: &sqlgraph.NodeSpec{
			Table: deprecatedcharacter.Table,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeUUID,
				Column: deprecatedcharacter.FieldID,
			},
		},
	}
	if ps := dcd.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return sqlgraph.DeleteNodes(ctx, dcd.driver, _spec)
}

// DeprecatedCharacterDeleteOne is the builder for deleting a single DeprecatedCharacter entity.
type DeprecatedCharacterDeleteOne struct {
	dcd *DeprecatedCharacterDelete
}

// Exec executes the deletion query.
func (dcdo *DeprecatedCharacterDeleteOne) Exec(ctx context.Context) error {
	n, err := dcdo.dcd.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{deprecatedcharacter.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (dcdo *DeprecatedCharacterDeleteOne) ExecX(ctx context.Context) {
	dcdo.dcd.ExecX(ctx)
}
