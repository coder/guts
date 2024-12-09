package codersdk

type Foo struct {
	Bar
	GenBar[string]
}

type Bar struct {
	BarField   int
	ErrorField error
}

type GenBar[T comparable] struct {
	GenBarField T
}

type IgnoreMe interface {
	IgnoreMe()
}
