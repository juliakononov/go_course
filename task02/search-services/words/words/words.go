package words

import (
	"maps"
	"slices"
	"strings"
	"unicode"

	"github.com/kljensen/snowball/english"
)

func Normalize(phrase string) []string {
	seen := make(map[string]bool)

	words := strings.FieldsFunc(phrase, func(r rune) bool {
		return !unicode.IsDigit(r) && !unicode.IsLetter(r)
	})

	for _, word := range words {
		word := strings.ToLower(word)
		if english.IsStopWord(word) {
			continue
		}
		seen[english.Stem(word, false)] = true
	}

	return slices.Collect(maps.Keys(seen))
}
