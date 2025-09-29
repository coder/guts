package config

import (
	"fmt"
	"log/slog"
	"reflect"
	"slices"
	"strings"

	"github.com/coder/guts"
	"github.com/coder/guts/bindings"
	"github.com/coder/guts/bindings/walk"
)

// SimplifyOptional removes the null type from union types that have a question
// token. This is because if 'omitempty' or 'omitzero' is set, then golang will
// omit the object key, rather than sending a null value to the client.
//
// Example:
// number?: number | null --> number?: number
func SimplifyOptional(ts *guts.Typescript) {
	ts.ForEach(func(key string, node bindings.Node) {
		switch node := node.(type) {
		case *bindings.Interface:
			for _, prop := range node.Fields {
				if union, ok := prop.Type.(*bindings.UnionType); prop.QuestionToken && ok {
					newTs := []bindings.ExpressionType{}
					for _, ut := range union.Types {
						if _, isNull := ut.(*bindings.Null); isNull {
							continue
						}
						newTs = append(newTs, ut)
					}
					union.Types = newTs
				}
			}
		}
	})
}

// SimplifyOmitEmpty is a deprecated alias for SimplifyOptional.
// It was implemented for the `omitempty` case, however it's usage affects
// 'omitzero' as well.
// Deprecated: Use SimplifyOptional instead.
func SimplifyOmitEmpty(ts *guts.Typescript) {
	SimplifyOptional(ts)
}

// ExportTypes adds 'export' to all top level types.
// interface Foo {} --> export interface Foo{}
func ExportTypes(ts *guts.Typescript) {
	ts.ForEach(func(key string, node bindings.Node) {
		switch node := node.(type) {
		case *bindings.Alias:
			node.Modifiers = append(node.Modifiers, bindings.ModifierExport)
		case *bindings.Interface:
			node.Modifiers = append(node.Modifiers, bindings.ModifierExport)
		case *bindings.VariableStatement:
			node.Modifiers = append(node.Modifiers, bindings.ModifierExport)
		case *bindings.Enum:
			node.Modifiers = append(node.Modifiers, bindings.ModifierExport)
		default:
			panic(fmt.Sprintf("unexpected node type %T for exporting", node))
		}
	})
}

// ReadOnly sets all interface fields to 'readonly', resulting in
// all types being immutable.
// TODO: follow the AST all the way and find nested arrays
func ReadOnly(ts *guts.Typescript) {
	ts.ForEach(func(key string, node bindings.Node) {
		switch node := node.(type) {
		case *bindings.Alias:
			if _, isArray := node.Type.(*bindings.ArrayType); isArray {
				node.Type = bindings.OperatorNode(bindings.KeywordReadonly, node.Type)
			}
		case *bindings.Interface:
			for _, prop := range node.Fields {
				prop.Modifiers = append(prop.Modifiers, bindings.ModifierReadonly)
				if _, isArray := prop.Type.(*bindings.ArrayType); isArray {
					prop.Type = bindings.OperatorNode(bindings.KeywordReadonly, prop.Type)
				}
			}
		case *bindings.VariableStatement:
		case *bindings.Enum:
			// Enums are immutable by default
		default:
			panic("unexpected node type for exporting")
		}
	})
}

// TrimEnumPrefix removes the enum name from the member names.
func TrimEnumPrefix(ts *guts.Typescript) {
	ts.ForEach(func(key string, node bindings.Node) {
		enum, ok := node.(*bindings.Enum)
		if !ok {
			return
		}

		for _, member := range enum.Members {
			member.Name = strings.TrimPrefix(member.Name, enum.Name.Name)
		}
	})
}

