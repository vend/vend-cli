package stringutil

import (
	"regexp"
	"strings"
)

// Include returns true if str contains the given string
// or character.
func Include(s string, substr string) bool {
	r := strings.Contains(s, substr)
	if firstChar(substr) == "?" {
		r, _ = regexp.MatchString(patternRe2(substr), s)
	}
	return r
}

func firstChar(s string) string {
	return string(s[0:1])
}

// patternRe2 takes in ruby user-friendly
// pseudo regex and returns RE2 compliant regex pattern.
// pseudo regex pattern: "?h"
// RE2 compliant regex pattern: "h?"
func patternRe2(s string) string {
	return s[1:] + s[0:1]
}
