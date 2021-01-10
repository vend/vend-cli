package stringutil

// Reverse returns a new string
// with the characters from s
// in reverse order.
func Reverse(s string) string {
	r := ""
	for i := len(s); i != 0; i-- {
		r += s[i-1 : i]
	}
	return r
}
