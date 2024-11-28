package json

import (
	"unicode"
	"unicode/utf8"
)

// the first symbol is the double quote, the second one is the name
func secondToLower(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError && size <= 1 {
		return s
	}

	r2, size2 := utf8.DecodeRuneInString(s[size:])
	if r2 == utf8.RuneError && size2 <= 1 {
		return s
	}

	lc := unicode.ToLower(r2)
	if r2 == lc {
		// won't happen
		return s
	}
	return string(r) + string(lc) + s[size+size2:]
}
