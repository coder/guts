package nullable

import "database/sql"

type NullableFields struct {
	OmitEmpty         string       `json:"omitEmpty,omitempty"`
	Nullable          *string      `json:"nullable"`
	NullableOmitEmpty *string      `json:"nullableOmitEmpty,omitempty"`
	NullTime          sql.NullTime `json:"nullTime"`
}

type EmptyFields struct {
	Empty interface{} `json:"empty"`
}
