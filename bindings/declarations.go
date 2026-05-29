package bindings

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
	SupportComments
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
	SupportComments
}

func (*PropertySignature) isNode() {}

type Alias struct {
	Name       Identifier
	Modifiers  []Modifier
	Type       ExpressionType
	Parameters []*TypeParameter
	SupportComments
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
	params := []*TypeParameter{}
	set := make(map[string]bool)
	for _, tp := range p {
		ref := tp.Name.Ref()
		if _, ok := set[ref]; !ok {
			params = append(params, tp)
			set[ref] = true

			if union, ok := tp.Type.(*UnionType); ok {
				simplifyUnionLiterals(union)
			}
		}
	}
	return params, nil
}

func simplifyUnionLiterals(union *UnionType) *UnionType {
	types := []ExpressionType{}
	literalSet := map[string]bool{}
	for _, arg := range union.Types {
		switch v := arg.(type) {
		case *LiteralKeyword:
			key := v.String()
			if _, ok := literalSet[key]; !ok {
				literalSet[key] = true
				types = append(types, arg)
			}
		default:
			types = append(types, arg)
		}
	}
	union.Types = types
	return union
}

// VariableStatement is a top level declaration of a variable
// var foo: string = "bar"
// const foo: string = "bar"
type VariableStatement struct {
	Modifiers    []Modifier
	Declarations *VariableDeclarationList
	Source
	SupportComments
}

func (*VariableStatement) isNode()            {}
func (*VariableStatement) isDeclarationType() {}

type Enum struct {
	Name      Identifier
	Modifiers []Modifier
	Members   []*EnumMember
	SupportComments
	Source
}

func (*Enum) isNode()            {}
func (*Enum) isDeclarationType() {}

// ImportDeclaration is a top-level ECMAScript import statement.
//
//	import { z } from "zod"
//	import { Foo as Bar } from "./schemas"
//	import type { Baz } from "./types"
//
// Only the named-imports form is modeled. Default imports and namespace
// imports (`import * as ns from "x"`) are not yet supported; add them if
// you need them.
type ImportDeclaration struct {
	// Module is the module specifier (the right-hand side of `from`).
	Module string
	// Named is the list of imported names. Order is preserved in the
	// generated output.
	Named []*ImportSpecifier
	// IsTypeOnly emits `import type { ... }` instead of `import { ... }`.
	IsTypeOnly bool
	SupportComments
	Source
}

func (*ImportDeclaration) isNode()            {}
func (*ImportDeclaration) isDeclarationType() {}

// ImportSpecifier is a single named import entry inside the braces of an
// `import { ... }` clause.
//
//	Name="foo", Alias=""     ->  foo
//	Name="foo", Alias="bar"  ->  foo as bar
//	IsTypeOnly=true          ->  type foo, type foo as bar
type ImportSpecifier struct {
	Name       string
	Alias      string
	IsTypeOnly bool
}

func (*ImportSpecifier) isNode() {}

