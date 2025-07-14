package enums

import "time"

type (
	EnumString    string
	EnumSliceType []EnumString

	EnumInt int

	Audience string
)

const (
	EnumFoo EnumString = "foo"
	EnumBar EnumString = "bar"
	EnumBaz EnumString = "baz"
	EnumQux EnumString = "qux"
)

const (
	EnumNumFoo EnumInt = 5
	EnumNumBar EnumInt = 10
)

const (
	AudienceWorld  Audience = "world"
	AudienceTenant Audience = "tenant"
	AudienceTeam   Audience = "team"
)

// EmptyEnum references `time.Duration`, so the constant is considered an enum.
// However, 'time.Duration' is not a referenced type, so the enum does not exist
// in the output.
// For now, this kind of constant is ignored.
const EmptyEnum = 30 * time.Second
