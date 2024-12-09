//go:build !windows
// +build !windows

// Windows tests fail because the \n\r vs \n. It's not worth trying
// to replace newlines for os tests. If people start using this tool on windows
// and are seeing problems, then we can add build tags and figure it out.
package guts_test

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/coder/guts"
	"github.com/coder/guts/config"
)

func ExampleNewGolangParser() {
	// gen will convert the golang package to the typescript AST.
	// Configure it before calling ToTypescript().
	gen, _ := guts.NewGolangParser()

	// Pass in the directory of the package you want to convert.
	// You can mark a package as 'false' to include it as a reference, but not
	// generate types for it.
	_ = gen.IncludeGenerate("github.com/coder/guts/testdata/generics")

	// Default type mappings are useful, feel free to add your own
	gen.IncludeCustomDeclaration(config.StandardMappings())
	_ = gen.IncludeCustom(map[string]string{
		// To configure a custom type for a golang type, use the full package path.
		"github.com/coder/guts/testdata/generics.ExampleType": "string",
		// You can use golang type syntax to specify a type.
		"github.com/coder/guts/testdata/generics.AnotherExampleType": "map[string]*string",
	})

	// ts is the typescript AST. It can be mutated before serializing.
	ts, _ := gen.ToTypescript()

	ts.ApplyMutations(
		// Generates a constant which lists all enum values.
		config.EnumLists,
		// Adds 'readonly' to all interface fields.
		config.ReadOnly,
		// Adds 'export' to all top level types.
		config.ExportTypes,
	)

	output, _ := ts.Serialize()
	// Output is the typescript file text
	fmt.Println(output)
}

// updateGoldenFiles is a flag that can be set to update golden files.
var updateGoldenFiles = flag.Bool("update", false, "Update golden files")

func TestGeneration(t *testing.T) {
	t.Parallel()
	files, err := os.ReadDir("testdata")
	require.NoError(t, err, "read dir")

	for _, f := range files {
		if !f.IsDir() {
			// Only test directories
			continue
		}
		f := f
		t.Run(f.Name(), func(t *testing.T) {
			t.Parallel()

			gen, err := guts.NewGolangParser()
			require.NoError(t, err, "new convert")

			dir := filepath.Join(".", "testdata", f.Name())
			err = gen.IncludeGenerate("./" + dir)
			require.NoErrorf(t, err, "include %q", dir)

			switch dir {
			case "testdata/anyreference":
				err = gen.IncludeReference("github.com/coder/guts/testdata/prefix", "Prefix")
				require.NoErrorf(t, err, "include %q", dir)
			}

			gen.IncludeCustomDeclaration(config.StandardMappings())

			ts, err := gen.ToTypescript()
			require.NoError(t, err, "to typescript")

			// Export all top level types
			ts.ApplyMutations(
				config.EnumLists,
				config.ExportTypes,
				config.ReadOnly,
			)

			output, err := ts.Serialize()
			require.NoErrorf(t, err, "generate %q", dir)

			golden := filepath.Join(dir, f.Name()+".ts")
			expected, err := os.ReadFile(golden)
			require.NoErrorf(t, err, "read file %s", golden)
			expectedString := strings.TrimSpace(string(expected))
			output = strings.TrimSpace(output)
			if *updateGoldenFiles {
				// nolint:gosec
				err := os.WriteFile(golden, []byte(output+"\n"), 0o644)
				require.NoError(t, err, "write golden file")
			} else {
				require.Equal(t, expectedString, output, "matched output")
			}
		})
	}
}
