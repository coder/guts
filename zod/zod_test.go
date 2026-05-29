package zod_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/coder/guts"
	"github.com/coder/guts/bindings"
	"github.com/coder/guts/zod"
)

// newTS constructs a fresh *guts.Typescript with no Go-derived nodes so a
// test can install its own minimal fixture with SetNode.
func newTS(t *testing.T) *guts.Typescript {
	t.Helper()
	gen, err := guts.NewGolangParser()
	require.NoError(t, err)
	ts, err := gen.ToTypescript()
	require.NoError(t, err)
	return ts
}

func runZod(t *testing.T, ts *guts.Typescript) string {
	t.Helper()
	ts.ApplyMutations(zod.AsSchemas)
	out, err := ts.Serialize()
	require.NoError(t, err)
	return out
}

func ident(name string) bindings.Identifier { return bindings.Identifier{Name: name} }

func kw(k bindings.LiteralKeyword) *bindings.LiteralKeyword { return &k }

// TestObjectFields exercises the keyword-to-z mappings inside an interface
// (z.string, z.number, z.boolean) and the optional refinement attached to
// fields with a question token.
func TestObjectFields(t *testing.T) {
	t.Parallel()

	ts := newTS(t)
	require.NoError(t, ts.SetNode("User", &bindings.Interface{
		Name: ident("User"),
		Fields: []*bindings.PropertySignature{
			{Name: "id", Type: kw(bindings.KeywordString)},
			{Name: "name", Type: kw(bindings.KeywordString), QuestionToken: true},
			{Name: "age", Type: kw(bindings.KeywordNumber)},
			{Name: "active", Type: kw(bindings.KeywordBoolean)},
		},
	}))

	out := runZod(t, ts)
	require.Contains(t, out, "const UserSchema = z.object(")
	require.Contains(t, out, "id: z.string()")
	require.Contains(t, out, "name: z.string().optional()")
	require.Contains(t, out, "age: z.number()")
	require.Contains(t, out, "active: z.boolean()")
	require.Contains(t, out, "type User = z.infer<typeof UserSchema>")
}

// TestStringLiteralUnionBecomesEnum checks that an Alias whose Type is a
// union of string literals collapses to z.enum([...]) rather than the more
// verbose z.union of z.literal calls.
func TestStringLiteralUnionBecomesEnum(t *testing.T) {
	t.Parallel()

	ts := newTS(t)
	require.NoError(t, ts.SetNode("Status", &bindings.Alias{
		Name: ident("Status"),
		Type: bindings.Union(
			&bindings.LiteralType{Value: "active"},
			&bindings.LiteralType{Value: "inactive"},
			&bindings.LiteralType{Value: "banned"},
		),
	}))

	out := runZod(t, ts)
	require.Contains(t, out, "const StatusSchema = z.enum(")
	require.Contains(t, out, `"active"`)
	require.Contains(t, out, `"inactive"`)
	require.Contains(t, out, `"banned"`)
	require.NotContains(t, out, "z.literal", "string-literal union must not fall back to z.literal+z.union")
}

// TestNullableField checks the union-with-null collapse: `T | null` becomes
// `T.nullable()` instead of `z.union([T, z.null()])`.
func TestNullableField(t *testing.T) {
	t.Parallel()

	ts := newTS(t)
	require.NoError(t, ts.SetNode("Thing", &bindings.Interface{
		Name: ident("Thing"),
		Fields: []*bindings.PropertySignature{
			{
				Name: "error",
				Type: bindings.Union(kw(bindings.KeywordString), &bindings.Null{}),
			},
		},
	}))

	out := runZod(t, ts)
	require.Contains(t, out, "error: z.string().nullable()")
	require.NotContains(t, out, "z.union", "T|null must not emit z.union")
}

