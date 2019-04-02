package stringutil

import (
	"strings"
	"unicode"
)

// Rstrip returns a copy of s
// with trailing whitespace removed.
// See also stringutil.Lstrip and stringutil.Strip
func Rstrip(s string) string {
	return strings.TrimRightFunc(s, unicode.IsSpace)
}
