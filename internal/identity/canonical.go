// Package identity implementa el "Identity Gate" (Pilar 3): resolución y prevención
// de duplicados de identidad (actores) por CUIT + nombre legal.
package identity

import (
	"regexp"
	"strings"
)

// accentReplacer saca tildes de vocales + ç/ÿ pero NO toca la ñ (preparado para
// español). Equivale al translate de normalize_name (migr 240).
var accentReplacer = strings.NewReplacer(
	"á", "a", "à", "a", "ä", "a", "â", "a", "ã", "a",
	"é", "e", "è", "e", "ë", "e", "ê", "e",
	"í", "i", "ì", "i", "ï", "i", "î", "i",
	"ó", "o", "ò", "o", "ö", "o", "ô", "o", "õ", "o",
	"ú", "u", "ù", "u", "ü", "u", "û", "u",
	"ç", "c", "ÿ", "y",
)

var reNonNameChar = regexp.MustCompile(`[^a-z0-9ñ ]+`)
var reMultiSpace = regexp.MustCompile(` +`)

// Canonicalize: forma canónica para clave de nombre. CONSERVA espacios (separador de
// palabras) y la ñ. lower → saca tildes de vocales → deja solo [a-z0-9ñ y espacio] →
// colapsa espacios → trim.
//
//	"Acme  S.A."  -> "acme s a"
//	"La Plata"    -> "la plata"   (≠ "laplata")
//	"Peña SA"     -> "peña sa"    (≠ "pena sa")
//	"José Pérez"  -> "jose perez"
func Canonicalize(s string) string {
	s = strings.ToLower(s)
	s = accentReplacer.Replace(s)
	s = reNonNameChar.ReplaceAllString(s, " ")
	s = reMultiSpace.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}
