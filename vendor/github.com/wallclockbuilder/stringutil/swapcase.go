package stringutil

import (
  "unicode"
  "strings"
)

// Swapcase returns a copy of str
// with uppercase alphabetic characters converted to lowercase and
// lowercase characters converted to uppercase.
// Note: case conversion is effective only in ASCII region.
func Swapcase(s string) string {
  return strings.Map(swapRuneCase, s)
}

func swapRuneCase(r rune) rune{
  if (unicode.IsLower(r)){
    return unicode.ToUpper(r)
  }
  return unicode.ToLower(r)
}
