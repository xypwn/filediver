package textutils

import (
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/search"
)

var itemMatcher = search.New(language.English, search.Loose)

// QueryMatchesAny returns true if query is empty or
// all words (separated by [strings.Fields]) in the query
// find a match in any of the supplied searchWords.
//
// Ignores casing, diacritics and width variant.
func QueryMatchesAny(query string, searchWords ...string) bool {
	for queryWord := range strings.FieldsSeq(query) {
		var searchWordsMatch bool
		for _, word := range searchWords {
			idx, _ := itemMatcher.IndexString(word, queryWord)
			if idx != -1 {
				searchWordsMatch = true
				break
			}
		}
		if !searchWordsMatch {
			return false
		}
	}
	return true
}
