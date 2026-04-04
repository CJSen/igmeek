package markdown

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type MarkdownFile struct {
	Content string
	AbsPath string
	Title   string
}

func ReadFile(path string) (*MarkdownFile, error) {
	absPath, err := NormalizePath(path)
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	contentStr := string(content)
	title := ExtractTitle(contentStr)
	if title == "" {
		title = ExtractTitleFromFileName(absPath)
	}

	return &MarkdownFile{
		Content: contentStr,
		AbsPath: absPath,
		Title:   title,
	}, nil
}

func NormalizePath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return filepath.Clean(path), nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	return filepath.Clean(filepath.Join(wd, path)), nil
}

func ExtractTitle(content string) string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
		if line != "" && !strings.HasPrefix(line, "<") {
			break
		}
	}
	return ""
}

func ExtractTitleFromFileName(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	if ext != "" {
		base = base[:len(base)-len(ext)]
	}
	return base
}
