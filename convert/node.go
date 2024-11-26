package convert

import (
	"fmt"

	"github.com/coder/gots/bindings"
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
	for _, m := range t.mutations {
		var err error
		t.Node, err = m(t.Node)
		if err != nil {
			return t, fmt.Errorf("apply mutation: %w", err)
		}
	}
	t.mutations = nil
	return t, nil
}

func (t *typescriptNode) AddEnum(enum bindings.ExpressionType) {
	t.mutations = append(t.mutations, func(v bindings.Node) (bindings.Node, error) {
		alias, ok := v.(*bindings.Alias)
		if !ok {
			return nil, fmt.Errorf("expected alias type, got %T", t.Node)
		}

		union, ok := alias.Type.(*bindings.UnionType)
		if !ok {
			// Make it a union, this removes the original type.
			union = bindings.Union()
			alias.Type = union
		}

		union.Types = append(union.Types, enum)
		alias.Type = union
		return alias, nil
	})
}
