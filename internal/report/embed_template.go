package report

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/xuri/excelize/v2"
)

// EmbedTemplateProvider 从 embed.FS 提供模板文件访问。
// 将嵌入的模板在首次使用时解压到临时目录，然后以普通文件形式使用（兼容 excelize）。
type EmbedTemplateProvider struct {
	embedFS   fs.FS
	prefix    string
	tempDir   string
	initOnce  sync.Once
	initErr   error
}

// NewEmbedTemplateProvider 创建嵌入模板提供者。
// prefix 是 embed.FS 内的模板目录前缀，如 "templates/reports"。
func NewEmbedTemplateProvider(embedFS fs.FS, prefix string) *EmbedTemplateProvider {
	return &EmbedTemplateProvider{
		embedFS: embedFS,
		prefix:  prefix,
	}
}

// TempDir 返回解压后的临时目录路径（首次调用会触发解压）。
func (p *EmbedTemplateProvider) TempDir() (string, error) {
	p.initOnce.Do(func() {
		p.initErr = p.extractAll()
	})
	return p.tempDir, p.initErr
}

// ResolvePath 根据模板文件名返回解压后的绝对路径。
func (p *EmbedTemplateProvider) ResolvePath(filename string) (string, error) {
	dir, err := p.TempDir()
	if err != nil {
		return "", err
	}
	fullPath := filepath.Join(dir, filename)
	if _, err := os.Stat(fullPath); err != nil {
		return "", fmt.Errorf("template not found in embed: %s", filename)
	}
	return fullPath, nil
}

// ListTemplates 返回所有可用的模板文件名列表。
func (p *EmbedTemplateProvider) ListTemplates() ([]string, error) {
	dir, err := p.TempDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.EqualFold(filepath.Ext(entry.Name()), ".xlsx") {
			files = append(files, entry.Name())
		}
	}
	return files, nil
}

// LoadTemplateFromEmbed 直接从 embed.FS 加载模板为 excelize.File。
// 适用于不需要文件路径的场景。
func (p *EmbedTemplateProvider) LoadTemplateFromEmbed(filename string) (*excelize.File, error) {
	path := filepath.Join(p.prefix, filename)
	f, err := p.embedFS.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open embed template %s: %w", filename, err)
	}
	defer f.Close()

	// excelize.OpenReader 需要 io.ReaderAt，先将内容读入内存
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read embed template %s: %w", filename, err)
	}
	return excelize.OpenReader(bytes.NewReader(data))
}

// extractAll 将 embed.FS 中的模板文件全部解压到临时目录。
func (p *EmbedTemplateProvider) extractAll() error {
	tempDir, err := os.MkdirTemp("", "cal1604-templates-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	p.tempDir = tempDir

	return fs.WalkDir(p.embedFS, p.prefix, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		// 计算相对路径
		relPath, err := filepath.Rel(p.prefix, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(tempDir, relPath)

		// 确保目标目录存在
		if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			return err
		}

		// 复制文件
		src, err := p.embedFS.Open(path)
		if err != nil {
			return err
		}
		defer src.Close()

		dest, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer dest.Close()

		if _, err := io.Copy(dest, src); err != nil {
			return err
		}
		return nil
	})
}

// Cleanup 删除临时目录（应用退出时调用）。
func (p *EmbedTemplateProvider) Cleanup() error {
	if p.tempDir != "" {
		return os.RemoveAll(p.tempDir)
	}
	return nil
}
