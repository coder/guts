package bindings_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/coder/guts/bindings"
)

// TestImportDeclaration exercises the goja factory wrappers for
// ImportSpecifier and ImportDeclaration end-to-end: build the bindings node,
// convert through ToTypescriptNode, then ask the embedded printer for its
// TypeScript representation.
func TestImportDeclaration(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		decl    *bindings.ImportDeclaration
		want    string
		notWant string
	}{
		{
			name: "single",
			decl: &bindings.ImportDeclaration{
				Module: "zod",
				Named: []*bindings.ImportSpecifier{
					{Name: "z"},
				},
			},
			want: `import { z } from "zod";`,
		},
		{
			name: "multiple",
			decl: &bindings.ImportDeclaration{
				Module: "./schemas",
				Named: []*bindings.ImportSpecifier{
					{Name: "Foo"},
					{Name: "Bar"},
				},
			},
			want: `import { Foo, Bar } from "./schemas";`,
		},
		{
			name: "aliased",
			decl: &bindings.ImportDeclaration{
				Module: "./schemas",
				Named: []*bindings.ImportSpecifier{
					{Name: "Foo", Alias: "Bar"},
				},
			},
			want: `import { Foo as Bar } from "./schemas";`,
		},
		{
			name: "type only",
			decl: &bindings.ImportDeclaration{
				Module:     "./types",
				IsTypeOnly: true,
				Named: []*bindings.ImportSpecifier{
					{Name: "Baz"},
				},
			},
			want: `import type { Baz } from "./types";`,
		},
		{
			name: "mixed",
			decl: &bindings.ImportDeclaration{
				Module: "./schemas",
				Named: []*bindings.ImportSpecifier{
					{Name: "Foo"},
					{Name: "Bar", Alias: "Baz"},
				},
			},
			want: `import { Foo, Bar as Baz } from "./schemas";`,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			b, err := bindings.New()
			require.NoError(t, err)

			node, err := b.ToTypescriptNode(tc.decl)
			require.NoError(t, err)

			got, err := b.SerializeToTypescript(node)
			require.NoError(t, err)

			require.Equal(t, tc.want, strings.TrimSpace(got))
		})
	}
}
