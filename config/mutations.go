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
			panic(fmt.Sprintf("unexpected node type %T for exporting", node))
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
				if _, isArray := prop.Type.(*bindings.ArrayType); isArray {
					prop.Type = bindings.OperatorNode(bindings.KeywordReadonly, prop.Type)
				}
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

// BiomeLintIgnoreAnyTypeParameters adds a biome-ignore comment to any type parameters that are of type "any".
// It is questionable if we should even add 'extends any' at all to the typescript.
func BiomeLintIgnoreAnyTypeParameters(ts *guts.Typescript) {
	ts.ForEach(func(key string, node bindings.Node) {
		walk.Walk(&anyLintIgnore{}, node)
	})
}

type anyLintIgnore struct {
}

func (r *anyLintIgnore) Visit(node bindings.Node) (w walk.Visitor) {
	switch node := node.(type) {
	case *bindings.Interface:
		anyParam := false
		for _, param := range node.Parameters {
			if isLiteral, ok := param.Type.(*bindings.LiteralKeyword); ok {
				if *isLiteral == bindings.KeywordAny {
					anyParam = true
					break
				}
			}
		}
		if anyParam {
			node.Comments = append(node.Comments, "biome-ignore lint lint/complexity/noUselessTypeConstraint: golang does 'any' for generics, typescript does not like it")
		}

		return nil
	}

	return r
}
