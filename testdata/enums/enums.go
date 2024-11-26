package codersdk

type (
	EnumString    string
	EnumSliceType []EnumString

	EnumInt int
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
