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
	// Comments maybe should be its own AST node?
	Comments []string
	Source
}

func (*Interface) isNode()            {}
func (*Interface) isDeclarationType() {}

// PropertySignature is a field in an interface
type PropertySignature struct {
	// Name is the field name
	Name          string
	Modifiers     []Modifier
	QuestionToken bool
	Type          ExpressionType
	// FieldComments maybe should be its own AST node?
	FieldComments []string
}

func (*PropertySignature) isNode() {}

type Alias struct {
	Name       Identifier
	Modifiers  []Modifier
	Type       ExpressionType
	Parameters []*TypeParameter
	Source
}

func (*Alias) isNode()            {}
func (*Alias) isDeclarationType() {}

// TypeParameter are generics in Go
// Foo[T comparable] ->
// - name: T
// - Modifiers: []
// - Type: Comparable
// - DefaultType: nil
type TypeParameter struct {
	Name      Identifier
	Modifiers []Modifier
	Type      ExpressionType
	// DefaultType does not map to any Golang concepts and will never be
	// used.
	DefaultType ExpressionType
}

func (p *TypeParameter) isNode() {}

// Simplify removes duplicate type parameters
func Simplify(p []*TypeParameter) ([]*TypeParameter, error) {
	params := make([]*TypeParameter, 0, len(p))
	exists := make(map[string]*TypeParameter)
	for _, tp := range p {
		if found, ok := exists[tp.Name.Ref()]; ok {
			// Compare types, make sure they are the same
			equal := reflect.DeepEqual(found, tp)
			if !equal {
				return nil, fmt.Errorf("type parameter %q already exists with different type", tp.Name)
			}
			continue
		}
		params = append(params, tp)
		exists[tp.Name.Ref()] = tp
	}
	return params, nil
}

// VariableStatement is a top level declaration of a variable
// var foo: string = "bar"
// const foo: string = "bar"
type VariableStatement struct {
	Modifiers    []Modifier
	Declarations *VariableDeclarationList
	Source
}

func (*VariableStatement) isNode()            {}
func (*VariableStatement) isDeclarationType() {}
