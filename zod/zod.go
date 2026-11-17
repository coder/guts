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
// # Ordering
//
// Zod schemas must be declared before any other schema references them
// because `const` bindings are not hoisted. AsSchemas does not reorder
// nodes itself; the caller is expected to pass SortByDependencies to
// Typescript.SerializeInOrder:
//
//	ts.ApplyMutations(zod.AsSchemas)
//	out, err := ts.SerializeInOrder(zod.SortByDependencies)
//
// SortByDependencies performs a Kahn's-algorithm topological sort over
// the schema VariableStatements, then pairs each schema with its inferred
// type alias. Self-references inside the same declaration are already
// broken by z.lazy in convertInterface and convertAlias, so the sort
// treats arrow-function bodies as non-dependencies.
//
// # Pipeline
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
	"sort"
	"strings"

	"github.com/coder/guts"
	"github.com/coder/guts/bindings"
	"github.com/coder/guts/bindings/walk"
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

// schemaSuffix is the suffix appended to a type name to produce its
// schema binding. `Foo` becomes `FooSchema`.
const schemaSuffix = "Schema"

// schemaIdent returns the Identifier for the schema binding paired with a
// type. `Foo` becomes `FooSchema`, with Package and Prefix preserved so
// cross-package disambiguation flows through .Ref() to the emitted name.
func schemaIdent(typeName bindings.Identifier) bindings.Identifier {
	return bindings.Identifier{
		Name:    typeName.Name + schemaSuffix,
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

// converter holds the per-declaration conversion state. Methods on
// converter recurse through a TypeScript type expression and emit the
// equivalent Zod expression with access to the surrounding type's name
// (so self-references can become z.lazy) and to its generic type
// parameters (so a reference to a type parameter becomes z.unknown).
type converter struct {
	self       bindings.Identifier
	typeParams map[string]bool
}

// newConverter builds a converter for one declaration. params may be nil
// when the declaration has no generic parameters.
func newConverter(self bindings.Identifier, params []*bindings.TypeParameter) *converter {
	tp := make(map[string]bool, len(params))
	for _, p := range params {
		tp[p.Name.Ref()] = true
	}
	return &converter{self: self, typeParams: tp}
}

// convertInterface rewrites an Interface into a schema VariableStatement
// plus an inferred type alias. The original key in ts.typescriptNodes is
// reused for the alias; the schema is added under <key>Schema.
func convertInterface(ts *guts.Typescript, key string, iface *bindings.Interface) {
	typeName := iface.Name
	schemaName := schemaIdent(typeName)
	c := newConverter(typeName, iface.Parameters)

	objLit := c.buildFieldsObject(iface.Fields)

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
	c := newConverter(typeName, alias.Parameters)

	var initializer bindings.ExpressionType
	if union, ok := alias.Type.(*bindings.UnionType); ok && isStringLiteralUnion(union) {
		initializer = zMethod("enum", stringLiteralArray(union))
	} else {
		initializer = c.exprToZod(alias.Type)
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
func (c *converter) buildFieldsObject(fields []*bindings.PropertySignature) *bindings.ObjectLiteralExpression {
	props := make([]*bindings.PropertyAssignment, 0, len(fields))
	for _, f := range fields {
		expr := c.exprToZod(f.Type)
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
// equivalent Zod schema expression. References back to the surrounding
// type use z.lazy() to avoid reference-before-declaration errors.
func (c *converter) exprToZod(expr bindings.ExpressionType) bindings.ExpressionType {
	if expr == nil {
		return zMethod("unknown")
	}
	switch e := expr.(type) {
	case *bindings.LiteralKeyword:
		return keywordToZod(e)
	case *bindings.LiteralType:
		return zMethod("literal", &bindings.LiteralType{Value: e.Value})
	case *bindings.ReferenceType:
		return c.referenceToZod(e)
	case *bindings.ArrayType:
		return zMethod("array", c.exprToZod(e.Node))
	case *bindings.TupleType:
		// Tuples are emitted as arrays today. A future variant could
		// switch on TupleType.Length to emit a true z.tuple().
		return zMethod("array", c.exprToZod(e.Node))
	case *bindings.UnionType:
		return c.unionToZod(e)
	case *bindings.Null:
		return zMethod("null")
	case *bindings.TypeLiteralNode:
		return c.typeLiteralToZod(e)
	case *bindings.TypeIntersection:
		return c.intersectionToZod(e)
	case *bindings.OperatorNodeType:
		// readonly/keyof/unique wrappers do not affect the Zod schema;
		// unwrap and emit the inner type directly.
		return c.exprToZod(e.Type)
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

// referenceToZod converts a type reference to a Zod expression.
//
// Resolution order:
//
//  1. References to a generic type parameter on the surrounding
//     declaration fall back to z.unknown(). Zod has no runtime
//     equivalent for an unbound type parameter.
//  2. Record<K, V> becomes z.record(K, V).
//  3. Other utility-type generics (Omit, Pick, Partial, Required) are not
//     yet modeled and fall back to z.unknown().
//  4. A reference to the surrounding declaration emits z.lazy to break
//     the value-position cycle.
//  5. Anything else emits the paired `<Name>Schema` identifier.
func (c *converter) referenceToZod(ref *bindings.ReferenceType) bindings.ExpressionType {
	name := ref.Name.Ref()

	if c.typeParams[name] {
		return zMethod("unknown")
	}

	if name == "Record" && len(ref.Arguments) == 2 {
		return zMethod("record",
			c.exprToZod(ref.Arguments[0]),
			c.exprToZod(ref.Arguments[1]),
		)
	}
	switch name {
	case "Omit", "Pick", "Partial", "Required":
		return zMethod("unknown")
	}

	if name == c.self.Ref() {
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
func (c *converter) unionToZod(u *bindings.UnionType) bindings.ExpressionType {
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
		return chain(c.exprToZod(nonNull[0]), "nullable")
	}
	if !hasNull && len(nonNull) == 1 {
		return c.exprToZod(nonNull[0])
	}

	args := make([]bindings.ExpressionType, 0, len(u.Types))
	for _, t := range u.Types {
		args = append(args, c.exprToZod(t))
	}
	return zMethod("union", &bindings.ArrayLiteralType{Elements: args})
}

// typeLiteralToZod inlines an object type literal as a `z.object({...})`
// expression. Members carry through the same optional-marker handling
// as top-level interface fields.
func (c *converter) typeLiteralToZod(tl *bindings.TypeLiteralNode) bindings.ExpressionType {
	props := make([]*bindings.PropertyAssignment, 0, len(tl.Members))
	for _, m := range tl.Members {
		expr := c.exprToZod(m.Type)
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
func (c *converter) intersectionToZod(it *bindings.TypeIntersection) bindings.ExpressionType {
	switch len(it.Types) {
	case 0:
		return zMethod("unknown")
	case 1:
		return c.exprToZod(it.Types[0])
	}
	out := c.exprToZod(it.Types[0])
	for _, t := range it.Types[1:] {
		out = zMethod("intersection", out, c.exprToZod(t))
	}
	return out
}

// SortByDependencies returns the nodes from a Typescript map ordered so
// that each schema's dependencies are emitted before the schema itself.
// It is intended to be passed to Typescript.SerializeInOrder when
// emitting Zod output:
//
//	out, err := ts.SerializeInOrder(zod.SortByDependencies)
//
// The algorithm:
//
//  1. Partition the input into schema VariableStatements (keys ending in
//     "Schema"), their paired type aliases, and other nodes.
//  2. Build a dependency graph by scanning each schema's initializer for
//     IdentifierExpression references that name another schema. Bodies
//     of ArrowFunction nodes are skipped, so z.lazy(() => OtherSchema)
//     does not create a hard dependency on OtherSchema. This lets users
//     break cross-type cycles manually with z.lazy.
//  3. Topologically sort using Kahn's algorithm with alphabetical
//     tie-breaking so the output is deterministic.
//  4. Anything left in a cycle is appended in alphabetical order. The
//     resulting TypeScript will compile only if those nodes use z.lazy
//     to defer their references.
//
// Each schema is emitted immediately followed by its paired alias so the
// `type Foo = z.infer<typeof FooSchema>` line stays next to its schema.
// Other nodes (anything not matching the schema-plus-alias shape) are
// emitted first in alphabetical order so they do not interleave with
// the sorted schemas.
func SortByDependencies(nodes map[string]bindings.Node) []bindings.Node {
	schemaKeys, aliasOf, otherKeys := partitionNodes(nodes)

	indegree, outEdges := buildDependencyGraph(nodes, schemaKeys)

	sorted := kahnSort(schemaKeys, indegree, outEdges)

	out := make([]bindings.Node, 0, len(nodes))
	for _, k := range otherKeys {
		out = append(out, nodes[k])
	}
	for _, k := range sorted {
		out = append(out, nodes[k])
		if alias, ok := aliasOf[k]; ok {
			out = append(out, nodes[alias])
		}
	}
	return out
}

// partitionNodes splits a Typescript node map into three groups:
//   - schemaKeys: keys of VariableStatement nodes ending in "Schema",
//     sorted alphabetically for deterministic seed order.
//   - aliasOf: a map from each schema key to its paired alias key
//     (e.g. "FooSchema" -> "Foo"), present only when both exist.
//   - otherKeys: all remaining keys, sorted alphabetically.
//
// Aliases that are paired with a schema are not included in otherKeys
// because SortByDependencies emits them next to their schema.
func partitionNodes(nodes map[string]bindings.Node) (schemaKeys []string, aliasOf map[string]string, otherKeys []string) {
	aliasOf = map[string]string{}
	schemaSet := map[string]bool{}
	for k, n := range nodes {
		if _, ok := n.(*bindings.VariableStatement); !ok {
			continue
		}
		if !strings.HasSuffix(k, schemaSuffix) {
			continue
		}
		schemaSet[k] = true
		schemaKeys = append(schemaKeys, k)
		aliasName := strings.TrimSuffix(k, schemaSuffix)
		if _, ok := nodes[aliasName]; ok {
			aliasOf[k] = aliasName
		}
	}
	pairedAlias := map[string]bool{}
	for _, v := range aliasOf {
		pairedAlias[v] = true
	}
	for k := range nodes {
		if schemaSet[k] || pairedAlias[k] {
			continue
		}
		otherKeys = append(otherKeys, k)
	}
	sort.Strings(schemaKeys)
	sort.Strings(otherKeys)
	return schemaKeys, aliasOf, otherKeys
}

// buildDependencyGraph walks each schema's initializer and records edges
// from dependency to dependent. outEdges[dep] lists the schemas that
// must be emitted after dep, and indegree[schema] counts how many
// schemas it depends on. ArrowFunction bodies are skipped so z.lazy
// references do not contribute hard dependencies.
func buildDependencyGraph(nodes map[string]bindings.Node, schemaKeys []string) (indegree map[string]int, outEdges map[string][]string) {
	schemaSet := make(map[string]bool, len(schemaKeys))
	for _, k := range schemaKeys {
		schemaSet[k] = true
	}

	indegree = make(map[string]int, len(schemaKeys))
	outEdges = make(map[string][]string, len(schemaKeys))
	for _, k := range schemaKeys {
		indegree[k] = 0
	}
	for _, k := range schemaKeys {
		vs := nodes[k].(*bindings.VariableStatement)
		deps := collectSchemaDeps(vs, schemaSet, k)
		for dep := range deps {
			outEdges[dep] = append(outEdges[dep], k)
			indegree[k]++
		}
	}
	for k := range outEdges {
		sort.Strings(outEdges[k])
	}
	return indegree, outEdges
}

// kahnSort runs Kahn's algorithm with alphabetical tie-breaking. Nodes
// remaining in a cycle after the queue drains are appended in
// alphabetical order. Callers must still break cross-type cycles with
// z.lazy; this fallback only keeps Serialize from dropping nodes.
func kahnSort(schemaKeys []string, indegree map[string]int, outEdges map[string][]string) []string {
	queue := make([]string, 0, len(schemaKeys))
	for _, k := range schemaKeys {
		if indegree[k] == 0 {
			queue = append(queue, k)
		}
	}
	sort.Strings(queue)

	sorted := make([]string, 0, len(schemaKeys))
	for len(queue) > 0 {
		head := queue[0]
		queue = queue[1:]
		sorted = append(sorted, head)
		for _, dep := range outEdges[head] {
			indegree[dep]--
			if indegree[dep] == 0 {
				queue = append(queue, dep)
			}
		}
		sort.Strings(queue)
	}

	placed := make(map[string]bool, len(sorted))
	for _, k := range sorted {
		placed[k] = true
	}
	for _, k := range schemaKeys {
		if !placed[k] {
			sorted = append(sorted, k)
		}
	}
	return sorted
}

// collectSchemaDeps returns the schema keys referenced as
// IdentifierExpression inside a VariableStatement, excluding the schema
// itself and excluding references inside ArrowFunction bodies (which is
// how z.lazy is emitted).
func collectSchemaDeps(vs *bindings.VariableStatement, schemas map[string]bool, self string) map[string]bool {
	deps := map[string]bool{}
	walk.Walk(&depVisitor{deps: deps, schemas: schemas, self: self}, vs)
	return deps
}

type depVisitor struct {
	deps    map[string]bool
	schemas map[string]bool
	self    string
}

func (d *depVisitor) Visit(node bindings.Node) walk.Visitor {
	if _, ok := node.(*bindings.ArrowFunction); ok {
		// Skip arrow function bodies. z.lazy(() => Other) defers its
		// reference at runtime, so it should not force Other to be
		// declared first.
		return nil
	}
	if ident, ok := node.(*bindings.IdentifierExpression); ok {
		name := ident.Name.Ref()
		if name != d.self && d.schemas[name] {
			d.deps[name] = true
		}
	}
	return d
}
