package tags

type Tags struct {
	NoTag         int `json:"no_tag"`
	OmitEmpty     int `json:"omit_empty,omitempty"`
	OnlyOmitEmpty int `json:",omitempty"`
	Ignore        int `json:"-"`
	// Hypen appears in JSON as key "-"
	// See https://pkg.go.dev/encoding/json@master#Marshal
	// https://go.dev/play/p/vsW07aIi_Pj
	Hyphen       int `json:"-,"`
	OmitZero     int `json:"omit_zero,omitzero"`
	OnlyOmitZero int `json:",omitzero"`
}
