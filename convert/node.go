package convert

import (
	"fmt"

	"github.com/coder/gots/bindings"
)

// TypescriptNode is any type that can be serialized into typescript
type TypescriptNode struct {
	Node any
	// mutations is a list of functions that need to be applied to the node before
	// it can be serialized to typescript. It exists for ensuring consistent ordering
	// of execution, regardless of the parsing order.
	// These mutations can be anything.
	mutations []func(v any) (any, error)
}

func (t TypescriptNode) applyMutations() (TypescriptNode, error) {
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

func (t TypescriptNode) Typescript(vm *bindings.Bindings) (string, error) {
	obj, err := vm.ToTypescriptNode(t.Node)
	if err != nil {
		return "", fmt.Errorf("convert node: %w", err)
	}

	typescript, err := vm.SerializeToTypescript(obj)
	if err != nil {
		return "", fmt.Errorf("serialize to typescript: %w", err)
	}
	return typescript, nil
}

func (t *TypescriptNode) AddEnum(enum bindings.ExpressionType) {
	t.mutations = append(t.mutations, func(v any) (any, error) {
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
