package identity

import (
	"regexp"
	"strings"
)

var reNonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

// NormalizeTaxID deja solo alfanumérico en minúscula (CUIT/CUIL/DNI sin guiones,
// puntos ni espacios). "20-12345678-6" -> "20123456786". Es el key_value de TAX_ID.
func NormalizeTaxID(s string) string {
	return reNonAlnum.ReplaceAllString(strings.ToLower(s), "")
}

// ValidCUIT valida un CUIT/CUIL argentino: 11 dígitos + dígito verificador mod-11.
// Es validación ADVISORY (el resolver puede usar NormalizeTaxID como clave aunque
// el checksum no valide, p.ej. identificadores extranjeros) — útil para el front.
func ValidCUIT(s string) bool {
	n := NormalizeTaxID(s)
	if len(n) != 11 {
		return false
	}
	for i := 0; i < 11; i++ {
		if n[i] < '0' || n[i] > '9' {
			return false
		}
	}
	mult := [10]int{5, 4, 3, 2, 7, 6, 5, 4, 3, 2}
	sum := 0
	for i := 0; i < 10; i++ {
		sum += int(n[i]-'0') * mult[i]
	}
	check := 11 - (sum % 11)
	switch check {
	case 11:
		check = 0
	case 10:
		check = 9
	}
	return check == int(n[10]-'0')
}
