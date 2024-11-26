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

// FooBuzz has a json tag for the embedded
// See: https://go.dev/play/p/-p6QYmY8mtR
type FooBuzz struct {
	Buzz `json:"foo"` // Json tag changes the inheritance
	Bazz string       `json:"bazz"`
}

type Buzz struct {
	Bar string `json:"bar"`
}
