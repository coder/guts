package union

type UnionConstraint[T string | int64] struct {
	Value T
}

// Repeated constraints are redundant
type Repeated[T string | string | int64 | uint64] struct {
	Value T
}
