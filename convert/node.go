package convert

import (
	"fmt"

	"github.com/coder/gots/bindings"
)

// TypescriptNode is any type that can be serialized into typescript
type TypescriptNode struct {
	Node any
	// Mutations is a list of functions that need to be applied to the node before
	// it can be serialized to typescript.
	// These mutations can be anything.
	Mutations []func(v any) (any, error)
}

func (t TypescriptNode) Typescript(vm *bindings.Bindings) (string, error) {
	node := t.Node
	var err error
	for i, mut := range t.Mutations {
		node, err = mut(node)
		if err != nil {
			return "", fmt.Errorf("mutation %d: %w", i, err)
		}
	}

	obj, err := vm.ToTypescriptNode(node)
	if err != nil {
		return "", fmt.Errorf("convert node: %w", err)
	}

	typescript, err := vm.SerializeToTypescript(obj)
	if err != nil {
		return "", fmt.Errorf("serialize to typescript: %w", err)
	}
	return typescript, nil
}

func (t TypescriptNode) AddEnum(enum bindings.ExpressionType) TypescriptNode {
	t.Mutations = append(t.Mutations, func(v any) (any, error) {
		alias, ok := v.(bindings.Alias)
		if !ok {
			return nil, fmt.Errorf("expected alias type, got %T", t.Node)
		}

		union, ok := alias.Type.(bindings.UnionType)
		if !ok {
			// Make it a union, this removes the original type.
			union = bindings.Union()
			alias.Type = union
		}

		union.Types = append(union.Types, enum)
		alias.Type = union
		return alias, nil
	})
	return t
}
