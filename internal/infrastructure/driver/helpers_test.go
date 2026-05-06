package driver

import "testing"

func TestNormalizePressureUnit(t *testing.T) {
	cases := []struct{ input, want string }{
		{"kpa", "kPa"},
		{"KPA", "kPa"},
		{"kPa", "kPa"},
		{"mpa", "MPa"},
		{"MPa", "MPa"},
		{"pa", "Pa"},
		{"bar", "bar"},
		{"psi", "psi"},
		{"mmhg", "mmHg"},
		{"mmHg", "mmHg"},
		{"atm", "atm"},
		{"inhg", "inHg"},
		{"kgf/cm2", "kgf/cm2"},
	}
	for _, c := range cases {
		got := NormalizePressureUnit(c.input)
		if got != c.want {
			t.Errorf("NormalizePressureUnit(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestParseConSTGeneralUnit_NormalizesCase(t *testing.T) {
	cases := []struct{ input, want string }{
		{"1133", "kPa"},
		{"1132", "MPa"},
		{"kpa", "kPa"},
		{"mpa", "MPa"},
		{"mmhg", "mmHg"},
		{"kPa", "kPa"},
	}
	for _, c := range cases {
		got := parseConSTGeneralUnit(c.input)
		if got != c.want {
			t.Errorf("parseConSTGeneralUnit(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestParseConST820Unit_NormalizesCase(t *testing.T) {
	cases := []struct{ input, want string }{
		{"1", "kPa"},
		{"2", "MPa"},
		{"kpa", "kPa"},
		{"mpa", "MPa"},
		{"mmhg", "mmHg"},
	}
	for _, c := range cases {
		got := parseConST820Unit(c.input)
		if got != c.want {
			t.Errorf("parseConST820Unit(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}
