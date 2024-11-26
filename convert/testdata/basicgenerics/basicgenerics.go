package basicgenerics

type Constraint interface {
	string | bool
}

type Basic[A Constraint] struct {
	Foo A
}
