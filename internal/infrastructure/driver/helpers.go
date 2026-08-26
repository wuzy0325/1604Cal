package driver

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"cal1604/internal/device"
)

// ---------------------------------------------------------------------------
// 辅助函数（单位转换、通道位图、SCPI 解析）
// ---------------------------------------------------------------------------

func isValidPressureUnit(unit string) bool {
	switch strings.ToLower(strings.TrimSpace(unit)) {
	case "mpa", "kpa", "pa", "bar", "mbar", "psi", "kgf/cm2", "mmhg", "atm", "inhg":
		return true
	default:
		return false
	}
}

// NormalizePressureUnit 将单位字符串规范化为标准大小写形式。
// 实现已上移到 device 包（ports 层），此处保留转发以维持 driver 包内部调用不变。
// application 层应直接调用 device.NormalizePressureUnit，避免依赖 adapters 层。
func NormalizePressureUnit(unit string) string {
	return device.NormalizePressureUnit(unit)
}

func pressureUnitToCode(unit string) (string, bool) {
	m := map[string]string{
		"pa": "1130", "kpa": "1133", "mpa": "1132", "bar": "1105", "mbar": "1104",
		"psi": "1141", "kgf/cm2": "1145", "mmhg": "1134", "atm": "1135",
	}
	code, ok := m[strings.ToLower(strings.TrimSpace(unit))]
	return code, ok
}

// pressureUnitToCode820 将压力单位映射为 LabVIEW 程序已验证的 820 单位值。
// kgf/cm2 的枚举序号虽然是 4，但设备命令参数必须使用 10。
func pressureUnitToCode820(unit string) (string, bool) {
	m := map[string]string{
		"pa": "0", "kpa": "1", "mpa": "2", "psi": "3", "kgf/cm2": "10",
	}
	code, ok := m[strings.ToLower(strings.TrimSpace(unit))]
	return code, ok
}

func pressureUnitToCodeSPC4000(unit string) (string, bool) {
	m := map[string]string{
		"psi": "1", "atm": "13", "bar": "14", "mbar": "15", "mmhg": "19",
		"kpa": "22", "pa": "23", "kgf/cm2": "26", "mpa": "36",
	}
	code, ok := m[strings.ToLower(strings.TrimSpace(unit))]
	return code, ok
}

// parseConSTGeneralUnit 解析 ConST 811A/860 等通用型号的单位查询响应。
// 支持 SCPI 标准单位代码（1132=MPa 等）和字符串格式。
func parseConSTGeneralUnit(resp string) string {
	trimmed := strings.TrimSpace(resp)
	m := map[string]string{
		"1130": "Pa", "1133": "kPa", "1132": "MPa",
		"1105": "bar", "1104": "mbar", "1141": "psi",
		"1145": "kgf/cm2", "1134": "mmHg", "1135": "atm",
	}
	if unit, ok := m[trimmed]; ok {
		return unit
	}
	if isValidPressureUnit(trimmed) {
		return NormalizePressureUnit(trimmed)
	}
	return "MPa"
}

// parseConST820Unit 解析 ConST 820 的单位查询响应。
// 820 的 UNIT:PRESsure? 实际返回字符串单位（PA/KPA/MPA/PSI/BAR/MBAR...），
// 而非数字码，故直接按字符串规范化即可。
func parseConST820Unit(resp string) string {
	trimmed := strings.TrimSpace(resp)
	if isValidPressureUnit(trimmed) {
		return NormalizePressureUnit(trimmed)
	}
	return "MPa"
}

func channelsToBitmap(channels []int) string {
	var bitmap uint16
	for _, ch := range channels {
		if ch >= 1 && ch <= 16 {
			bitmap |= 1 << (ch - 1)
		}
	}
	return fmt.Sprintf("%04X", bitmap)
}

func coefficientToUnit(coef string) string {
	v, err := strconv.ParseFloat(coef, 64)
	if err != nil {
		return coef
	}
	switch {
	case v == 1.0:
		return "psi"
	case approxEqual(v, 0.07031):
		return "kgf/cm2"
	case approxEqual(v, 0.0689476):
		return "bar"
	case approxEqual(v, 68.9476):
		return "mbar"
	case approxEqual(v, 6.89476):
		return "kPa"
	case approxEqual(v, 0.00689476):
		return "MPa"
	case approxEqual(v, 51.7149):
		return "mmHg"
	case approxEqual(v, 0.068046):
		return "atm"
	case approxEqual(v, 6894.76):
		return "Pa"
	default:
		return coef
	}
}

func unitToCoefficient(unit string) (string, bool) {
	coefficients := map[string]float64{
		"psi": 1.0, "kgf/cm2": 0.07031, "bar": 0.0689476, "mbar": 68.9476,
		"kpa": 6.89476, "mpa": 0.00689476, "mmhg": 51.7149, "atm": 0.068046, "pa": 6894.76,
	}
	v, ok := coefficients[strings.ToLower(strings.TrimSpace(unit))]
	if !ok {
		return "", false
	}
	return fmt.Sprintf("%g", v), true
}

func parseSCPIPressure(resp string) (float64, error) {
	resp = strings.TrimSpace(resp)
	parts := strings.SplitN(resp, ",", 2)
	v, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return 0, fmt.Errorf("parse pressure value %q: %w", resp, err)
	}
	return v, nil
}

func parseTargetRange(resp string) (min, max float64, err error) {
	parts := strings.Split(resp, ",")
	if len(parts) < 2 {
		return 0, 0, fmt.Errorf("invalid target range response: %q", resp)
	}
	min, err = strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parse min: %w", err)
	}
	max, err = strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parse max: %w", err)
	}
	return min, max, nil
}

func parseWTN1604Unit(response string) (string, bool) {
	val := strings.TrimSpace(response)
	if strings.HasPrefix(val, "A") {
		val = strings.TrimSpace(strings.TrimPrefix(val, "A"))
	}
	v, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return "", false
	}
	// 先尝试系数匹配，避免 bar/atm 等非整数码因舍入误判为 kgf/cm2
	if unit, ok := matchCoefficientToUnit(v); ok {
		return unit, true
	}
	// 回退到整数码匹配（设备原生支持的 0=kgf/cm2, 1=psi, 6=kPa, 6894=Pa）
	unitInt := int(math.Round(v))
	switch unitInt {
	case 0:
		return "kgf/cm2", true
	case 1:
		return "psi", true
	case 6:
		return "kPa", true
	case 6894:
		return "Pa", true
	default:
		return "", false
	}
}

// matchCoefficientToUnit 按系数值匹配单位，与 coefficientToUnit 逻辑一致但返回 bool。
func matchCoefficientToUnit(v float64) (string, bool) {
	switch {
	case v == 1.0:
		return "psi", true
	case approxEqual(v, 0.07031):
		return "kgf/cm2", true
	case approxEqual(v, 0.0689476):
		return "bar", true
	case approxEqual(v, 68.9476):
		return "mbar", true
	case approxEqual(v, 6.89476):
		return "kPa", true
	case approxEqual(v, 0.00689476):
		return "MPa", true
	case approxEqual(v, 51.7149):
		return "mmHg", true
	case approxEqual(v, 0.068046):
		return "atm", true
	case approxEqual(v, 6894.76):
		return "Pa", true
	default:
		return "", false
	}
}

func approxEqual(a, b float64) bool {
	if a == 0 && b == 0 {
		return true
	}
	diff := a - b
	avg := (a + b) / 2
	if avg == 0 {
		return diff == 0
	}
	return (diff/avg) < 0.01 && (diff/avg) > -0.01
}
