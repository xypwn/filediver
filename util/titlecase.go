package util

import (
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Returns the (english) string title cased, with minor words (a, an, and, of, on, the, to) lowercased
// unless they are the first word of the string
func PrettyTitleCase(name string) string {
	words := strings.Split(strings.ToLower(name), " ")
	smallwords := " a an and of on the to "
	caser := cases.Title(language.English)

	for index, word := range words {
		if strings.Contains(smallwords, " "+word+" ") && index != 0 {
			words[index] = word
		} else {
			words[index] = caser.String(word)
		}
	}
	return strings.Join(words, " ")
}
