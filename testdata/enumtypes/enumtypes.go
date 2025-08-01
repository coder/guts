package enumtypes

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

type EnumAlias = string

const (
	EnumAliasString     EnumAlias = "string"
	EnumAliasNumber     EnumAlias = "number"
	EnumAliasBoolean    EnumAlias = "bool"
	EnumAliasListString EnumAlias = "list(string)"
)
