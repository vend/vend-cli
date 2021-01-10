package stringutil

import (
	"strings"
)

// Chomp Returns a new String
// with the given record separator(sep) removed
// from the end of string(s) if present.
// Chomp also removes carriage return characters
// (that is it will remove \n, \r, and \r\n).
// If sep is an empty string,
// it will remove all trailing newlines from the string.
func Chomp(s string, sep string) string {
	if sep == "" {
		return trimWhitespace(s)
	}
	return strings.TrimSuffix(s, sep)
}

func trimWhitespace(s string) string {
	if isWhiteSpace(lastCharOf(s)) {
		return trimCarriageReturnEtc(s)
	}
	return s
}

func trimCarriageReturnEtc(s string) string {
	if !strings.HasSuffix(s, "\r\n") {
		return trimReturnEtc(s)
	}

	return trimAllCarriageReturn(s)

}

func trimReturnEtc(s string) string {
	if lastCharOf(s) == "\n" {
		return trimNewLine(s)
	}
	return strings.TrimSuffix(s, "\r")
}

func trimNewLine(s string) string {
	return strings.TrimSuffix(s, "\n")
}

func lastCharOf(s string) string {
	i := len(s) - 1
	return s[i:]
}

func isWhiteSpace(s string) bool {
	return strings.ContainsAny(s, "\n\r")
}

func trimAllCarriageReturn(s string) string {
	if s = chopReturn(s); hasSuffix(s, "\r\n") {
		s = trimAllCarriageReturn(s)
	}
	return s
}

func chopReturn(s string) string {
	return strings.TrimSuffix(s, "\r\n")
}

func hasSuffix(s string, suffix string) bool {
	return strings.HasSuffix(s, suffix)
}
