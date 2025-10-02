package codersdk

type Score[T int | float32 | float64] struct {
	Points T   `json:"points"`
	Level  int `json:"level"`
}

type User[T comparable] struct {
	ID       T      `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	IsActive bool   `json:"is_active"`
}

type Player[ID comparable, P int | float32 | float64] struct {
	User[ID]
	Score[P]

	X int `json:"x"`
	Y int `json:"y"`
}

type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	Country string `json:"country"`
}

type GenericContainer[T any] struct {
	Value T   `json:"value"`
	Count int `json:"count"`
}
