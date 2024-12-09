package guts

import (
	"fmt"
	"regexp"
	"strings"
)

// parsePackageError parses the error returned from the go parser when it fails to parse a package
// and returns a human-readable error message.
func parsePackageError(e error) string {
	switch {
	case strings.Contains(e.Error(), "could not import"):
		// This error message is because the generator does not have access to some package.
		// This error message could be improved.
		parts := regexp.MustCompile(`could not import ([^\s]+)`).FindStringSubmatch(e.Error())
		if len(parts) >= 2 {
			return fmt.Sprintf("parsing package, suggest running 'go get %s' where calling the go generator to include the referenced package.", parts[1])
		}
		return "parsing package, import unavailable to generating code, try to add the package as a reference to the go generator"
	default:
		// Log error as is
		return "parsing package"
	}
}
