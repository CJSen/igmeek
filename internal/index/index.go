package index

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type IssueEntry struct {
	IssueNumber int        `json:"issue_number"`
	FilePath    string     `json:"file_path"`
	Title       string     `json:"title"`
	Labels      []string   `json:"labels"`
	State       string     `json:"state"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
	ClosedAt    *time.Time `json:"closed_at"`
	URL         string     `json:"url"`
	HTMLURL     string     `json:"html_url"`
}

type TagEntry struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type IssueIndex struct {
	repoDir string
}

type TagCache struct {
	repoDir string
}

func NewIssueIndex(repoDir string) *IssueIndex {
	return &IssueIndex{repoDir: repoDir}
}

func NewTagCache(repoDir string) *TagCache {
	return &TagCache{repoDir: repoDir}
}

func (i *IssueIndex) filePath() string {
	return filepath.Join(i.repoDir, "issues_num_name.json")
}

func (i *IssueIndex) Save(issues []IssueEntry) error {
	if issues == nil {
		issues = []IssueEntry{}
	}
	data, err := json.MarshalIndent(issues, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal issues: %w", err)
	}
	return os.WriteFile(i.filePath(), data, 0644)
}

func (i *IssueIndex) Load() ([]IssueEntry, error) {
	data, err := os.ReadFile(i.filePath())
	if err != nil {
		if os.IsNotExist(err) {
			return []IssueEntry{}, nil
		}
		return nil, fmt.Errorf("failed to read issues: %w", err)
	}

	var issues []IssueEntry
	if err := json.Unmarshal(data, &issues); err != nil {
		return nil, fmt.Errorf("failed to parse issues: %w", err)
	}

	return issues, nil
}

func (i *IssueIndex) FindByNumber(number int) (IssueEntry, bool) {
	issues, err := i.Load()
	if err != nil {
		return IssueEntry{}, false
	}

	for _, issue := range issues {
		if issue.IssueNumber == number {
			return issue, true
		}
	}

	return IssueEntry{}, false
}

func (i *IssueIndex) FindByFilePath(path string) (IssueEntry, bool) {
	issues, err := i.Load()
	if err != nil {
		return IssueEntry{}, false
	}

	for _, issue := range issues {
		if issue.FilePath == path {
			return issue, true
		}
	}

	return IssueEntry{}, false
}

func (t *TagCache) filePath() string {
	return filepath.Join(t.repoDir, "tags.json")
}

func (t *TagCache) Save(tags []TagEntry) error {
	if tags == nil {
		tags = []TagEntry{}
	}
	data, err := json.MarshalIndent(tags, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}
	return os.WriteFile(t.filePath(), data, 0644)
}

func (t *TagCache) Load() ([]TagEntry, error) {
	data, err := os.ReadFile(t.filePath())
	if err != nil {
		if os.IsNotExist(err) {
			return []TagEntry{}, nil
		}
		return nil, fmt.Errorf("failed to read tags: %w", err)
	}

	var tags []TagEntry
	if err := json.Unmarshal(data, &tags); err != nil {
		return nil, fmt.Errorf("failed to parse tags: %w", err)
	}

	return tags, nil
}
