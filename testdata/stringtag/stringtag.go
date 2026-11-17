package stringtag

type Priority int64

type Fields struct {
	Priority  int64            `json:"priority,string"`
	Named     Priority         `json:"named,string"`
	Optional  *int64           `json:"optional,string,omitempty"`
	Enabled   bool             `json:"enabled,string"`
	Ratio     float64          `json:"ratio,string"`
	Text      string           `json:"text,string"`
	Composite map[string]int64 `json:"composite,string"`
}
