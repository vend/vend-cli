package stringutil

import (
	"strings"
)

// Upcase returns a copy of string s
// with all lowercase letters replaced
// with their uppercase counterparts.
func Upcase(s string) string {
	return strings.ToUpper(s)
}
