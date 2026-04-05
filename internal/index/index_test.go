package index

import (
	"os"
	"reflect"
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

func TestIssueIndexFindByNumberReturnsLoadError(t *testing.T) {
	tmpDir := t.TempDir()
	index := NewIssueIndex(tmpDir)

	if err := os.WriteFile(index.filePath(), []byte("{"), 0644); err != nil {
		t.Fatalf("failed to write invalid index: %v", err)
	}

	if _, ok := index.FindByNumber(1); ok {
		t.Fatal("expected invalid index not to return a match")
	}
	if err := index.LastError(); err == nil {
		t.Fatal("expected FindByNumber to expose load error")
	}
}

func TestIssueIndexFindByFilePath(t *testing.T) {
	tmpDir := t.TempDir()
	index := NewIssueIndex(tmpDir)

	issues := []IssueEntry{
		{IssueNumber: 1, FilePath: "/path/to/post.md", Title: "Test"},
	}
	index.Save(issues)

	found, ok, err := index.FindByFilePath("/path/to/post.md")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !ok {
		t.Fatal("expected to find issue by path")
	}
	if found.IssueNumber != 1 {
		t.Errorf("expected issue number 1, got %d", found.IssueNumber)
	}
}

func TestIssueIndexFindFileMatchesBasenameFallback(t *testing.T) {
	tmpDir := t.TempDir()
	index := NewIssueIndex(tmpDir)

	issues := []IssueEntry{
		{IssueNumber: 1, FilePath: "posts/post.md", Title: "Post"},
	}
	if err := index.Save(issues); err != nil {
		t.Fatalf("failed to save index: %v", err)
	}

	result, err := index.FindFileMatches("/tmp/work/post.md")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.Found {
		t.Fatal("expected basename fallback to find issue")
	}
	if result.Entry.IssueNumber != 1 {
		t.Fatalf("expected issue number 1, got %d", result.Entry.IssueNumber)
	}
	if result.Ambiguous {
		t.Fatal("expected basename fallback to be unambiguous")
	}
}

func TestIssueIndexFindFileMatchesWindowsSeparators(t *testing.T) {
	tmpDir := t.TempDir()
	index := NewIssueIndex(tmpDir)

	issues := []IssueEntry{
		{IssueNumber: 2, FilePath: `posts\\note.md`, Title: "Note"},
	}
	if err := index.Save(issues); err != nil {
		t.Fatalf("failed to save index: %v", err)
	}

	result, err := index.FindFileMatches(`C:\work\note.md`)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.Found {
		t.Fatal("expected Windows-style path to match")
	}
	if result.Entry.IssueNumber != 2 {
		t.Fatalf("expected issue number 2, got %d", result.Entry.IssueNumber)
	}
}

func TestIssueIndexFindFileMatchesAmbiguousBasename(t *testing.T) {
	tmpDir := t.TempDir()
	index := NewIssueIndex(tmpDir)

	issues := []IssueEntry{
		{IssueNumber: 1, FilePath: "posts/shared.md", Title: "Post"},
		{IssueNumber: 2, FilePath: "notes/shared.md", Title: "Note"},
	}
	if err := index.Save(issues); err != nil {
		t.Fatalf("failed to save index: %v", err)
	}

	result, err := index.FindFileMatches("/tmp/work/shared.md")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Found {
		t.Fatal("expected ambiguous basename not to auto-resolve")
	}
	if !result.Ambiguous {
		t.Fatal("expected ambiguous result")
	}
	if len(result.Candidates) != 2 {
		t.Fatalf("expected 2 candidates, got %d", len(result.Candidates))
	}
	if !reflect.DeepEqual([]string{result.Candidates[0].FilePath, result.Candidates[1].FilePath}, []string{"posts/shared.md", "notes/shared.md"}) {
		t.Fatalf("unexpected candidates: %#v", result.Candidates)
	}
	if _, ok, err := index.FindByFilePath("/tmp/work/shared.md"); err != nil || ok {
		t.Fatal("expected FindByFilePath to reject ambiguous basename")
	}
}

func TestIssueIndexFindFileMatchesSuggestionsWhenMissing(t *testing.T) {
	tmpDir := t.TempDir()
	index := NewIssueIndex(tmpDir)

	issues := []IssueEntry{
		{IssueNumber: 1, FilePath: "posts/post.md", Title: "Post"},
		{IssueNumber: 2, FilePath: "notes/guide.md", Title: "Guide"},
	}
	if err := index.Save(issues); err != nil {
		t.Fatalf("failed to save index: %v", err)
	}

	result, err := index.FindFileMatches("/tmp/work/posta.md")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Found {
		t.Fatal("expected missing file not to be found")
	}
	if result.Ambiguous {
		t.Fatal("expected missing file not to be ambiguous")
	}
	if len(result.Suggestions) == 0 {
		t.Fatal("expected similar file suggestions")
	}
	if result.Suggestions[0].FilePath != "posts/post.md" {
		t.Fatalf("expected posts/post.md to be suggested first, got %#v", result.Suggestions)
	}
}

func TestIssueIndexFindFileMatchesReturnsLoadError(t *testing.T) {
	tmpDir := t.TempDir()
	index := NewIssueIndex(tmpDir)

	if err := os.WriteFile(index.filePath(), []byte("{"), 0644); err != nil {
		t.Fatalf("failed to write invalid index: %v", err)
	}

	if _, err := index.FindFileMatches("posts/post.md"); err == nil {
		t.Fatal("expected load error from FindFileMatches")
	}

	if _, _, err := index.FindByFilePath("posts/post.md"); err == nil {
		t.Fatal("expected load error from FindByFilePath")
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
