package guts

import "github.com/coder/guts/bindings"

// Some references are built into either Golang or Typescript.
var (
	// builtInComparable is a reference to the 'comparable' type in Golang.
	builtInComparable = bindings.Identifier{Name: "Comparable"}
	// builtInRecord is a reference to the 'Record' type in Typescript.
	builtInRecord = bindings.Identifier{Name: "Record"}
)

// RecordReference creates a reference to the 'Record' type in Typescript.
// The Record type takes in 2 type parameters, key and value.
func RecordReference(key, value bindings.ExpressionType) *bindings.ReferenceType {
	return bindings.Reference(builtInRecord, key, value)
}

func (ts *Typescript) includeComparable() {
	// The zzz just pushes it to the end of the sorting.
	// Kinda strange, but it works.
	_ = ts.setNode(builtInComparable.Ref(), typescriptNode{
		Node: &bindings.Alias{
			Name:      builtInComparable,
			Modifiers: []bindings.Modifier{},
			Type: bindings.Union(
				ptr(bindings.KeywordString),
				ptr(bindings.KeywordNumber),
				ptr(bindings.KeywordBoolean),
			),
			Parameters: []*bindings.TypeParameter{},
			Source:     bindings.Source{},
		},
	})
}
