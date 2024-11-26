package nullable

type NullableFields struct {
	OmitEmpty         string  `json:"omitEmpty,omitempty"`
	Nullable          *string `json:"nullable"`
	NullableOmitEmpty *string `json:"nullableOmitEmpty,omitempty"`
}

type EmptyFields struct {
	Empty interface{} `json:"empty"`
}
