package stringutil

import (
	"strings"
)

// Strip returns a copy of str
// with leading and trailing whitespace removed.
func Strip(s string) string {
	return strings.TrimSpace(s)
}
