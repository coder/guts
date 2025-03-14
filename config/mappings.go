package config

import (
	"github.com/coder/guts"
	"github.com/coder/guts/bindings"
)

func OverrideLiteral(keyword bindings.LiteralKeyword) guts.TypeOverride {
	return func() bindings.ExpressionType {
		return ptr(keyword)
	}
}

func OverrideNullable(t guts.TypeOverride) guts.TypeOverride {
	return func() bindings.ExpressionType {
		return bindings.Union(t(), &bindings.Null{})
	}
}

// StandardMappings is a list of standard mappings for Go types to Typescript types.
func StandardMappings() map[string]guts.TypeOverride {
	return map[string]guts.TypeOverride{
		"time.Time":     OverrideLiteral(bindings.KeywordString),
		"time.Duration": OverrideLiteral(bindings.KeywordNumber),

		"database/sql.NullTime":    OverrideNullable(OverrideLiteral(bindings.KeywordString)),
		"database/sql.NullString":  OverrideNullable(OverrideLiteral(bindings.KeywordString)),
		"database/sql.NullBool":    OverrideNullable(OverrideLiteral(bindings.KeywordBoolean)),
		"database/sql.NullInt64":   OverrideNullable(OverrideLiteral(bindings.KeywordNumber)),
		"database/sql.NullInt32":   OverrideNullable(OverrideLiteral(bindings.KeywordNumber)),
		"database/sql.NullInt16":   OverrideNullable(OverrideLiteral(bindings.KeywordNumber)),
		"database/sql.NullFloat64": OverrideNullable(OverrideLiteral(bindings.KeywordNumber)),

		"github.com/google/uuid.UUID":     OverrideLiteral(bindings.KeywordString),
		"github.com/google/uuid.NullUUID": OverrideNullable(OverrideLiteral(bindings.KeywordString)),

		"net/netip.Addr": OverrideLiteral(bindings.KeywordString),
		"net/url.URL":    OverrideLiteral(bindings.KeywordString),
		"regexp.Regexp":  OverrideLiteral(bindings.KeywordString),
	}
}

func ptr[T any](v T) *T {
	return &v
}
