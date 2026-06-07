package identity

import "strings"

// PartyType clasifica al actor.
type PartyType string

const (
	Org     PartyType = "org"
	Person  PartyType = "person"
	Unknown PartyType = "unknown"
)

// legalFormMap mapea la forma jurídica (tokens canónicos juntos, sin espacios) a su
// forma canónica. Se construye de los datos reales (SA/SRL/SAS/SH) + variantes habladas.
var legalFormMap = map[string]string{
	"sa": "SA", "sociedadanonima": "SA", "sacif": "SA", "saci": "SA",
	"srl": "SRL", "sociedadderesponsabilidadlimitada": "SRL",
	"sas": "SAS",
	"sh": "SH", "sociedaddehecho": "SH",
	"coop": "COOP", "cooperativa": "COOP",
	"ute": "UTE",
}

// Parsed es el resultado de parsear un nombre de identidad.
type Parsed struct {
	PartyType PartyType
	Core      string // nombre canónico SIN la forma jurídica (conserva espacios)
	Form      string // forma jurídica canónica (SA/SRL/...) o "" si no hay
}

// KeyType devuelve el tipo de clave de nombre para actor_keys.
func (p Parsed) KeyType() string {
	if p.Form != "" {
		return "LEGAL_NAME"
	}
	return "PERSON_NAME"
}

// KeyValue devuelve el valor de clave de nombre. Para org: "core|FORMA"
// (ej. "acme|SA"); para persona/unknown: el core ("juan perez").
func (p Parsed) KeyValue() string {
	if p.Form != "" {
		return p.Core + "|" + p.Form
	}
	return p.Core
}

// ParseLegalName canonicaliza el nombre y detecta la forma jurídica al final (por
// tokens, tolerante a puntos/espacios: "S.A." = "S A" = "SA"). El core conserva
// espacios → "La Plata SA" (la plata|SA) ≠ "Laplata SA" (laplata|SA); y
// "Acme SA" (acme|SA) ≠ "Acme SRL" (acme|SRL). La ñ se conserva.
func ParseLegalName(raw string) Parsed {
	c := Canonicalize(raw)
	if c == "" {
		return Parsed{PartyType: Unknown}
	}
	tokens := strings.Fields(c)

	// Probar la forma jurídica tomando los últimos k tokens (k = 3,2,1), juntos sin
	// espacios. El primer match con core NO vacío gana.
	for k := 3; k >= 1; k-- {
		if k >= len(tokens) {
			continue // dejar al menos 1 token de core
		}
		joined := strings.Join(tokens[len(tokens)-k:], "")
		if form, ok := legalFormMap[joined]; ok {
			core := strings.Join(tokens[:len(tokens)-k], " ")
			return Parsed{PartyType: Org, Core: core, Form: form}
		}
	}

	// Sin forma jurídica detectable → persona/unknown; el core es el nombre completo.
	return Parsed{PartyType: Unknown, Core: c, Form: ""}
}
