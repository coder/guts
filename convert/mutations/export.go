package mutations

import "github.com/coder/gots/bindings"

func ExportTypes() func(key string, node any) {
	return func(key string, node any) {
		switch node := node.(type) {
		case bindings.Alias:
			_ = node
		case bindings.Interface:
			_ = node
		case bindings.VariableStatement:
			_ = node
		}
	}
}