// TestOptionalNullableField checks that QuestionToken and a `| null` union
// both apply, in nullable-then-optional order.
func TestOptionalNullableField(t *testing.T) {
	t.Parallel()

	ts := newTS(t)
	require.NoError(t, ts.SetNode("Thing", &bindings.Interface{
		Name: ident("Thing"),
		Fields: []*bindings.PropertySignature{
			{
				Name:          "parent_id",
				Type:          bindings.Union(kw(bindings.KeywordString), &bindings.Null{}),
				QuestionToken: true,
			},
		},
	}))

	out := runZod(t, ts)
	require.Contains(t, out, "parent_id: z.string().nullable().optional()")
}

// TestArrayField checks that a TypeScript array maps to z.array.
func TestArrayField(t *testing.T) {
	t.Parallel()

	ts := newTS(t)
	require.NoError(t, ts.SetNode("Thing", &bindings.Interface{
		Name: ident("Thing"),
		Fields: []*bindings.PropertySignature{
			{Name: "tags", Type: bindings.Array(kw(bindings.KeywordString))},
		},
	}))

	require.Contains(t, runZod(t, ts), "tags: z.array(z.string())")
}

// TestReferenceField checks that a bare type reference resolves to the
// paired Schema identifier.
func TestReferenceField(t *testing.T) {
	t.Parallel()

	ts := newTS(t)
	require.NoError(t, ts.SetNode("ErrorInfo", &bindings.Interface{
		Name: ident("ErrorInfo"),
		Fields: []*bindings.PropertySignature{
			{Name: "message", Type: kw(bindings.KeywordString)},
		},
	}))
	require.NoError(t, ts.SetNode("Chat", &bindings.Interface{
		Name: ident("Chat"),
		Fields: []*bindings.PropertySignature{
			{Name: "last_error", Type: bindings.Reference(ident("ErrorInfo"))},
		},
	}))

	require.Contains(t, runZod(t, ts), "last_error: ErrorInfoSchema")
}

// TestRecordField checks that Record<K, V> maps to z.record(K, V).
func TestRecordField(t *testing.T) {
	t.Parallel()

	ts := newTS(t)
	require.NoError(t, ts.SetNode("Thing", &bindings.Interface{
		Name: ident("Thing"),
		Fields: []*bindings.PropertySignature{
			{
				Name: "labels",
				Type: bindings.Reference(ident("Record"),
					kw(bindings.KeywordString),
					kw(bindings.KeywordString),
				),
			},
		},
	}))

	require.Contains(t, runZod(t, ts), "labels: z.record(z.string(), z.string())")
}

// TestInlineObjectLiteral checks that an inline object type produces an
// inline z.object expression rather than a free-standing schema.
func TestInlineObjectLiteral(t *testing.T) {
	t.Parallel()

	ts := newTS(t)
	require.NoError(t, ts.SetNode("Outer", &bindings.Interface{
		Name: ident("Outer"),
		Fields: []*bindings.PropertySignature{
			{
				Name: "nested",
				Type: &bindings.TypeLiteralNode{
					Members: []*bindings.PropertySignature{
						{Name: "x", Type: kw(bindings.KeywordNumber)},
						{Name: "y", Type: kw(bindings.KeywordNumber)},
					},
				},
			},
		},
	}))

	out := runZod(t, ts)
	require.Contains(t, out, "nested: z.object(")
	require.Contains(t, out, "x: z.number()")
	require.Contains(t, out, "y: z.number()")
}

// TestSelfReferenceLazy checks that a field whose type references the
// enclosing type wraps the schema in z.lazy() so the value-position
// reference does not fire before the binding exists.
func TestSelfReferenceLazy(t *testing.T) {
	t.Parallel()

	ts := newTS(t)
	require.NoError(t, ts.SetNode("Tree", &bindings.Interface{
		Name: ident("Tree"),
		Fields: []*bindings.PropertySignature{
			{Name: "value", Type: kw(bindings.KeywordNumber)},
			{Name: "children", Type: bindings.Array(bindings.Reference(ident("Tree")))},
		},
	}))

	out := runZod(t, ts)
	require.Contains(t, out, "children: z.array(z.lazy((): z.ZodType => TreeSchema))")
}

