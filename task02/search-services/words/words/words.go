package words

import (
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/kljensen/snowball/english"
)

var re = regexp.MustCompile(`[^a-z0-9]+`)

func Normalize(phrase string) []string {
	words := re.Split(strings.ToLower(phrase), -1)
	seen := make(map[string]struct{})

	for _, word := range words {
		if word == "" || english.IsStopWord(word) {
			continue
		}

		seen[english.Stem(word, true)] = struct{}{}
	}

	return slices.Collect(maps.Keys(seen))
}
