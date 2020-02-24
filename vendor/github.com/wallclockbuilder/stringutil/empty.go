package stringutil

// Empty returns true
// if s has a length of zero.
func Empty(s string) bool {
	return len(s) == 0
}
