package text

import "testing"

func TestCanonicalizeName(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"  AGRO LAJITAS  ", "agro lajitas"},
		{"María Ángeles", "maria angeles"},
		{"EL SUEÑO", "el sueño"},
		{"ÑANDÚ", "ñandu"},
		{"AGRO LAJITAS S.R.L.", "agro lajitas s r l"},
		{"JIMENES 25-26", "jimenes 25 26"},
		{"E.VEDOYA", "e vedoya"},
		{"J.M. PEREZ", "j m perez"},
		{"  doble   espacio  ", "doble espacio"},
		{"", ""},
		{"///", ""},
	}
	for _, c := range cases {
		got := CanonicalizeName(c.in)
		if got != c.want {
			t.Errorf("CanonicalizeName(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestFormatProperName(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"agro lajitas srl", "Agro Lajitas SRL"},
		{"AGRO LAJITAS SRL", "Agro Lajitas SRL"},
		{"juan de la torre", "Juan de la Torre"},
		{"y griega", "Y Griega"},
		{"María Ángeles", "Maria Angeles"},
		{"EL SUEÑO", "El Sueño"},
		{"ÑANDÚ", "Ñandu"},
		{"E.VEDOYA", "E Vedoya"},
		{"J.M. Perez", "J M Perez"},
		{"JIMENES 25-26", "Jimenes 25 26"},
		{"LOTE 1", "Lote 1"},
		{"soalen sa", "Soalen SA"},
		{"perez y gomez sh", "Perez y Gomez SH"},
		{"inta pergamino", "INTA Pergamino"},
		{"", ""},
	}
	for _, c := range cases {
		got := FormatProperName(c.in)
		if got != c.want {
			t.Errorf("FormatProperName(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
