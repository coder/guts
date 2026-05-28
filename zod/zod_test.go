package zod_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coder/guts"
	"github.com/coder/guts/bindings"
	"github.com/coder/guts/zod"
)

// TestSerializeInterface verifies z.object() generation from an
// Interface node with various field types and optionality.
func TestSerializeInterface(t *testing.T) {
	t.Parallel()

	ts := newTestTS(t)
	ts.SetNode(t, "User", &bindings.Interface{
		Name: ident("User"),
		Fields: []*bindings.PropertySignature{
			{Name: "id", Type: kw(bindings.KeywordString)},
			{Name: "name", Type: kw(bindings.KeywordString), QuestionToken: true},
			{Name: "age", Type: kw(bindings.KeywordNumber)},
			{Name: "active", Type: kw(bindings.KeywordBoolean)},
		},
	})

	out := zod.Serialize(ts.TS)

	assert.Contains(t, out, `export const UserSchema = z.object({`)
	assert.Contains(t, out, `id: z.string(),`)
	assert.Contains(t, out, `name: z.string().optional(),`)
	assert.Contains(t, out, `age: z.number(),`)
	assert.Contains(t, out, `active: z.boolean(),`)
	assert.Contains(t, out, `export type User = z.infer<typeof UserSchema>;`)
}

// TestSerializeStringEnum verifies z.enum([...]) generation from a
// union-of-literals alias.
func TestSerializeStringEnum(t *testing.T) {
	t.Parallel()

	ts := newTestTS(t)
	ts.SetNode(t, "Status", &bindings.Alias{
		Name: ident("Status"),
		Type: bindings.Union(
			&bindings.LiteralType{Value: "active"},
			&bindings.LiteralType{Value: "inactive"},
			&bindings.LiteralType{Value: "banned"},
		),
	})

	out := zod.Serialize(ts.TS)

	assert.Contains(t, out, `export const StatusSchema = z.enum([`)
	assert.Contains(t, out, `"active"`)
	assert.Contains(t, out, `"inactive"`)
	assert.Contains(t, out, `"banned"`)
}

// TestSerializeNullable verifies T | null maps to .nullable().
func TestSerializeNullable(t *testing.T) {
	t.Parallel()

	ts := newTestTS(t)
	ts.SetNode(t, "Thing", &bindings.Interface{
		Name: ident("Thing"),
		Fields: []*bindings.PropertySignature{
			{
				Name: "error",
				Type: bindings.Union(kw(bindings.KeywordString), &bindings.Null{}),
			},
		},
	})

	out := zod.Serialize(ts.TS)

	assert.Contains(t, out, `error: z.string().nullable(),`)
}

// TestSerializeOptionalNullable verifies *T with omitempty maps to
// .nullable().optional() when QuestionToken is set.
func TestSerializeOptionalNullable(t *testing.T) {
	t.Parallel()

	ts := newTestTS(t)
	ts.SetNode(t, "Thing", &bindings.Interface{
		Name: ident("Thing"),
		Fields: []*bindings.PropertySignature{
			{
				Name:          "parent_id",
				Type:          bindings.Union(kw(bindings.KeywordString), &bindings.Null{}),
				QuestionToken: true,
			},
		},
	})

	out := zod.Serialize(ts.TS)

	assert.Contains(t, out, `parent_id: z.string().nullable().optional(),`)
}

// TestSerializeArray verifies z.array() generation.
func TestSerializeArray(t *testing.T) {
	t.Parallel()

	ts := newTestTS(t)
	ts.SetNode(t, "Thing", &bindings.Interface{
		Name: ident("Thing"),
		Fields: []*bindings.PropertySignature{
			{Name: "tags", Type: bindings.Array(kw(bindings.KeywordString))},
		},
	})

	out := zod.Serialize(ts.TS)

	assert.Contains(t, out, `tags: z.array(z.string()),`)
}

