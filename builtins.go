package guts

import "github.com/coder/guts/bindings"

var (
	builtInComparable = bindings.Identifier{Name: "Comparable"}
	builtInString     = bindings.Identifier{Name: "string"}
	builtInNumber     = bindings.Identifier{Name: "number"}
	builtInBoolean    = bindings.Identifier{Name: "boolean"}
	builtInRecord     = bindings.Identifier{Name: "Record"}
)

func (ts *Typescript) includeComparable() {
	// The zzz just pushes it to the end of the sorting.
	// Kinda strange, but it works.
	_ = ts.setNode(builtInComparable.Ref(), typescriptNode{
		Node: &bindings.Alias{
			Name:      builtInComparable,
			Modifiers: []bindings.Modifier{},
			Type: bindings.Union(
				bindings.Reference(builtInString),
				bindings.Reference(builtInNumber),
				bindings.Reference(builtInBoolean),
			),
			Parameters: []*bindings.TypeParameter{},
			Source:     bindings.Source{},
		},
	})
}
