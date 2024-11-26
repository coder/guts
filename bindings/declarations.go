package bindings

import (
	"fmt"
	"reflect"
)

// DeclarationType is any type that can exist at the top level of a AST.
// Meaning it can be serialized into valid Typescript.
type DeclarationType interface {
	isDeclarationType()
	Node
}

type Interface struct {
	Name       Identifier
	Modifiers  []Modifier
	Fields     []*PropertySignature
	Parameters []*TypeParameter
	Heritage   []*HeritageClause
	Source

	isTypescriptNode
}

func (*Interface) isDeclarationType() {}

type PropertySignature struct {
	// Name is the field name
	Name          string
	Modifiers     []Modifier
	QuestionToken bool
	Type          ExpressionType
	// FieldComments maybe should be its own AST node?
	FieldComments []string

	isTypescriptNode
}

type Alias struct {
	Name       Identifier
	Modifiers  []Modifier
	Type       ExpressionType
	Parameters []*TypeParameter
	Source

	isTypescriptNode
}

func (*Alias) isDeclarationType() {}

// TypeParameter are generics in Go
// Foo[T comparable] ->
// - name: T
// - Modifiers: []
// - Type: Comparable
// - DefaultType: nil
type TypeParameter struct {
	Name      string
	Modifiers []Modifier
	Type      ExpressionType
	// DefaultType does not map to any Golang concepts and will never be
	// used.
	DefaultType ExpressionType

	isTypescriptNode
}

// Simplify removes duplicate type parameters
func Simplify(p []*TypeParameter) ([]*TypeParameter, error) {
	params := make([]*TypeParameter, 0, len(p))
	exists := make(map[string]*TypeParameter)
	for _, tp := range p {
		if found, ok := exists[tp.Name]; ok {
			// Compare types, make sure they are the same
			equal := reflect.DeepEqual(found, tp)
			if !equal {
				return nil, fmt.Errorf("type parameter %q already exists with different type", tp.Name)
			}
			continue
		}
		params = append(params, tp)
		exists[tp.Name] = tp
	}
	return params, nil
}

type VariableStatement struct {
	Modifiers    []Modifier
	Declarations *VariableDeclarationList
	Source

	isTypescriptNode
}

func (*VariableStatement) isDeclarationType() {}
