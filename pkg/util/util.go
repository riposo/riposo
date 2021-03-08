package util

import "strings"

// SplitFunc behaves like strings.Split but with a callback fn.
func SplitFunc(s string, sep string, fn func(string)) {
	for {
		m := strings.Index(s, sep)
		if m < 0 {
			break
		}
		fn(s[:m])
		s = s[m+len(sep):]
	}
	if len(s) != 0 {
		fn(s)
	}
}
