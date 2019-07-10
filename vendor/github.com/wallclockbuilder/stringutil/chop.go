package stringutil

import (
	"strings"
)

// Chop returns a new String with the last character removed.
// If the string ends with \r\n, both characters are removed.
// Applying chop to an empty string returns an empty string.
func Chop(s string) string {
	if s == "" {
		return s
	}
	return chopChars(s)
}

func chopChars(s string) string {
	if strHasWhitespaceSuffix(s) {
		return chopWhitespace(s)
	}
	return chopPrintableChar(s)
}

func chopWhitespace(s string) string {
	return strings.TrimSuffix(s, whitespaceSuffix(s))
}

func strHasWhitespaceSuffix(s string) bool {
	return strings.HasSuffix(s, "\r") ||
		strings.HasSuffix(s, "\n")
}

func chopPrintableChar(s string) string {
	suffix := printableSuffix(s)
	return strings.TrimSuffix(s, suffix)
}

func whitespaceSuffix(s string) string {
	if strings.HasSuffix(s, "\r\n") {
		return "\r\n"
	}
	return simpleWhitespaceSuffix(s)
}

func printableSuffix(s string) string {
	return suffix(s)
}

func suffix(s string) string {
	i := strings.LastIndex(s, "")
	return s[i-1:]
}

func simpleWhitespaceSuffix(s string) string {
	i := strings.LastIndex(s, "") - 1
	return s[i:]
}
