package config

import (
	"github.com/coder/guts"
	"github.com/coder/guts/bindings"
)

// InjectImport returns a mutation that appends a value-import statement
// for the given module to the top of the generated output.
//
//	ts.ApplyMutations(config.InjectImport("zod", "z"))
//	// import { z } from "zod"
//
// Repeat calls for the same module are merged by the underlying
// (*guts.Typescript).AppendImport: specifiers are unioned by (Name, Alias,
// IsTypeOnly) while preserving the order of first occurrence.
//
// Aliased names use the "Name=Alias" form, e.g.
//
//	config.InjectImport("./schemas", "Foo=Bar")
//	// import { Foo as Bar } from "./schemas"
func InjectImport(module string, names ...string) guts.MutationFunc {
	return injectImport(module, false, names...)
}

// InjectTypeImport returns a mutation that appends a type-only import
// statement for the given module:
//
//	ts.ApplyMutations(config.InjectTypeImport("./schemas", "Foo"))
//	// import type { Foo } from "./schemas"
//
// Aliasing follows the same "Name=Alias" form as InjectImport.
func InjectTypeImport(module string, names ...string) guts.MutationFunc {
	return injectImport(module, true, names...)
}

func injectImport(module string, isTypeOnly bool, names ...string) guts.MutationFunc {
	specs := make([]*bindings.ImportSpecifier, 0, len(names))
	for _, n := range names {
		name, alias := splitNameAlias(n)
		specs = append(specs, &bindings.ImportSpecifier{
			Name:  name,
			Alias: alias,
		})
	}
	decl := &bindings.ImportDeclaration{
		Module:     module,
		Named:      specs,
		IsTypeOnly: isTypeOnly,
	}
	return func(ts *guts.Typescript) {
		ts.AppendImport(decl)
	}
}

// splitNameAlias parses entries of the form "Name" or "Name=Alias".
func splitNameAlias(s string) (name, alias string) {
	for i := 0; i < len(s); i++ {
		if s[i] == '=' {
			return s[:i], s[i+1:]
		}
	}
	return s, ""
}
