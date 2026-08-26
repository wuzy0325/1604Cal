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
		{"1131", "Pa"},
		{"1130", "Pa"},
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

func TestPressureUnitToCode811AUsesDevicePaCode(t *testing.T) {
	got, ok := pressureUnitToCode811A("Pa")
	if !ok || got != "1131" {
		t.Fatalf("pressureUnitToCode811A(Pa) = %q, %v; want 1131, true", got, ok)
	}
}

func TestParseConST820Unit_NormalizesCase(t *testing.T) {
	// 820 的 UNIT:PRESsure? 实机返回字符串单位（大写），并非数字码。
	cases := []struct{ input, want string }{
		{"PA", "Pa"},
		{"KPA", "kPa"},
		{"MPA", "MPa"},
		{"PSI", "psi"},
		{"BAR", "bar"},
		{"MBAR", "mbar"},
		{"kpa", "kPa"},
		{"mpa", "MPa"},
	}
	for _, c := range cases {
		got := parseConST820Unit(c.input)
		if got != c.want {
			t.Errorf("parseConST820Unit(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestPressureUnitToCode820UsesLabVIEWValues(t *testing.T) {
	cases := map[string]string{
		"Pa": "0", "kPa": "1", "MPa": "2", "psi": "3", "kgf/cm2": "10",
	}
	for unit, want := range cases {
		got, ok := pressureUnitToCode820(unit)
		if !ok || got != want {
			t.Errorf("pressureUnitToCode820(%q) = %q, %v; want %q, true", unit, got, ok, want)
		}
	}
}
