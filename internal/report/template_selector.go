package report

import "fmt"

// SelectTemplate 根据测点数量与模式返回兼容旧系统的模板文件名。
func SelectTemplate(points int, mode string) (string, error) {
	if points < 2 || points > 6 {
		return "", fmt.Errorf("invalid point count: %d", points)
	}

	suffix, err := modeSuffix(mode)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%d%s.xlsx", points, suffix), nil
}

func modeSuffix(mode string) (string, error) {
	switch mode {
	case "single", "s":
		return "s", nil
	case "roundTrip", "return", "m":
		return "m", nil
	default:
		return "", fmt.Errorf("invalid pressure mode: %s", mode)
	}
}