// TestHeritageExtend checks that single-base heritage maps to
// `BaseSchema.extend({...})`.
func TestHeritageExtend(t *testing.T) {
	t.Parallel()

	ts := newTS(t)
	require.NoError(t, ts.SetNode("Base", &bindings.Interface{
		Name: ident("Base"),
		Fields: []*bindings.PropertySignature{
			{Name: "id", Type: kw(bindings.KeywordString)},
		},
	}))
	require.NoError(t, ts.SetNode("Child", &bindings.Interface{
		Name: ident("Child"),
		Heritage: []*bindings.HeritageClause{
			{Args: []bindings.ExpressionType{bindings.Reference(ident("Base"))}},
		},
		Fields: []*bindings.PropertySignature{
			{Name: "extra", Type: kw(bindings.KeywordString)},
		},
	}))

	out := runZod(t, ts)
	require.Contains(t, out, "const ChildSchema = BaseSchema.extend(")
	require.Contains(t, out, "extra: z.string()")
}

// TestAppendsZodImport pins that AsSchemas appends the zod import so the
// generated file is self-contained without the caller having to know to
// add config.InjectImport("zod", "z") separately.
func TestAppendsZodImport(t *testing.T) {
	t.Parallel()

	ts := newTS(t)
	out := runZod(t, ts)
	require.Contains(t, out, `import { z } from "zod";`)
}

// TestMixedUnionStaysUnion checks that a union that is not all string
// literals and not just T|null keeps the z.union shape.
func TestMixedUnionStaysUnion(t *testing.T) {
	t.Parallel()

	ts := newTS(t)
	require.NoError(t, ts.SetNode("MixedUnion", &bindings.Alias{
		Name: ident("MixedUnion"),
		Type: bindings.Union(
			kw(bindings.KeywordString),
			kw(bindings.KeywordNumber),
		),
	}))

	out := runZod(t, ts)
	require.Contains(t, out, "z.union(")
	require.Contains(t, out, "z.string()")
	require.Contains(t, out, "z.number()")
}

// TestSingleMemberUnionUnwraps checks the single-non-null-member shortcut:
// `union { T }` collapses to just `T` rather than `z.union([T])`.
func TestSingleMemberUnionUnwraps(t *testing.T) {
	t.Parallel()

	ts := newTS(t)
	require.NoError(t, ts.SetNode("Wrapped", &bindings.Alias{
		Name: ident("Wrapped"),
		Type: bindings.Union(kw(bindings.KeywordString)),
	}))

	out := runZod(t, ts)
	require.Contains(t, out, "const WrappedSchema = z.string()")
	require.NotContains(t, out, "z.union", "single-member union must not wrap in z.union")
}

// TestPrefixedReference pins the cross-package prefix passthrough. An
// Identifier with a Prefix must flow through schemaIdent and emit the
// prefixed Schema name in both the schema declaration and the reference.
func TestPrefixedReference(t *testing.T) {
	t.Parallel()

	prefixed := bindings.Identifier{Name: "Item", Prefix: "External"}

	ts := newTS(t)
	require.NoError(t, ts.SetNode(prefixed.Ref(), &bindings.Interface{
		Name: prefixed,
		Fields: []*bindings.PropertySignature{
			{Name: "id", Type: kw(bindings.KeywordString)},
		},
	}))
	require.NoError(t, ts.SetNode("Holder", &bindings.Interface{
		Name: ident("Holder"),
		Fields: []*bindings.PropertySignature{
			{Name: "item", Type: bindings.Reference(prefixed)},
		},
	}))

	out := runZod(t, ts)
	require.Contains(t, out, "const ExternalItemSchema = z.object(")
	require.Contains(t, out, "item: ExternalItemSchema")
	require.Contains(t, out, "type ExternalItem = z.infer<typeof ExternalItemSchema>",
		strings.TrimSpace(out))
}