// EnumAsTypes uses types to handle enums rather than using 'enum'.
// An enum will look like:
// type EnumString = "bar" | "baz" | "foo" | "qux";
func EnumAsTypes(ts *guts.Typescript) {
	ts.ForEach(func(key string, node bindings.Node) {
		enum, ok := node.(*bindings.Enum)
		if !ok {
			return
		}

		// Convert the enum to a union type
		union := &bindings.UnionType{
			Types: make([]bindings.ExpressionType, 0, len(enum.Members)),
		}
		for _, member := range enum.Members {
			union.Types = append(union.Types, member.Value)
		}

		// Replace the enum with an alias type
		ts.ReplaceNode(key, &bindings.Alias{
			Name:      enum.Name,
			Modifiers: enum.Modifiers,
			Type:      union,
			Source:    enum.Source,
		})
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
// TODO: Enums were changed to use proper enum types. This should be
// updated to support that. EnumLists only works with EnumAsTypes used first.
func EnumLists(ts *guts.Typescript) {
	addNodes := make(map[string]bindings.Node)
	ts.ForEach(func(key string, node bindings.Node) {
		// Find the enums, and make a list of values.
		// Only support primitive types.
		_, union, ok := isGoEnum(node)
		if !ok {
			return
		}

		values := make([]bindings.ExpressionType, 0, len(union.Types))
		for _, t := range union.Types {
			values = append(values, t)
		}

		// Pluralize the name
		name := key + "s"
		switch key[len(key)-1] {
		case 'x', 's', 'z':
			name = key + "es"
		}
		if strings.HasSuffix(key, "ch") || strings.HasSuffix(key, "sh") {
			name = key + "es"
		}

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
	})

	for name, node := range addNodes {
		if n, ok := ts.Node(name); ok {
			slog.Warn(fmt.Sprintf("enum list %s cannot be added, an existing declaration with that name exists. "+
				"To generate this enum list, the name collision must be resolved. ", name),
				slog.String("existing", fmt.Sprintf("%s", n)))
			continue
		}

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

		for _, field := range node.Fields {
			h := &hasAnyVisitor{}
			walk.Walk(h, field.Type)
			if h.hasAnyValue {
				field.FieldComments = append(field.FieldComments, "biome-ignore lint lint/complexity/noUselessTypeConstraint: ignore linter")
			}
		}

		return nil
	}

	return r
}

type hasAnyVisitor struct {
	hasAnyValue bool
}

func (h *hasAnyVisitor) Visit(node bindings.Node) walk.Visitor {
	if isLiteral, ok := node.(*bindings.LiteralKeyword); ok {
		if *isLiteral == bindings.KeywordAny {
			h.hasAnyValue = true
			return nil // stop here, the comment works for the whole field
		}
	}
	return h
}

// NullUnionSlices converts slices with nullable elements to remove the 'null'
// type from the union.
// This happens when a golang pointer is the element type of a slice.
// Example:
// GolangType: []*string
// TsType: (string | null)[] --> (string)[]
// TODO: Somehow remove the parenthesis from the output type.
// Might have to change the node from a union type to it's first element.
func NullUnionSlices(ts *guts.Typescript) {
	ts.ForEach(func(key string, node bindings.Node) {
		walk.Walk(&nullUnionVisitor{}, node)
	})
}

type nullUnionVisitor struct{}

func (v *nullUnionVisitor) Visit(node bindings.Node) walk.Visitor {
	if array, ok := node.(*bindings.ArrayType); ok {
		// Is array
		if union, ok := array.Node.(*bindings.UnionType); ok {
			hasNull := slices.ContainsFunc(union.Types, func(t bindings.ExpressionType) bool {
				_, isNull := t.(*bindings.Null)
				return isNull
			})

			// With union type
			if len(union.Types) == 2 && hasNull {
				// A union of 2 types, one being null
				// Remove the null type
				newTypes := make([]bindings.ExpressionType, 0, 1)
				for _, t := range union.Types {
					if _, isNull := t.(*bindings.Null); isNull {
						continue
					}
					newTypes = append(newTypes, t)
				}
				union.Types = newTypes

			}
		}
	}

	return v
}

// NotNullMaps assumes all maps will not be null.
// Example:
// GolangType: map[string]string
// TsType: Record<string,string> | null --> Record<string,string>
func NotNullMaps(ts *guts.Typescript) {
	ts.ForEach(func(key string, node bindings.Node) {
		walk.Walk(&notNullMaps{}, node)
	})
}

type notNullMaps struct{}

func (v *notNullMaps) Visit(node bindings.Node) walk.Visitor {
	if union, ok := node.(*bindings.UnionType); ok && len(union.Types) == 2 {
		hasNull := slices.ContainsFunc(union.Types, func(t bindings.ExpressionType) bool {
			_, isNull := t.(*bindings.Null)
			return isNull
		})

		var record bindings.ExpressionType
		index := slices.IndexFunc(union.Types, func(t bindings.ExpressionType) bool {
			ref, isRef := t.(*bindings.ReferenceType)
			if !isRef {
				return false
			}
			return ref.Name.Name == "Record"
		})
		if hasNull && index != -1 {
			record = union.Types[index]
			union.Types = []bindings.ExpressionType{record}
		}
	}

	return v
}

// InterfaceToType converts all interfaces to type aliases.
// interface User { name: string } --> type User = { name: string }
func InterfaceToType(ts *guts.Typescript) {
	ts.ForEach(func(key string, node bindings.Node) {
		intf, ok := node.(*bindings.Interface)
		if !ok {
			return
		}

		// Create a type literal node to represent the interface structure
		typeLiteral := &bindings.TypeLiteralNode{
			Members: intf.Fields,
		}

		// Replace the interface with a type alias
		ts.ReplaceNode(key, &bindings.Alias{
			Name:       intf.Name,
			Modifiers:  intf.Modifiers,
			Type:       typeLiteral,
			Parameters: intf.Parameters,
			Source:     intf.Source,
		})
	})
}

func isGoEnum(n bindings.Node) (*bindings.Alias, *bindings.UnionType, bool) {
	al, ok := n.(*bindings.Alias)
	if !ok {
		return nil, nil, false
	}

	union, ok := al.Type.(*bindings.UnionType)
	if !ok {
		return nil, nil, false
	}

	if len(union.Types) == 0 {
		return nil, nil, false
	}

	var expectedType *bindings.LiteralType
	// This might be a union type, if all elements are the same literal type.
	for _, t := range union.Types {
		value, ok := t.(*bindings.LiteralType)
		if !ok {
			return nil, nil, false
		}
		if expectedType == nil {
			expectedType = value
			continue
		}

		if reflect.TypeOf(expectedType.Value) != reflect.TypeOf(value.Value) {
			return nil, nil, false
		}
	}

	return al, union, true
}
