package pkgutils

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"github.com/devpablocristo/core/backend/go/stringutil"
)

// IsNumeric returns true if the string contains only digits.
// Deprecated: Use validations.ValidateNumeric instead.
func IsNumeric(s string) bool {
	s = strings.TrimSpace(s)
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// NormalizeString lowercases and removes accents from a string.
// Usa golang.org/x/text para normalización Unicode precisa (no delegada a core).
func NormalizeString(input string) string {
	input = strings.ToLower(strings.TrimSpace(input))
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, input)
	// Keep only a-z chars
	clean := make([]rune, 0, len(result))
	for _, r := range result {
		if r >= 'a' && r <= 'z' {
			clean = append(clean, r)
		}
	}
	return string(clean)
}

// BasicInputSanitizer delega al helper estándar de core.
func BasicInputSanitizer(input string) string {
	return stringutil.BasicInputSanitizer(input)
}
