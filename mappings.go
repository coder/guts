package gots

import "github.com/coder/gots/bindings"

// StandardMappings is a list of standard mappings for Go types to Typescript types.
func StandardMappings() map[string]bindings.ExpressionType {
	return map[string]bindings.ExpressionType{
		"time.Time":                   ptr(bindings.KeywordString),
		"github.com/google/uuid.UUID": ptr(bindings.KeywordString),
		"database/sql.NullTime":       ptr(bindings.KeywordString),
	}
}
