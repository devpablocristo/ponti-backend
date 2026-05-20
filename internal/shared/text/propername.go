// Package text provides canonical-storage and display formatting helpers for
// entity names (customer, project, manager, investor, field, lot, crop,
// season, actor display name). The semantics match the frontend helpers in
// ui/src/lib/properName.ts.
//
// CanonicalizeName produces the storage form: lowercase ASCII, only
// [a-z0-9 ]. Diacritics are stripped, any other character collapses to a
// single space, then whitespace is collapsed and trimmed.
//
// FormatProperName produces the display form: canonicalizes first, then
// title-cases each word except Spanish connectors (de, del, con, ...) which
// stay lowercase unless they are the first word.
package text

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var connectors = map[string]struct{}{
	"de":   {},
	"del":  {},
	"con":  {},
	"sin":  {},
	"a":    {},
	"al":   {},
	"y":    {},
	"e":    {},
	"o":    {},
	"u":    {},
	"en":   {},
	"para": {},
	"por":  {},
	"la":   {},
	"el":   {},
	"los":  {},
	"las":  {},
}

var stripDiacritics = transform.Chain(
	norm.NFD,
	runes.Remove(runes.In(unicode.Mn)),
	norm.NFC,
)

// CanonicalizeName returns the canonical storage form of an entity name.
func CanonicalizeName(value string) string {
	stripped, _, err := transform.String(stripDiacritics, value)
	if err != nil {
		stripped = value
	}
	lower := strings.ToLower(stripped)

	var b strings.Builder
	b.Grow(len(lower))
	lastWasSpace := true
	for _, r := range lower {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastWasSpace = false
			continue
		}
		if !lastWasSpace {
			b.WriteByte(' ')
			lastWasSpace = true
		}
	}
	return strings.TrimRight(b.String(), " ")
}

// FormatProperName returns the display form of an entity name.
func FormatProperName(value string) string {
	canonical := CanonicalizeName(value)
	if canonical == "" {
		return ""
	}
	words := strings.Split(canonical, " ")
	out := make([]string, len(words))
	for i, word := range words {
		if i > 0 {
			if _, ok := connectors[word]; ok {
				out[i] = word
				continue
			}
		}
		out[i] = capitalize(word)
	}
	return strings.Join(out, " ")
}

func capitalize(word string) string {
	if word == "" {
		return word
	}
	runesSlice := []rune(word)
	runesSlice[0] = unicode.ToUpper(runesSlice[0])
	return string(runesSlice)
}