// TestSerializeReference verifies schema references between types.
func TestSerializeReference(t *testing.T) {
	t.Parallel()

	ts := newTestTS(t)
	ts.SetNode(t, "ErrorInfo", &bindings.Interface{
		Name: ident("ErrorInfo"),
		Fields: []*bindings.PropertySignature{
			{Name: "message", Type: kw(bindings.KeywordString)},
		},
	})
	ts.SetNode(t, "Chat", &bindings.Interface{
		Name: ident("Chat"),
		Fields: []*bindings.PropertySignature{
			{Name: "last_error", Type: bindings.Reference(ident("ErrorInfo"))},
		},
	})

	out := zod.Serialize(ts.TS)

	assert.Contains(t, out, `last_error: ErrorInfoSchema,`)
}

// TestSerializeRecord verifies Record<K, V> maps to z.record().
func TestSerializeRecord(t *testing.T) {
	t.Parallel()

	ts := newTestTS(t)
	ts.SetNode(t, "Thing", &bindings.Interface{
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
	})

	out := zod.Serialize(ts.TS)

	assert.Contains(t, out, `labels: z.record(z.string(), z.string()),`)
}

// TestSerializeFilter verifies that SerializeFilter only includes
// nodes matching the filter predicate.
func TestSerializeFilter(t *testing.T) {
	t.Parallel()

	ts := newTestTS(t)
	ts.SetNode(t, "Keep", &bindings.Interface{
		Name:   ident("Keep"),
		Fields: []*bindings.PropertySignature{{Name: "id", Type: kw(bindings.KeywordString)}},
	})
	ts.SetNode(t, "Drop", &bindings.Interface{
		Name:   ident("Drop"),
		Fields: []*bindings.PropertySignature{{Name: "id", Type: kw(bindings.KeywordString)}},
	})

	out := zod.SerializeFilter(ts.TS, func(name string) bool {
		return name == "Keep"
	})

	assert.Contains(t, out, "KeepSchema")
	assert.NotContains(t, out, "DropSchema")
}

// TestSerializeSingleMemberUnion verifies that a union with one
// non-null member simplifies to the member directly.
func TestSerializeSingleMemberUnion(t *testing.T) {
	t.Parallel()

	ts := newTestTS(t)
	ts.SetNode(t, "Thing", &bindings.Interface{
		Name: ident("Thing"),
		Fields: []*bindings.PropertySignature{
			{
				Name:          "workspace_id",
				Type:          bindings.Union(kw(bindings.KeywordString)),
				QuestionToken: true,
			},
		},
	})

	out := zod.Serialize(ts.TS)

	// Should be z.string().optional(), not z.union([z.string()]).optional().
	assert.Contains(t, out, `workspace_id: z.string().optional(),`)
	assert.NotContains(t, out, "z.union")
}

// TestSerializeObjectLiteral verifies inline object types.
func TestSerializeObjectLiteral(t *testing.T) {
	t.Parallel()

	ts := newTestTS(t)
	ts.SetNode(t, "Outer", &bindings.Interface{
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
	})

	out := zod.Serialize(ts.TS)

	assert.Contains(t, out, `nested: z.object({`)
	assert.Contains(t, out, `x: z.number(),`)
	assert.Contains(t, out, `y: z.number(),`)
}

// TestSerializeHeader verifies the generated header.
func TestSerializeHeader(t *testing.T) {
	t.Parallel()

	ts := newTestTS(t)
	out := zod.Serialize(ts.TS)

	assert.True(t, strings.HasPrefix(out, "// Code generated by 'guts'. DO NOT EDIT.\n"))
	assert.Contains(t, out, `import { z } from "zod";`)
}

// testTS wraps guts.Typescript for testing convenience.
type testTS struct {
	TS *guts.Typescript
}

func newTestTS(t *testing.T) *testTS {
	t.Helper()
	gen, err := guts.NewGolangParser()
	require.NoError(t, err)
	ts, err := gen.ToTypescript()
	require.NoError(t, err)
	return &testTS{TS: ts}
}

func (tt *testTS) SetNode(t *testing.T, name string, node bindings.Node) {
	t.Helper()
	err := tt.TS.SetNode(name, node)
	require.NoError(t, err)
}

func ident(name string) bindings.Identifier {
	return bindings.Identifier{Name: name}
}

func kw(k bindings.LiteralKeyword) *bindings.LiteralKeyword {
	return &k
}
