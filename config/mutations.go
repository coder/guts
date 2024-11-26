package config

import "github.com/coder/gots/bindings"

func ExportTypes() func(key string, node bindings.Node) {
	return func(key string, node bindings.Node) {
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
	}
}

func ReadOnly() func(key string, node bindings.Node) {
	return func(key string, node bindings.Node) {
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
	}
}
