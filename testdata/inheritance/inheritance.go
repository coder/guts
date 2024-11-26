package codersdk

type Foo struct {
	Bar
	GenBar[string]
}

type Bar struct {
	BarField int
}

type GenBar[T comparable] struct {
	GenBarField T
}
