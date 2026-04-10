package report

import "path/filepath"

// Service 封装报告模板路径拼装逻辑。
type Service struct {
	templateDir string
}

// NewService 创建报告服务。
func NewService(templateDir string) *Service {
	return &Service{templateDir: templateDir}
}

// ResolveTemplatePath 解析模板绝对路径。
func (s *Service) ResolveTemplatePath(points int, mode string) (string, error) {
	filename, err := SelectTemplate(points, mode)
	if err != nil {
		return "", err
	}

	return filepath.Join(s.templateDir, filename), nil
}
