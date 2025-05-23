package codersdk

type Buzz struct {
	Foo  `json:"foo"`
	Bazz string `json:"bazz"`
}

type Foo struct {
	Bar string `json:"bar"`
}

type FooBuzz[R Custom] struct {
	Something []R `json:"something"`
}

type Custom interface {
	Foo | Buzz
}

type FooBuzzMap[R Custom] struct {
	Something map[string]R `json:"something"`
}

type FooBuzzAnonymousUnion[R Foo | Buzz] struct {
	Something []R `json:"something"`
}
