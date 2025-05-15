package nullable

import "database/sql"

type NullableFields struct {
	OmitEmpty         string       `json:"omitEmpty,omitempty"`
	OmitZero          string       `json:"omitZero,omitzero"`
	Nullable          *string      `json:"nullable"`
	NullableOmitEmpty *string      `json:"nullableOmitEmpty,omitempty"`
	NullableOmitZero  *string      `json:"nullableOmitZero,omitzero"`
	NullTime          sql.NullTime `json:"nullTime"`
	SlicePointer      []*string    `json:"slicePointer"`
}

type EmptyFields struct {
	Empty interface{} `json:"empty"`
}
