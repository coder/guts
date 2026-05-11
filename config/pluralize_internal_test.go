package config

import "testing"

func TestPluralize(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in   string
		want string
	}{
		// Default: just add "s".
		{"User", "Users"},
		{"Audience", "Audiences"},
		{"EnumString", "EnumStrings"},
		{"EnumInt", "EnumInts"},

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
		c := c
		t.Run(c.in, func(t *testing.T) {
			t.Parallel()
			if got := pluralize(c.in); got != c.want {
				t.Fatalf("pluralize(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}
