//go:build !windows

package zod_test

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coder/guts"
	"github.com/coder/guts/config"
	"github.com/coder/guts/zod"
)

var updateGolden = flag.Bool("update", false, "update golden files")

// TestSerializeEndToEnd runs the full guts pipeline on a real Go
// package and compares the Zod output against a golden file.
// Run with -update to regenerate the golden file.
func TestSerializeEndToEnd(t *testing.T) {
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
	)

	output := zod.Serialize(ts)

	golden := filepath.Join("..", "testdata", "zod", "golden.ts")
	if *updateGolden {
		err = os.WriteFile(golden, []byte(output), 0o644)
		require.NoError(t, err)
		return
	}

	expected, err := os.ReadFile(golden)
	require.NoError(t, err, "run with -update to generate golden file")
	assert.Equal(t, string(expected), output)
}
