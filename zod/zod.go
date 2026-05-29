// Package zod converts the guts intermediate TypeScript AST into Zod v4
// schema declarations.
//
// The package exposes a single mutation, AsSchemas, that walks every
// Interface and Alias in a *guts.Typescript and replaces each one with:
//
//   - a VariableStatement for `const FooSchema = z.<expr>;`, and
//   - an Alias for `type Foo = z.infer<typeof FooSchema>;`.
//
// It also injects `import { z } from "zod"` so the generated file is
// self-contained.
//
// AsSchemas composes with the rest of the config mutations. The intended
// pipeline is:
//
//	ts.ApplyMutations(
//	    config.EnumAsTypes,        // int and string enums -> union of literals
//	    config.SimplifyOmitEmpty,  // omitempty -> drop null, keep optional
//	    zod.AsSchemas,             // rewrite Interface/Alias into Zod
//	    config.ExportTypes,        // add `export` to the new declarations
//	)
//
// Other mutations that walk Interface or Alias (ExportTypes, ReadOnly,
// etc.) should run after AsSchemas because the originals are replaced.
package zod

import (
	"github.com/coder/guts"
	"github.com/coder/guts/bindings"
)

// AsSchemas is the mutation entry point. It walks ts.typescriptNodes and
// rewrites each Interface and Alias into a VariableStatement + Alias pair
// expressed in Zod, and appends `import { z } from "zod"`.
func AsSchemas(ts *guts.Typescript) {
	ts.AppendImport(&bindings.ImportDeclaration{
		Module: "zod",
		Named:  []*bindings.ImportSpecifier{{Name: "z"}},
	})

	// Collect keys before mutating so the map iteration is not invalidated
	// when we Replace and Set during conversion.
	var keys []string
	ts.ForEach(func(name string, _ bindings.Node) {
		keys = append(keys, name)
	})

	for _, key := range keys {
		node, ok := ts.Node(key)
		if !ok {
			continue
		}
		switch n := node.(type) {
		case *bindings.Interface:
			convertInterface(ts, key, n)
		case *bindings.Alias:
			convertAlias(ts, key, n)
		}
	}
}

// schemaIdent returns the Identifier for the schema binding paired with a
// type. `Foo` becomes `FooSchema`, with Package and Prefix preserved so
// cross-package disambiguation flows through .Ref() to the emitted name.
func schemaIdent(typeName bindings.Identifier) bindings.Identifier {
	return bindings.Identifier{
		Name:    typeName.Name + "Schema",
		Package: typeName.Package,
		Prefix:  typeName.Prefix,
	}
}

// inferAlias builds `type <typeName> = z.infer<typeof <schemaName>>` for a
// single converted declaration.
func inferAlias(typeName bindings.Identifier) *bindings.Alias {
	return &bindings.Alias{
		Name: typeName,
		Type: &bindings.ReferenceType{
			Name: bindings.Identifier{Name: "z.infer"},
			Arguments: []bindings.ExpressionType{
				&bindings.TypeQuery{Name: schemaIdent(typeName)},
			},
		},
	}
}

// constSchema builds `const <schemaName> = <initializer>` with no
// modifiers. The Export mutation, if applied afterwards, adds `export`.
func constSchema(schemaName bindings.Identifier, initializer bindings.ExpressionType) *bindings.VariableStatement {
	return &bindings.VariableStatement{
		Modifiers: []bindings.Modifier{},
		Declarations: &bindings.VariableDeclarationList{
			Flags: bindings.NodeFlagsConstant,
			Declarations: []*bindings.VariableDeclaration{
				{
					Name:        schemaName,
					Initializer: initializer,
				},
			},
		},
	}
}

// zMethod builds `z.<name>(args...)` as an expression. It is the most
// common shape Zod schemas need.
func zMethod(name string, args ...bindings.ExpressionType) *bindings.CallExpression {
	return &bindings.CallExpression{
		Expression: &bindings.PropertyAccessExpression{
			Expression: &bindings.IdentifierExpression{Name: bindings.Identifier{Name: "z"}},
			Name:       name,
		},
		Arguments: args,
	}
}

// chain wraps `<expr>.<method>()` to extend a schema with a refinement
// like `.optional()` or `.nullable()`.
func chain(expr bindings.ExpressionType, method string) *bindings.CallExpression {
	return &bindings.CallExpression{
		Expression: &bindings.PropertyAccessExpression{
			Expression: expr,
			Name:       method,
		},
	}
}

// convertInterface rewrites an Interface into a schema VariableStatement
// plus an inferred type alias. The original key in ts.typescriptNodes is
// reused for the alias; the schema is added under <key>Schema.
func convertInterface(ts *guts.Typescript, key string, iface *bindings.Interface) {
	typeName := iface.Name
	schemaName := schemaIdent(typeName)

	objLit := buildFieldsObject(iface.Fields, typeName)

	var initializer bindings.ExpressionType
	if base, ok := heritageBase(iface); ok {
		// BaseSchema.extend({...})
		initializer = &bindings.CallExpression{
			Expression: &bindings.PropertyAccessExpression{
				Expression: &bindings.IdentifierExpression{Name: schemaIdent(base)},
				Name:       "extend",
			},
			Arguments: []bindings.ExpressionType{objLit},
		}
	} else {
		// z.object({...})
		initializer = zMethod("object", objLit)
	}

	ts.ReplaceNode(key, inferAlias(typeName))
	_ = ts.SetNode(schemaName.Ref(), constSchema(schemaName, initializer))
}

