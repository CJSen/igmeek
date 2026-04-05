package index

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
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
	lastErr error
}

type FileMatchResult struct {
	Found       bool
	Entry       IssueEntry
	Ambiguous   bool
	Candidates  []IssueEntry
	Suggestions []IssueEntry
}

type TagCache struct {
	repoDir string
}

func NewIssueIndex(repoDir string) *IssueIndex {
	return &IssueIndex{repoDir: repoDir}
}

func (i *IssueIndex) LastError() error {
	return i.lastErr
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
	i.lastErr = nil
	issues, err := i.Load()
	if err != nil {
		i.lastErr = err
		return IssueEntry{}, false
	}

	for _, issue := range issues {
		if issue.IssueNumber == number {
			return issue, true
		}
	}

	return IssueEntry{}, false
}

func (i *IssueIndex) FindByFilePath(path string) (IssueEntry, bool, error) {
	i.lastErr = nil
	result, err := i.FindFileMatches(path)
	if err != nil {
		i.lastErr = err
		return IssueEntry{}, false, err
	}
	if !result.Found {
		return IssueEntry{}, false, nil
	}

	return result.Entry, true, nil
}

func (i *IssueIndex) FindFileMatches(input string) (FileMatchResult, error) {
	i.lastErr = nil
	issues, err := i.Load()
	if err != nil {
		i.lastErr = err
		return FileMatchResult{}, err
	}

	normalizedInput := normalizeMatchPath(input)
	for _, issue := range issues {
		if normalizeMatchPath(issue.FilePath) == normalizedInput {
			return FileMatchResult{Found: true, Entry: issue}, nil
		}
	}

	base := matchBase(normalizedInput)
	var candidates []IssueEntry
	for _, issue := range issues {
		if matchBase(normalizeMatchPath(issue.FilePath)) == base {
			candidates = append(candidates, issue)
		}
	}

	if len(candidates) == 1 {
		return FileMatchResult{Found: true, Entry: candidates[0]}, nil
	}
	if len(candidates) > 1 {
		return FileMatchResult{Ambiguous: true, Candidates: candidates}, nil
	}

	return FileMatchResult{Suggestions: findSuggestions(base, issues)}, nil
}

func normalizeMatchPath(input string) string {
	normalized := strings.TrimSpace(input)
	normalized = strings.ReplaceAll(normalized, "\\", "/")
	normalized = path.Clean(normalized)
	if normalized == "." {
		return ""
	}
	return normalized
}

func matchBase(input string) string {
	if input == "" {
		return ""
	}
	return path.Base(input)
}

func findSuggestions(targetBase string, issues []IssueEntry) []IssueEntry {
	if targetBase == "" {
		return nil
	}

	type scoredIssue struct {
		entry    IssueEntry
		distance int
	}

	limit := len(targetBase) / 3
	if limit < 2 {
		limit = 2
	}

	var scored []scoredIssue
	for _, issue := range issues {
		distance := levenshteinDistance(targetBase, matchBase(normalizeMatchPath(issue.FilePath)))
		if distance <= limit {
			scored = append(scored, scoredIssue{entry: issue, distance: distance})
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		if scored[i].distance != scored[j].distance {
			return scored[i].distance < scored[j].distance
		}
		return scored[i].entry.FilePath < scored[j].entry.FilePath
	})

	if len(scored) > 3 {
		scored = scored[:3]
	}

	suggestions := make([]IssueEntry, 0, len(scored))
	for _, item := range scored {
		suggestions = append(suggestions, item.entry)
	}
	return suggestions
}

func levenshteinDistance(a, b string) int {
	if a == b {
		return 0
	}
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	prev := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}

	for i := 1; i <= len(a); i++ {
		curr := make([]int, len(b)+1)
		curr[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}

			curr[j] = minInt(
				prev[j]+1,
				curr[j-1]+1,
				prev[j-1]+cost,
			)
		}
		prev = curr
	}

	return prev[len(b)]
}

func minInt(values ...int) int {
	min := values[0]
	for _, value := range values[1:] {
		if value < min {
			min = value
		}
	}
	return min
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
