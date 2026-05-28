// Package text provides canonical-storage and display formatting helpers for
// entity names (customer, project, manager, investor, field, lot, crop,
// season, actor display name). The semantics match the frontend helpers in
// ui/src/lib/properName.ts.
//
// CanonicalizeName produces the storage form: lowercase Spanish text, only
// [a-z0-9ñ ]. Diacritics are stripped, except ñ/Ñ; any other character
// collapses to a single space, then whitespace is collapsed and trimmed.
//
// FormatProperName produces the display form: canonicalizes first, then
// title-cases each word except Spanish connectors (de, del, con, ...) which
// stay lowercase unless they are the first word.
package text

import (
	"strings"
	"unicode"

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

// Tokens that render uppercase in display even though storage keeps them
// lowercase: legal-entity suffixes and common Argentine agro acronyms.
var uppercaseTokens = map[string]struct{}{
	"srl":  {},
	"sa":   {},
	"sas":  {},
	"saci": {},
	"saca": {},
	"sac":  {},
	"sh":   {},
	"sc":   {},
	"scs":  {},
	"inta": {},
	"ypf":  {},
	"afip": {},
	"arba": {},
}

// CanonicalizeName returns the canonical storage form of an entity name.
func CanonicalizeName(value string) string {
	stripped := stripDiacriticsPreservingEnye(value)
	lower := strings.ToLower(stripped)

	var b strings.Builder
	b.Grow(len(lower))
	lastWasSpace := true
	for _, r := range lower {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == 'ñ' {
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

func stripDiacriticsPreservingEnye(value string) string {
	normalized := norm.NFD.String(value)
	out := make([]rune, 0, len(normalized))
	for _, r := range normalized {
		if unicode.Is(unicode.Mn, r) {
			if r == '\u0303' && len(out) > 0 {
				last := out[len(out)-1]
				if last == 'n' {
					out[len(out)-1] = 'ñ'
					continue
				}
				if last == 'N' {
					out[len(out)-1] = 'Ñ'
					continue
				}
			}
			continue
		}
		out = append(out, r)
	}
	return string(out)
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
		out[i] = formatWord(word, i)
	}
	return strings.Join(out, " ")
}

func formatWord(word string, index int) string {
	if word == "" {
		return word
	}
	if _, ok := uppercaseTokens[word]; ok {
		return strings.ToUpper(word)
	}
	if index > 0 {
		if _, ok := connectors[word]; ok {
			return word
		}
	}
	runesSlice := []rune(word)
	runesSlice[0] = unicode.ToUpper(runesSlice[0])
	return string(runesSlice)
}
