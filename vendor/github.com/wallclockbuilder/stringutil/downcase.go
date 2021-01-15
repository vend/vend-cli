package stringutil

import (
	"strings"
)

// Downcase returns a copy of str
// with all uppercase letters replaced
// with their lowercase counterparts.
func Downcase(s string) string {
	r := strings.ToLower(s)
	return r
}
