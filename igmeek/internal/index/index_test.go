package index

import (
	"testing"
)

func TestIssueIndexSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	index := NewIssueIndex(tmpDir)

	issues := []IssueEntry{
		{
			IssueNumber: 1,
			FilePath:    "/path/to/post.md",
			Title:       "Test Post",
			Labels:      []string{"tech"},
			State:       "open",
		},
	}

	if err := index.Save(issues); err != nil {
		t.Fatalf("failed to save index: %v", err)
	}

	loaded, err := index.Load()
	if err != nil {
		t.Fatalf("failed to load index: %v", err)
	}

	if len(loaded) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(loaded))
	}
	if loaded[0].Title != "Test Post" {
		t.Errorf("expected title 'Test Post', got %s", loaded[0].Title)
	}
}

func TestIssueIndexFindByNumber(t *testing.T) {
	tmpDir := t.TempDir()
	index := NewIssueIndex(tmpDir)

	issues := []IssueEntry{
		{IssueNumber: 1, Title: "First", State: "open"},
		{IssueNumber: 2, Title: "Second", State: "closed"},
	}
	index.Save(issues)

	found, ok := index.FindByNumber(2)
	if !ok {
		t.Fatal("expected to find issue #2")
	}
	if found.Title != "Second" {
		t.Errorf("expected title 'Second', got %s", found.Title)
	}

	_, ok = index.FindByNumber(99)
	if ok {
		t.Error("expected not to find issue #99")
	}
}

func TestIssueIndexFindByFilePath(t *testing.T) {
	tmpDir := t.TempDir()
	index := NewIssueIndex(tmpDir)

	issues := []IssueEntry{
		{IssueNumber: 1, FilePath: "/path/to/post.md", Title: "Test"},
	}
	index.Save(issues)

	found, ok := index.FindByFilePath("/path/to/post.md")
	if !ok {
		t.Fatal("expected to find issue by path")
	}
	if found.IssueNumber != 1 {
		t.Errorf("expected issue number 1, got %d", found.IssueNumber)
	}
}

func TestTagCacheSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	cache := NewTagCache(tmpDir)

	tags := []TagEntry{
		{Name: "tech", Color: "0075ca"},
		{Name: "blog", Color: "0075ca"},
	}

	if err := cache.Save(tags); err != nil {
		t.Fatalf("failed to save tags: %v", err)
	}

	loaded, err := cache.Load()
	if err != nil {
		t.Fatalf("failed to load tags: %v", err)
	}

	if len(loaded) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(loaded))
	}
}

func TestIssueIndexEmptyLoad(t *testing.T) {
	tmpDir := t.TempDir()
	index := NewIssueIndex(tmpDir)

	loaded, err := index.Load()
	if err != nil {
		t.Fatalf("expected no error for empty index, got %v", err)
	}
	if loaded == nil {
		t.Fatal("expected empty slice, not nil")
	}
	if len(loaded) != 0 {
		t.Errorf("expected 0 issues, got %d", len(loaded))
	}
}
