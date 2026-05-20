package text

import "testing"

func TestCanonicalizeName(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"  AGRO LAJITAS  ", "agro lajitas"},
		{"María Ángeles", "maria angeles"},
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
		{"agro lajitas srl", "Agro Lajitas Srl"},
		{"AGRO LAJITAS SRL", "Agro Lajitas Srl"},
		{"juan de la torre", "Juan de la Torre"},
		{"y griega", "Y Griega"},
		{"María Ángeles", "Maria Angeles"},
		{"E.VEDOYA", "E Vedoya"},
		{"J.M. Perez", "J M Perez"},
		{"JIMENES 25-26", "Jimenes 25 26"},
		{"LOTE 1", "Lote 1"},
		{"", ""},
	}
	for _, c := range cases {
		got := FormatProperName(c.in)
		if got != c.want {
			t.Errorf("FormatProperName(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
