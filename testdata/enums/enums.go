package enums

import "time"

type (
	// EnumString is a string-based enum
	EnumString string
	// EnumSliceType is a slice of string-based enums
	EnumSliceType []EnumString

	EnumInt int

	Audience string
)

const (
	// EnumFoo is the "foo" value
	// This comment should be preserved
	EnumFoo EnumString = "foo"
	// EnumBar is the "bar" value
	EnumBar EnumString = "bar"
	EnumBaz EnumString = "baz"
	EnumQux EnumString = "qux"
)

const (
	// EnumNumFoo is the number 5
	EnumNumFoo EnumInt = 5
	EnumNumBar EnumInt = 10
)

const (
	AudienceWorld  Audience = "world"
	AudienceTenant Audience = "tenant"
	// AudienceTeam is the "team" value
	AudienceTeam Audience = "team"
)

// EmptyEnum references `time.Duration`, so the constant is considered an enum.
// However, 'time.Duration' is not a referenced type, so the enum does not exist
// in the output.
// For now, this kind of constant is ignored.
const EmptyEnum = 30 * time.Second
