package genericstruct

type Foo[A comparable, B any] struct {
	FA A
	FB B
}

type Baz[S Foo[string, string], I comparable, X Foo[I, I]] struct {
	A S
	B X
	C I
}
