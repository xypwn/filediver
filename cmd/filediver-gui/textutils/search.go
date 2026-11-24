package textutils

import (
	"strings"

	"github.com/charlievieth/strcase"
	"golang.org/x/text/language"
	"golang.org/x/text/search"
)

var itemMatcher = search.New(language.English, search.Loose)

func isAscii(strs ...string) bool {
	for _, s := range strs {
		if strcase.ContainsNonASCII(s) {
			return false
		}
	}
	return true
}

type Query struct {
	queryWords []string
	isAscii    bool
}

func NewQuery(query string) Query {
	qw := strings.Fields(query)
	for i := range qw {
		qw[i] = strings.ToLower(qw[i])
	}
	return Query{
		queryWords: qw,
		isAscii:    isAscii(qw...),
	}
}

func (q *Query) matchesAnyAscii(targetWords ...string) bool {
	for _, word := range targetWords {
		matchesAll := true
		for _, queryWord := range q.queryWords {
			if !strcase.Contains(word, queryWord) {
				matchesAll = false
				break
			}
		}
		if matchesAll {
			return true
		}
	}
	return false
}

// QueryMatchesAny returns true if query is empty or
// all words (separated by [strings.Fields]) in the query
// find a match in any of the supplied targetWords.
//
// Ignores casing, diacritics and width variant.
func (q *Query) MatchesAny(targetWords ...string) bool {
	if len(q.queryWords) == 0 {
		return true
	}

	if q.isAscii {
		if q.matchesAnyAscii(targetWords...) {
			// Fast path 1: Attempt to match ignoring diacritics
			return true
		} else if isAscii(targetWords...) {
			// Fast path 2: No diacritics exist -> there really is no match
			return false
		}
	}

	// Slow path: Use x/text diacritics matching.
	for _, word := range targetWords {
		matchesAll := true
		for _, queryWord := range q.queryWords {
			idx, _ := itemMatcher.IndexString(word, queryWord)
			if idx == -1 {
				matchesAll = false
				break
			}
		}
		if matchesAll {
			return true
		}
	}
	return false
}
