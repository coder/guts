package maps

type Bar[T any] struct {
	SimpleMap  map[string]string
	NumberMap  map[string]int
	GenericMap map[string]T
}
