package bindings_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/coder/guts/bindings"
)

// roundTrip runs a node through the goja factory and printer, returning the
// emitted TypeScript text. It is used by every test in this file because
// each node type needs the same boilerplate: new bindings VM, dispatch
// through ToTypescriptNode, serialize.
func roundTrip(t *testing.T, node bindings.Node) string {
	t.Helper()
	b, err := bindings.New()
	require.NoError(t, err)

	obj, err := b.ToTypescriptNode(node)
	require.NoError(t, err)

	out, err := b.SerializeToTypescript(obj)
	require.NoError(t, err)
	return strings.TrimSpace(out)
}

// kw is a small helper for building keyword pointers without leaking a temp
// in every test case.
func kw(k bindings.LiteralKeyword) *bindings.LiteralKeyword {
	return &k
}

// zMethodCall builds the expression `z.<name>(<args>...)`, the most common
// shape these tests exercise. Tests live or die on whether method-chained
// calls serialize correctly, so a helper keeps the test bodies focused on
// the case under test.
func zMethodCall(name string, args ...bindings.ExpressionType) *bindings.CallExpression {
	return &bindings.CallExpression{
		Expression: &bindings.PropertyAccessExpression{
			Expression: &bindings.IdentifierExpression{Name: "z"},
			Name:       name,
		},
		Arguments: args,
	}
}

func TestIdentifierExpression(t *testing.T) {
	t.Parallel()
	got := roundTrip(t, &bindings.IdentifierExpression{Name: "BaseSchema"})
	require.Equal(t, "BaseSchema", got)
}

func TestPropertyAccessExpression(t *testing.T) {
	t.Parallel()
	got := roundTrip(t, &bindings.PropertyAccessExpression{
		Expression: &bindings.IdentifierExpression{Name: "z"},
		Name:       "string",
	})
	require.Equal(t, "z.string", got)
}

// TestCallExpression covers the basic shape and the chained method shape
// that the zod mutation will lean on heavily.
func TestCallExpression(t *testing.T) {
	t.Parallel()

	t.Run("no args", func(t *testing.T) {
		t.Parallel()
		got := roundTrip(t, zMethodCall("string"))
		require.Equal(t, "z.string()", got)
	})

	t.Run("with arg", func(t *testing.T) {
		t.Parallel()
		got := roundTrip(t, zMethodCall("literal", &bindings.LiteralType{Value: int64(42)}))
		require.Equal(t, "z.literal(42)", got)
	})

	t.Run("chained", func(t *testing.T) {
		t.Parallel()
		// z.string().optional()
		inner := zMethodCall("string")
		got := roundTrip(t, &bindings.CallExpression{
			Expression: &bindings.PropertyAccessExpression{
				Expression: inner,
				Name:       "optional",
			},
		})
		require.Equal(t, "z.string().optional()", got)
	})

	t.Run("nested", func(t *testing.T) {
		t.Parallel()
		// z.array(z.string())
		got := roundTrip(t, zMethodCall("array", zMethodCall("string")))
		require.Equal(t, "z.array(z.string())", got)
	})
}

func TestObjectLiteralExpression(t *testing.T) {
	t.Parallel()

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		got := roundTrip(t, &bindings.ObjectLiteralExpression{})
		require.Equal(t, "{}", got)
	})

	t.Run("single property", func(t *testing.T) {
		t.Parallel()
		got := roundTrip(t, &bindings.ObjectLiteralExpression{
			Properties: []*bindings.PropertyAssignment{
				{Name: "id", Initializer: zMethodCall("string")},
			},
		})
		require.Contains(t, got, "id: z.string()")
		require.True(t, strings.HasPrefix(got, "{"))
		require.True(t, strings.HasSuffix(got, "}"))
	})

	t.Run("multiple properties preserve order", func(t *testing.T) {
		t.Parallel()
		got := roundTrip(t, &bindings.ObjectLiteralExpression{
			Properties: []*bindings.PropertyAssignment{
				{Name: "id", Initializer: zMethodCall("string")},
				{Name: "age", Initializer: zMethodCall("number")},
				{Name: "active", Initializer: zMethodCall("boolean")},
			},
		})
		idIdx := strings.Index(got, "id:")
		ageIdx := strings.Index(got, "age:")
		activeIdx := strings.Index(got, "active:")
		require.Less(t, idIdx, ageIdx, "property declaration order must be preserved")
		require.Less(t, ageIdx, activeIdx, "property declaration order must be preserved")
		require.Contains(t, got, "id: z.string()")
		require.Contains(t, got, "age: z.number()")
		require.Contains(t, got, "active: z.boolean()")
	})
}

