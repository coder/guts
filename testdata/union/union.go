package union

type UnionConstraint[T string | int64] struct {
	Value T
}
