package gots

import "github.com/coder/gots/bindings"

const (
	builtInComparable = "Comparable"
)

func (ts *Typescript) includeComparable() {
	// The zzz just pushes it to the end of the sorting.
	// Kinda strange, but it works.
	_ = ts.setNode(builtInComparable, typescriptNode{
		Node: &bindings.Alias{
			Name:      builtInComparable,
			Modifiers: []bindings.Modifier{},
			Type: bindings.Union(
				bindings.Reference("string"),
				bindings.Reference("number"),
				bindings.Reference("boolean"),
			),
			Parameters: []*bindings.TypeParameter{},
			Source:     bindings.Source{},
		},
	})
}
