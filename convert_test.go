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
			case "testdata/excludecustom":
				err = gen.ExcludeCustom("github.com/coder/guts/testdata/excludecustom.Secret")
				require.NoErrorf(t, err, "exclude %q", dir)
			}

			gen.IncludeCustomDeclaration(config.StandardMappings())

			ts, err := gen.ToTypescript()
			require.NoError(t, err, "to typescript")

			mutations := []func(typescript *guts.Typescript){
				config.EnumAsTypes,
				config.EnumLists,
				config.ExportTypes,
				config.ReadOnly,
				config.NullUnionSlices,
			}

			mutsCSV, err := os.ReadFile(filepath.Join(dir, "mutations"))
			if err == nil {
				mutations = make([]func(typescript *guts.Typescript), 0)
				// load specific mutations
				muts := strings.Split(strings.TrimSpace(string(mutsCSV)), ",")
				for _, m := range muts {
					switch m {
					case "NotNullMaps":
						mutations = append(mutations, config.NotNullMaps)
					case "EnumAsTypes":
						mutations = append(mutations, config.EnumAsTypes)
					case "EnumLists":
						mutations = append(mutations, config.EnumLists)
					case "ExportTypes":
						mutations = append(mutations, config.ExportTypes)
					case "ReadOnly":
						mutations = append(mutations, config.ReadOnly)
					case "NullUnionSlices":
						mutations = append(mutations, config.NullUnionSlices)
					case "TrimEnumPrefix":
						mutations = append(mutations, config.TrimEnumPrefix)
					default:
						t.Fatal("unknown mutation, add it to the list:", m)
					}
					t.Logf("using mutation %s", m)
				}
			} else {
				t.Logf("using default mutations")
			}

			// Export all top level types
			ts.ApplyMutations(mutations...)

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

func TestNotNullMaps(t *testing.T) {
	gen, err := guts.NewGolangParser()
	require.NoError(t, err, "new convert")

	dir := filepath.Join(".", "testdata", "maps")
	err = gen.IncludeGenerate("./" + dir)
	require.NoErrorf(t, err, "include %q", dir)

	gen.IncludeCustomDeclaration(config.StandardMappings())

	ts, err := gen.ToTypescript()
	require.NoError(t, err, "to typescript")

	ts.ApplyMutations(
		config.NotNullMaps,
	)

	output, err := ts.Serialize()
	require.NoErrorf(t, err, "generate %q", dir)

	// Not perfect, this asserts if the record is a nullable type.
	require.Contains(t, output, "SimpleMap: Record<string, string>;", "no nullable Record")
}
