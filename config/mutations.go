package config

import (
	"fmt"
	"log/slog"
	"reflect"

	"github.com/coder/guts"
	"github.com/coder/guts/bindings"
	"github.com/coder/guts/bindings/walk"
)

func ExportTypes(ts *guts.Typescript) {
	ts.ForEach(func(key string, node bindings.Node) {
		switch node := node.(type) {
		case *bindings.Alias:
			node.Modifiers = append(node.Modifiers, bindings.ModifierExport)
		case *bindings.Interface:
			node.Modifiers = append(node.Modifiers, bindings.ModifierExport)
		case *bindings.VariableStatement:
			node.Modifiers = append(node.Modifiers, bindings.ModifierExport)
		default:
			panic("unexpected node type for exporting")
		}
	})
}

func ReadOnly(ts *guts.Typescript) {
	ts.ForEach(func(key string, node bindings.Node) {
		switch node := node.(type) {
		case *bindings.Alias:
		case *bindings.Interface:
			for _, prop := range node.Fields {
				prop.Modifiers = append(prop.Modifiers, bindings.ModifierReadonly)
			}
		case *bindings.VariableStatement:
		default:
			panic("unexpected node type for exporting")
		}
	})
}

// EnumLists adds a constant that lists all the values in a given enum.
// Example:
// type MyEnum = string
// const (
// EnumFoo = "foo"
// EnumBar = "bar"
// )
// const MyEnums: string = ["foo", "bar"] <-- this is added
func EnumLists(ts *guts.Typescript) {
	addNodes := make(map[string]bindings.Node)
	ts.ForEach(func(key string, node bindings.Node) {
		switch node := node.(type) {
		// Find the enums, and make a list of values.
		// Only support primitive types.
		case *bindings.Alias:
			if union, ok := node.Type.(*bindings.UnionType); ok {
				if len(union.Types) == 0 {
					return
				}

				var expectedType *bindings.LiteralType
				// This might be a union type, if all elements are the same literal type.
				for _, t := range union.Types {
					value, ok := t.(*bindings.LiteralType)
					if !ok {
						return
					}
					if expectedType == nil {
						expectedType = value
						continue
					}

					if reflect.TypeOf(expectedType.Value) != reflect.TypeOf(value.Value) {
						return
					}
				}

				values := make([]bindings.ExpressionType, 0, len(union.Types))
				for _, t := range union.Types {
					values = append(values, t)
				}
				name := key + "s"
				addNodes[name] = &bindings.VariableStatement{
					Modifiers: []bindings.Modifier{},
					Declarations: &bindings.VariableDeclarationList{
						Declarations: []*bindings.VariableDeclaration{
							{
								// TODO: Fix this with Identifier's instead of "string"
								Name:            bindings.Identifier{Name: name},
								ExclamationMark: false,
								Type: &bindings.ArrayType{
									// The type is the enum type
									Node: bindings.Reference(bindings.Identifier{Name: key}),
								},
								Initializer: &bindings.ArrayLiteralType{
									Elements: values,
								},
							},
						},
						Flags: bindings.NodeFlagsConstant,
					},
					Source: bindings.Source{},
				}
			}
		}
	})

	for name, node := range addNodes {
		err := ts.SetNode(name, node)
		if err != nil {
			slog.Error(fmt.Sprintf("failed to add enum list %s: %v", name, err))
		}
	}
}

// MissingReferencesToAny will change any references to types that are not found in the
// typescript tree to 'any'.
// These can be resolved by adding generation for the missing types.
func MissingReferencesToAny(ts *guts.Typescript) {
	// Find all valid references to types
	valid := make(map[string]struct{})
	ts.ForEach(func(key string, node bindings.Node) {
		switch node.(type) {
		case *bindings.Alias, *bindings.Interface, *bindings.VariableDeclaration:
			valid[key] = struct{}{}
		}
	})

	ts.ForEach(func(key string, node bindings.Node) {
		walk.Walk(&referenceFixer{valid: valid}, node)
	})
}

type referenceFixer struct {
	valid map[string]struct{}
	pkg   string
	msgs  []string
}

func (r *referenceFixer) Visit(node bindings.Node) (w walk.Visitor) {
	switch node := node.(type) {
	case *bindings.ReferenceType:
		if node.Name.Package == nil {
			// Unpackaged types are probably builtins
			return nil
		}
		if node.Name.PkgName() == r.pkg {
			// TypeParameters (Generics) are excluded here
			return r // Same package, skip
		}
		if _, ok := r.valid[node.Name.Ref()]; !ok {
			id := node.Name.GoName()
			// Invalid reference, change to 'any'
			node.Name = bindings.Identifier{Name: "any"}
			node.Arguments = []bindings.ExpressionType{}

			slog.Info(fmt.Sprintf("Type %q not found, changed to 'any'", id))
			r.msgs = append(r.msgs, fmt.Sprintf("Type %q not found, changed to 'any'", id))
			return nil // stop walking
		}
	case *bindings.Interface:
		for _, field := range node.Fields {
			fieldFixer := &referenceFixer{
				valid: r.valid,
				pkg:   node.Name.PkgName(),
			}
			walk.Walk(fieldFixer, field.Type)
			field.FieldComments = append(field.FieldComments, fieldFixer.msgs...)
		}
		return nil
	}

	return r
}
