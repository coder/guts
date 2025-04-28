package notnullmap

type Foo struct {
	Bar    map[string]bool
	Nested GenericFoo[map[string]int]
}

type GenericFoo[T any] struct {
	Bar T
}
