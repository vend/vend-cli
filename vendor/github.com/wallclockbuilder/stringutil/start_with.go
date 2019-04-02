package stringutil

import (
	"strings"
)

// Start_with returns true
// if s starts the prefix given.
func Start_with(s string, prefix string) bool {

	return strings.HasPrefix(s, prefix)
}
