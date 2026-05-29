package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSplitNameAlias(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in        string
		wantName  string
		wantAlias string
	}{
		{in: "Foo", wantName: "Foo"},
		{in: "Foo=Bar", wantName: "Foo", wantAlias: "Bar"},
		{in: "=Bar", wantAlias: "Bar"},
		{in: "Foo=", wantName: "Foo"},
		{in: "", wantName: ""},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()

			name, alias := splitNameAlias(tc.in)
			require.Equal(t, tc.wantName, name)
			require.Equal(t, tc.wantAlias, alias)
		})
	}
}
