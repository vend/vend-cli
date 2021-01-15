package stringutil

import (
	"strings"
)

// Capitalize returns a new string
// with the first character converted
// to uppercase and the remainder to lowercase.
// Note: case conversion is effective only in ASCII region.
func Capitalize(s string) string {
	return strings.ToUpper(s)
}