// TestPropertyAssignment exercises the leaf node directly. It rarely appears
// on its own outside an ObjectLiteralExpression, but verifying it as a
// standalone makes failures in the larger tests easier to bisect.
func TestPropertyAssignment(t *testing.T) {
	t.Parallel()
	got := roundTrip(t, &bindings.PropertyAssignment{
		Name:        "title",
		Initializer: zMethodCall("string"),
	})
	require.Equal(t, "title: z.string()", got)
}

func TestArrowFunction(t *testing.T) {
	t.Parallel()

	t.Run("no params no return type", func(t *testing.T) {
		t.Parallel()
		got := roundTrip(t, &bindings.ArrowFunction{
			Body: &bindings.IdentifierExpression{Name: "TicketSchema"},
		})
		require.Equal(t, "() => TicketSchema", got)
	})

	t.Run("with return type annotation", func(t *testing.T) {
		t.Parallel()
		// (): z.ZodType => TicketSchema
		got := roundTrip(t, &bindings.ArrowFunction{
			ReturnType: bindings.Reference(bindings.Identifier{Name: "z.ZodType"}),
			Body:       &bindings.IdentifierExpression{Name: "TicketSchema"},
		})
		require.Equal(t, "(): z.ZodType => TicketSchema", got)
	})

	t.Run("with typed parameter", func(t *testing.T) {
		t.Parallel()
		// (x: number) => x
		got := roundTrip(t, &bindings.ArrowFunction{
			Parameters: []*bindings.Parameter{
				{Name: "x", Type: kw(bindings.KeywordNumber)},
			},
			Body: &bindings.IdentifierExpression{Name: "x"},
		})
		require.Equal(t, "(x: number) => x", got)
	})
}

// TestTypeQuery exercises the `typeof <name>` node both standalone and as a
// generic argument inside a ReferenceType (its primary use site).
func TestTypeQuery(t *testing.T) {
	t.Parallel()

	t.Run("standalone", func(t *testing.T) {
		t.Parallel()
		got := roundTrip(t, &bindings.TypeQuery{Name: "FooSchema"})
		require.Equal(t, "typeof FooSchema", got)
	})

	t.Run("as generic argument", func(t *testing.T) {
		t.Parallel()
		// Foo<typeof FooSchema>
		got := roundTrip(t, bindings.Reference(
			bindings.Identifier{Name: "Foo"},
			&bindings.TypeQuery{Name: "FooSchema"},
		))
		require.Equal(t, "Foo<typeof FooSchema>", got)
	})
}

// TestComposeZodObject is an integration-level case that builds the exact
// `z.object({...})` shape the upcoming zod mutation will need to produce.
// If this test breaks, the zod mutation work cannot succeed.
func TestComposeZodObject(t *testing.T) {
	t.Parallel()

	// z.object({ id: z.string(), name: z.string().optional() })
	expr := zMethodCall("object", &bindings.ObjectLiteralExpression{
		Properties: []*bindings.PropertyAssignment{
			{Name: "id", Initializer: zMethodCall("string")},
			{
				Name: "name",
				Initializer: &bindings.CallExpression{
					Expression: &bindings.PropertyAccessExpression{
						Expression: zMethodCall("string"),
						Name:       "optional",
					},
				},
			},
		},
	})

	got := roundTrip(t, expr)
	require.Contains(t, got, "z.object({")
	require.Contains(t, got, "id: z.string()")
	require.Contains(t, got, "name: z.string().optional()")
}

// TestComposeZodLazyRef builds `z.lazy((): z.ZodType => TicketSchema)`, the
// self-reference workaround the zod mutation will use to break cycles.
func TestComposeZodLazyRef(t *testing.T) {
	t.Parallel()

	expr := zMethodCall("lazy", &bindings.ArrowFunction{
		ReturnType: bindings.Reference(bindings.Identifier{Name: "z.ZodType"}),
		Body:       &bindings.IdentifierExpression{Name: "TicketSchema"},
	})
	got := roundTrip(t, expr)
	require.Equal(t, "z.lazy((): z.ZodType => TicketSchema)", got)
}
