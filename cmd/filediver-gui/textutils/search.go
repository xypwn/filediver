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
	includes []string
	excludes []string
	isAscii  bool
}

func NewQuery(query string) Query {
	q := Query{}
	q.isAscii = true
	for s := range strings.FieldsSeq(query) {
		excl := false
		if len(s) >= 2 && s[0] == '-' {
			s = s[1:]
			excl = true
		} else if strings.HasPrefix(s, "\\-") {
			s = s[1:]
		}
		if !excl {
			q.includes = append(q.includes, s)
		} else {
			q.excludes = append(q.excludes, s)
		}
		q.isAscii = q.isAscii && isAscii(s)
	}
	return q
}

func (q *Query) matchesAnyAscii(targetWords ...string) bool {
	for _, word := range targetWords {
		for _, queryWord := range q.excludes {
			if strcase.Contains(word, queryWord) {
				return false
			}
		}
	}
	for _, word := range targetWords {
		matchesAll := true
		for _, queryWord := range q.includes {
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

// MatchesAny returns true if query is empty.
//
// Inclusions: All inclusion words in the query
// find a match in any of the supplied targetWords.
//
// Exclusions: Any of the excluded query words matches
// any of the target words.
//
// Query words prefixed with "-" are excluded (use
// "\-" to escape).
//
// Ignores casing, diacritics and width variant.
func (q *Query) MatchesAny(targetWords ...string) bool {
	if len(q.includes) == 0 && len(q.excludes) == 0 {
		return true
	}

	if q.isAscii {
		if q.matchesAnyAscii(targetWords...) {
			// Fast path 1: Attempt to match ignoring diacritics
			if len(q.excludes) == 0 || isAscii(targetWords...) {
				return true
			}
			// targetWords may now contain diacritics that would
			// exclude the match
		} else if isAscii(targetWords...) {
			// Fast path 2: No diacritics exist -> there really is no match
			return false
		}
	}

	// Slow path: Use x/text diacritics matching
	for _, word := range targetWords {
		for _, queryWord := range q.excludes {
			idx, _ := itemMatcher.IndexString(word, queryWord)
			if idx != -1 {
				return false
			}
		}
	}
	for _, word := range targetWords {
		matchesAll := true
		for _, queryWord := range q.includes {
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
