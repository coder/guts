package config

import (
	"github.com/coder/gots"
	"github.com/coder/gots/bindings"
)

func OverrideLiteral(keyword bindings.LiteralKeyword) gots.TypeOverride {
	return func() bindings.ExpressionType {
		return ptr(keyword)
	}
}

func OverrideNullable(t gots.TypeOverride) gots.TypeOverride {
	return func() bindings.ExpressionType {
		return bindings.Union(t(), &bindings.Null{})
	}
}

// StandardMappings is a list of standard mappings for Go types to Typescript types.
func StandardMappings() map[string]gots.TypeOverride {
	return map[string]gots.TypeOverride{
		"time.Time": OverrideLiteral(bindings.KeywordString),

		"database/sql.NullTime":    OverrideNullable(OverrideLiteral(bindings.KeywordString)),
		"database/sql.NullString":  OverrideNullable(OverrideLiteral(bindings.KeywordString)),
		"database/sql.NullBool":    OverrideNullable(OverrideLiteral(bindings.KeywordBoolean)),
		"database/sql.NullInt64":   OverrideNullable(OverrideLiteral(bindings.KeywordNumber)),
		"database/sql.NullInt32":   OverrideNullable(OverrideLiteral(bindings.KeywordNumber)),
		"database/sql.NullInt16":   OverrideNullable(OverrideLiteral(bindings.KeywordNumber)),
		"database/sql.NullFloat64": OverrideNullable(OverrideLiteral(bindings.KeywordNumber)),

		"github.com/google/uuid.UUID":     OverrideLiteral(bindings.KeywordString),
		"github.com/google/uuid.NullUUID": OverrideNullable(OverrideLiteral(bindings.KeywordString)),
	}
}

func ptr[T any](v T) *T {
	return &v
}
