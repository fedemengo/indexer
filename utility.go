package indexer

import (
	"strings"
)

// SPECIAL represent all characters that should be ignored
var SPECIAL = map[rune]bool{
	'\\': true, '!': true,
	'"': true, '£': true,
	'$': true, '%': true,
	'^': true, '~': true,
	'#': true, '<': true,
	'>': true, ']': true,
	'}': true, '-': true,
	'1': true, '2': true,
	'3': true, '4': true,
	'5': true, '6': true,
	'7': true, '8': true,
	'9': true, '0': true,
}

// ExtractKeywords receive a plain string and extract "relevant" keyword
func ExtractKeywords(text string) []string {
	uniqueW := make(map[string]bool)
	words := strings.Split(text, " ")
	var result []string

	for _, s := range words {
		s = strings.ToLower(s)
		if _, ok := uniqueW[s]; !ok && isWordValid(s) {
			uniqueW[s] = true
			result = append(result, s)
		}
	}
	return result
}

// ExtractURLs receive a string of seed urls and create a slice of url
func ExtractURLs(urls string) []string {
	tmp := strings.FieldsFunc(urls, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\n'
	})

	var result []string
	for _, s := range tmp {
		if len(s) > len("http://") {
			result = append(result, s)
		}
	}

	return result
}

func isWordValid(word string) (ok bool) {
	if len(word) < 4 {
		return false
	}

	for _, char := range word {
		if _, ok := SPECIAL[char]; ok {
			return false
		}
	}

	return true
}
