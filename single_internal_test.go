package gots

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/coder/gots/bindings"
)

func TestParseSingle(t *testing.T) {
	t.Parallel()

	expr, err := parseExpression("string")
	require.NoError(t, err)
	key, isString := expr.(*bindings.LiteralKeyword)
	require.Equal(t, bindings.KeywordString, *key)
	require.True(t, isString)

	expr, err = parseExpression("[]string")
	require.NoError(t, err)
	array, is := expr.(*bindings.ArrayType)
	require.True(t, is)
	key, isString = array.Node.(*bindings.LiteralKeyword)
	require.True(t, isString)
	require.Equal(t, bindings.KeywordString, *key)
}
