package lib

import (
	"strings"
	"unicode"
)

// The only part where LeetCode actually helped :sob:
func Normalize(s string) string {
	var c strings.Builder
	// We allocate instantly, not letting strings to guess
	c.Grow(len(s))

	for _, i := range s {
		if i >= 'A' && i <= 'Z' {
			// Since 'A' is 65, we add up 32 - so it become 'a' (works for all the letters)
			c.WriteRune(i + 32)
		} else if i >= 'a' && i <= 'z' || i >= '0' && i <= '9' {
			c.WriteRune(i)
		} else if unicode.Is(unicode.Scripts["Hiragana"], i) || unicode.Is(unicode.Scripts["Katakana"], i) || unicode.Is(unicode.Scripts["Han"], i) {
			c.WriteRune(i)
		}
	}

	return c.String()
}
