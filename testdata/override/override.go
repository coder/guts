package override

import (
	"net/url"
	"regexp"
)

type OverrideTypes struct {
	Reg *regexp.Regexp
	U   *url.URL
}
