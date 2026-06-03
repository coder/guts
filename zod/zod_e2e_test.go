//go:build !windows
// +build !windows

package zod_test

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/coder/guts"
	"github.com/coder/guts/config"
	"github.com/coder/guts/zod"
)

var updateGolden = flag.Bool("update", false, "update the zod e2e golden file")

// TestAsSchemasEndToEnd runs the full guts pipeline on testdata/zod and
// compares the Zod output against a golden file. Run with -update to
// regenerate after intentional schema changes.
//
// The pipeline mirrors the recommended composition: EnumAsTypes lowers Go
// int and string enums into unions of literals, SimplifyOmitEmpty removes
// the redundant null from optional fields, zod.AsSchemas rewrites every
// Interface and Alias into a Zod schema plus inferred type, and
// ExportTypes adds the `export` modifier.
func TestAsSchemasEndToEnd(t *testing.T) {
	t.Parallel()

	gen, err := guts.NewGolangParser()
	require.NoError(t, err)

	err = gen.IncludeGenerate("github.com/coder/guts/testdata/zod")
	require.NoError(t, err)

	gen.IncludeCustomDeclaration(config.StandardMappings())

	ts, err := gen.ToTypescript()
	require.NoError(t, err)

	ts.ApplyMutations(
		config.EnumAsTypes,
		config.SimplifyOmitEmpty,
		zod.AsSchemas,
		config.ExportTypes,
	)

	output, err := ts.SerializeInOrder(zod.SortByDependencies)
	require.NoError(t, err)

	golden := filepath.Join("..", "testdata", "zod", "golden.ts")
	if *updateGolden {
		err = os.WriteFile(golden, []byte(output), 0o644)
		require.NoError(t, err)
		return
	}

	expected, err := os.ReadFile(golden)
	require.NoError(t, err, "run with -update to generate the golden file")
	require.Equal(t, string(expected), output)
}