// convertAlias rewrites an Alias into a schema VariableStatement plus an
// inferred type alias.
func convertAlias(ts *guts.Typescript, key string, alias *bindings.Alias) {
	typeName := alias.Name
	schemaName := schemaIdent(typeName)

	var initializer bindings.ExpressionType
	if union, ok := alias.Type.(*bindings.UnionType); ok && isStringLiteralUnion(union) {
		initializer = zMethod("enum", stringLiteralArray(union))
	} else {
		initializer = exprToZod(alias.Type, typeName)
	}

	ts.ReplaceNode(key, inferAlias(typeName))
	_ = ts.SetNode(schemaName.Ref(), constSchema(schemaName, initializer))
}

// heritageBase returns the single heritage base of an Interface as an
// Identifier, if any. Zod's `.extend()` only models single inheritance,
// so multiple heritage clauses cause a panic to surface the mismatch
// rather than silently dropping one.
func heritageBase(iface *bindings.Interface) (bindings.Identifier, bool) {
	var base bindings.Identifier
	found := false
	for _, h := range iface.Heritage {
		for _, arg := range h.Args {
			ident, ok := heritageArgIdent(arg)
			if !ok {
				continue
			}
			if found {
				panic("zod: multiple heritage bases on " + iface.Name.Ref() + " (Zod has no multiple inheritance)")
			}
			base = ident
			found = true
		}
	}
	return base, found
}

// heritageArgIdent unwraps a heritage argument to the underlying
// Identifier when it is a plain type reference. Other shapes are not
// modeled and return false.
func heritageArgIdent(arg bindings.ExpressionType) (bindings.Identifier, bool) {
	switch n := arg.(type) {
	case *bindings.ExpressionWithTypeArguments:
		if rt, ok := n.Expression.(*bindings.ReferenceType); ok {
			return rt.Name, true
		}
	case *bindings.ReferenceType:
		return n.Name, true
	}
	return bindings.Identifier{}, false
}

// buildFieldsObject collects an Interface's fields into a single
// ObjectLiteralExpression whose values are zod expressions.
func buildFieldsObject(fields []*bindings.PropertySignature, selfName bindings.Identifier) *bindings.ObjectLiteralExpression {
	props := make([]*bindings.PropertyAssignment, 0, len(fields))
	for _, f := range fields {
		expr := exprToZod(f.Type, selfName)
		if f.QuestionToken {
			expr = chain(expr, "optional")
		}
		props = append(props, &bindings.PropertyAssignment{
			Name:        f.Name,
			Initializer: expr,
		})
	}
	return &bindings.ObjectLiteralExpression{Properties: props}
}

// isStringLiteralUnion reports whether every member of a union is a
// string literal. Such unions become z.enum([...]) rather than
// z.union([z.literal(...), ...]) for readability.
func isStringLiteralUnion(u *bindings.UnionType) bool {
	if len(u.Types) == 0 {
		return false
	}
	for _, t := range u.Types {
		lit, ok := t.(*bindings.LiteralType)
		if !ok {
			return false
		}
		if _, ok := lit.Value.(string); !ok {
			return false
		}
	}
	return true
}

// stringLiteralArray collects the string values from a string-literal
// union into an ArrayLiteralType suitable for `z.enum([...])`.
func stringLiteralArray(u *bindings.UnionType) *bindings.ArrayLiteralType {
	elems := make([]bindings.ExpressionType, 0, len(u.Types))
	for _, t := range u.Types {
		if lit, ok := t.(*bindings.LiteralType); ok {
			elems = append(elems, &bindings.LiteralType{Value: lit.Value})
		}
	}
	return &bindings.ArrayLiteralType{Elements: elems}
}

// exprToZod recursively converts a TypeScript type expression into the
// equivalent Zod schema expression. selfName is the type currently being
// emitted; references back to it use z.lazy() to avoid
// reference-before-declaration errors.
func exprToZod(expr bindings.ExpressionType, selfName bindings.Identifier) bindings.ExpressionType {
	if expr == nil {
		return zMethod("unknown")
	}
	switch e := expr.(type) {
	case *bindings.LiteralKeyword:
		return keywordToZod(e)
	case *bindings.LiteralType:
		return zMethod("literal", &bindings.LiteralType{Value: e.Value})
	case *bindings.ReferenceType:
		return referenceToZod(e, selfName)
	case *bindings.ArrayType:
		return zMethod("array", exprToZod(e.Node, selfName))
	case *bindings.TupleType:
		// Tuples are emitted as arrays today. A future variant could
		// switch on TupleType.Length to emit a true z.tuple().
		return zMethod("array", exprToZod(e.Node, selfName))
	case *bindings.UnionType:
		return unionToZod(e, selfName)
	case *bindings.Null:
		return zMethod("null")
	case *bindings.TypeLiteralNode:
		return typeLiteralToZod(e, selfName)
	case *bindings.TypeIntersection:
		return intersectionToZod(e, selfName)
	case *bindings.OperatorNodeType:
		// readonly/keyof/unique wrappers do not affect the Zod schema;
		// unwrap and emit the inner type directly.
		return exprToZod(e.Type, selfName)
	default:
		return zMethod("unknown")
	}
}

