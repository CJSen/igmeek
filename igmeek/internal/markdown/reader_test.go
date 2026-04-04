package markdown

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")
	content := "# Hello World\n\nThis is a test."
	os.WriteFile(testFile, []byte(content), 0644)

	result, err := ReadFile(testFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Content != content {
		t.Errorf("expected content %q, got %q", content, result.Content)
	}
}

func TestReadFileNotFound(t *testing.T) {
	_, err := ReadFile("/nonexistent/file.md")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestNormalizePath(t *testing.T) {
	abs, err := NormalizePath("relative/path.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !filepath.IsAbs(abs) {
		t.Errorf("expected absolute path, got %s", abs)
	}
}

func TestNormalizePathAlreadyAbs(t *testing.T) {
	absPath := "/absolute/path/file.md"
	result, err := NormalizePath(absPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != absPath {
		t.Errorf("expected %s, got %s", absPath, result)
	}
}

func TestExtractTitle(t *testing.T) {
	content := "# My Blog Post\n\nSome content here."
	title := ExtractTitle(content)
	if title != "My Blog Post" {
		t.Errorf("expected 'My Blog Post', got %q", title)
	}
}

func TestExtractTitleNoHeading(t *testing.T) {
	content := "Some content without heading."
	title := ExtractTitle(content)
	if title != "" {
		t.Errorf("expected empty title, got %q", title)
	}
}

func TestExtractTitleFromFileName(t *testing.T) {
	title := ExtractTitleFromFileName("/path/to/my-blog-post.md")
	if title != "my-blog-post" {
		t.Errorf("expected 'my-blog-post', got %q", title)
	}
}

func TestExtractTitleFromFileNameNoExt(t *testing.T) {
	title := ExtractTitleFromFileName("/path/to/my-post")
	if title != "my-post" {
		t.Errorf("expected 'my-post', got %q", title)
	}
}
