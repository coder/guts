package guts

import (
	"fmt"

	"github.com/coder/guts/bindings"
)

type typescriptNode struct {
	Node bindings.Node
	// mutations is a list of functions that need to be applied to the node before
	// it can be serialized to typescript. It exists for ensuring consistent ordering
	// of execution, regardless of the parsing order.
	// These mutations can be anything.
	mutations []func(v bindings.Node) (bindings.Node, error)
}

func (t typescriptNode) applyMutations() (typescriptNode, error) {
	for i, m := range t.mutations {
		var err error
		t.Node, err = m(t.Node)
		if err != nil {
			return t, fmt.Errorf("apply mutation %d: %w", i, err)
		}
	}
	t.mutations = nil
	return t, nil
}

func (t *typescriptNode) AddEnum(member *bindings.EnumMember) {
	t.mutations = append(t.mutations, func(v bindings.Node) (bindings.Node, error) {
		alias, ok := v.(*bindings.Alias)
		if ok {
			// Switch to an enum
			enum := &bindings.Enum{
				Name:      alias.Name,
				Modifiers: alias.Modifiers,
				Members:   []*bindings.EnumMember{member},
				Source:    alias.Source,
			}
			return enum, nil
		}

		enum, ok := v.(*bindings.Enum)
		if !ok {
			return v, fmt.Errorf("expected enum, got %T", v)
		}

		enum.Members = append(enum.Members, member)
		return enum, nil
	})
}