// keywordToZod maps a TypeScript keyword to its z.<keyword>() form.
func keywordToZod(kw *bindings.LiteralKeyword) bindings.ExpressionType {
	switch *kw {
	case bindings.KeywordString:
		return zMethod("string")
	case bindings.KeywordNumber:
		return zMethod("number")
	case bindings.KeywordBoolean:
		return zMethod("boolean")
	case bindings.KeywordAny, bindings.KeywordUnknown:
		return zMethod("unknown")
	case bindings.KeywordVoid, bindings.KeywordUndefined:
		return zMethod("undefined")
	case bindings.KeywordNever:
		return zMethod("never")
	default:
		return zMethod("unknown")
	}
}

// referenceToZod converts a type reference to a Zod expression. Bare
// references emit the paired `<Name>Schema` identifier. The Record
// generic becomes `z.record(K, V)`. Other utility-type generics (Omit,
// Pick, Partial, Required) are not yet modeled and fall back to
// z.unknown().
func referenceToZod(ref *bindings.ReferenceType, selfName bindings.Identifier) bindings.ExpressionType {
	name := ref.Name.Ref()

	if name == "Record" && len(ref.Arguments) == 2 {
		return zMethod("record",
			exprToZod(ref.Arguments[0], selfName),
			exprToZod(ref.Arguments[1], selfName),
		)
	}
	switch name {
	case "Omit", "Pick", "Partial", "Required":
		return zMethod("unknown")
	}

	if name == selfName.Ref() {
		// z.lazy((): z.ZodType => SelfSchema) breaks a value-position
		// reference cycle without making the surrounding type lazy.
		return zMethod("lazy", &bindings.ArrowFunction{
			ReturnType: bindings.Reference(bindings.Identifier{Name: "z.ZodType"}),
			Body:       &bindings.IdentifierExpression{Name: schemaIdent(ref.Name)},
		})
	}

	return &bindings.IdentifierExpression{Name: schemaIdent(ref.Name)}
}

// unionToZod handles three union shapes:
//   - T | null collapses to <T>.nullable().
//   - A union with a single non-null member emits just that member; the
//     null is dropped because the surrounding optional marker covers it.
//   - Anything else becomes z.union([...]).
func unionToZod(u *bindings.UnionType, selfName bindings.Identifier) bindings.ExpressionType {
	nonNull := make([]bindings.ExpressionType, 0, len(u.Types))
	hasNull := false
	for _, t := range u.Types {
		if _, ok := t.(*bindings.Null); ok {
			hasNull = true
			continue
		}
		nonNull = append(nonNull, t)
	}

	if hasNull && len(nonNull) == 1 {
		return chain(exprToZod(nonNull[0], selfName), "nullable")
	}
	if !hasNull && len(nonNull) == 1 {
		return exprToZod(nonNull[0], selfName)
	}

	args := make([]bindings.ExpressionType, 0, len(u.Types))
	for _, t := range u.Types {
		args = append(args, exprToZod(t, selfName))
	}
	return zMethod("union", &bindings.ArrayLiteralType{Elements: args})
}

// typeLiteralToZod inlines an object type literal as a `z.object({...})`
// expression. Members carry through the same optional-marker handling
// as top-level interface fields.
func typeLiteralToZod(tl *bindings.TypeLiteralNode, selfName bindings.Identifier) bindings.ExpressionType {
	props := make([]*bindings.PropertyAssignment, 0, len(tl.Members))
	for _, m := range tl.Members {
		expr := exprToZod(m.Type, selfName)
		if m.QuestionToken {
			expr = chain(expr, "optional")
		}
		props = append(props, &bindings.PropertyAssignment{
			Name:        m.Name,
			Initializer: expr,
		})
	}
	return zMethod("object", &bindings.ObjectLiteralExpression{Properties: props})
}

// intersectionToZod folds an intersection into a left-associative chain
// of z.intersection(a, b) calls so the schema preserves intersection
// semantics for arbitrary member counts.
func intersectionToZod(it *bindings.TypeIntersection, selfName bindings.Identifier) bindings.ExpressionType {
	switch len(it.Types) {
	case 0:
		return zMethod("unknown")
	case 1:
		return exprToZod(it.Types[0], selfName)
	}
	out := exprToZod(it.Types[0], selfName)
	for _, t := range it.Types[1:] {
		out = zMethod("intersection", out, exprToZod(t, selfName))
	}
	return out
}
