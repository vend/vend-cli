package stringutil

import (
	"strings"
	"unicode"
)

// Lstrip returns a copy of s
// with leading whitespace removed.
// See also stringutil.Rstrip and stringutil.Strip
func Lstrip(s string) string {
	return strings.TrimLeftFunc(s, f)
}

func f(r rune) bool {
	return unicode.IsSpace(r)
}
