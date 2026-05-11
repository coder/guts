package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPluralize(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in   string
		want string
	}{
		// Default: just add "s".
		{"User", "Users"},
		{"Audience", "Audiences"},

		// Ends in x, s, z: add "es".
		{"Box", "Boxes"},
		{"Bus", "Buses"},
		{"Buzz", "Buzzes"},

		// Ends in ch, sh: add "es".
		{"Church", "Churches"},
		{"Bush", "Bushes"},

		// Consonant + y: drop "y", add "ies".
		{"Policy", "Policies"},
		{"Category", "Categories"},
		{"Story", "Stories"},
		{"City", "Cities"},
		{"HealthSeverity", "HealthSeverities"},

		// Vowel + y: just add "s".
		{"Day", "Days"},
		{"Key", "Keys"},
		{"Boy", "Boys"},

		// Single-character edge cases.
		{"", ""},
		{"y", "ys"},
		{"A", "As"},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, c.want, pluralize(c.in))
		})
	}
}
