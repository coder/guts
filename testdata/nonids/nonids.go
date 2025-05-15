package nonids

type Foo struct {
	// Hyphen is an odd case, but this field is not ignored
	Hyphen     string `json:"-,"`
	Ignored    string `json:"-"`
	Hyphenated string `json:"hyphenated-string"`
	Numbered   int    `json:"1numbered"`
}
