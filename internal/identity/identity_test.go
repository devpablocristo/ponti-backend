package identity

import "testing"

func TestCanonicalize(t *testing.T) {
	cases := map[string]string{
		"Acme  S.A.":  "acme s a",
		"La Plata":    "la plata",
		"Laplata":     "laplata",
		"Peña SA":     "peña sa",
		"Pena SA":     "pena sa",
		"Niño":        "niño",
		"José Pérez":  "jose perez",
		"  ÑANDÚ  ":   "ñandu",
		"":            "",
		"///":         "",
	}
	for in, want := range cases {
		if got := Canonicalize(in); got != want {
			t.Errorf("Canonicalize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestParseLegalName(t *testing.T) {
	cases := []struct {
		raw   string
		party PartyType
		key   string // KeyValue esperado
	}{
		{"Acme S.A.", Org, "acme|SA"},
		{"Acme SA", Org, "acme|SA"},
		{"ACME  S.A.", Org, "acme|SA"},
		{"Acme Sociedad Anónima", Org, "acme|SA"},
		{"Acme SRL", Org, "acme|SRL"},
		{"Acme S.R.L.", Org, "acme|SRL"},
		{"Acme S.A.S.", Org, "acme|SAS"},
		{"La Plata SA", Org, "la plata|SA"},
		{"Laplata SA", Org, "laplata|SA"},
		{"Peña SA", Org, "peña|SA"},
		{"Pena SA", Org, "pena|SA"},
		{"Juan Pérez", Unknown, "juan perez"},
		{"Savino", Unknown, "savino"}, // NO debe detectar "sa" dentro de "savino"
	}
	for _, c := range cases {
		p := ParseLegalName(c.raw)
		if p.PartyType != c.party {
			t.Errorf("ParseLegalName(%q).PartyType = %q, want %q", c.raw, p.PartyType, c.party)
		}
		if got := p.KeyValue(); got != c.key {
			t.Errorf("ParseLegalName(%q).KeyValue = %q, want %q", c.raw, got, c.key)
		}
	}

	// Cross-checks clave: distinguen lo que deben distinguir y unifican lo que deben.
	if ParseLegalName("Acme SA").KeyValue() == ParseLegalName("Acme SRL").KeyValue() {
		t.Error("Acme SA y Acme SRL NO deben colisionar")
	}
	if ParseLegalName("La Plata SA").KeyValue() == ParseLegalName("Laplata SA").KeyValue() {
		t.Error("La Plata SA y Laplata SA NO deben colisionar")
	}
	if ParseLegalName("Peña SA").KeyValue() == ParseLegalName("Pena SA").KeyValue() {
		t.Error("Peña y Pena NO deben colisionar (ñ)")
	}
	if ParseLegalName("Acme S.A.").KeyValue() != ParseLegalName("Acme Sociedad Anónima").KeyValue() {
		t.Error("Acme S.A. y Acme Sociedad Anónima DEBEN unificar")
	}
}

func TestTaxID(t *testing.T) {
	if NormalizeTaxID("20-12345678-6") != "20123456786" {
		t.Errorf("NormalizeTaxID falló: %q", NormalizeTaxID("20-12345678-6"))
	}
	if NormalizeTaxID("20.123.456/78-6") != "20123456786" {
		t.Errorf("NormalizeTaxID con puntos/barras falló: %q", NormalizeTaxID("20.123.456/78-6"))
	}
	if !ValidCUIT("20123456786") {
		t.Error("20123456786 debería ser CUIT válido")
	}
	if !ValidCUIT("20-12345678-6") {
		t.Error("20-12345678-6 (con guiones) debería ser válido")
	}
	if ValidCUIT("20123456789") {
		t.Error("20123456789 (check incorrecto) NO debería ser válido")
	}
	if ValidCUIT("123") {
		t.Error("123 (corto) NO debería ser válido")
	}
}
