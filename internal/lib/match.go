package lib

import (
	"strings"
)

func Match(text string, keyword []string) bool {
	txt := Normalize(text)
	for _, k := range keyword {
		if k != "" && strings.Contains(txt, k) {
			return true
		}
	}

	return false
}
