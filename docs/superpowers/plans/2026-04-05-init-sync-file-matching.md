# Init Sync And File Matching Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让 `igmeek init` 在首次初始化时同时保存 token 和初始 repo、自动执行一次 `sync` 并输出结果与数据路径，同时修复 `file` 参数在绝对路径/相对路径及跨平台路径场景下无法匹配 issue 的问题。

**Architecture:** 继续沿用现有 cobra + internal 分层，不改配置 schema。把 repo 输入规范化和 repo 列表去重放进 `internal/config`，把同步逻辑提炼为 `cmd` 层可复用 helper 以便 `init` 直接调用；把文件匹配增强收敛到 `internal/index`，由 `update` 命令消费统一的匹配结果和建议信息。

**Tech Stack:** Go, cobra, go-github, standard library (`net/url`, `path/filepath`, `strings`, `fmt`, `bytes`, `os`)

---

## File Structure Map

| File | Responsibility |
|---|---|
| `internal/config/config.go` | repo 输入规范化、repo 列表追加去重、全局配置路径工具 |
| `internal/config/config_test.go` | repo 规范化和 repo 列表写入测试 |
| `cmd/sync.go` | 提供可复用的同步执行 helper，供 `sync` 和 `init` 共用 |
| `cmd/init.go` | 交互读取 token/repo、保存配置、自动触发 sync、输出结果与路径 |
| `cmd/init_test.go` | `init` 成功和失败场景测试 |
| `internal/index/index.go` | 路径规范化、basename 匹配、歧义检测、相似文件建议 |
| `internal/index/index_test.go` | 精确路径、basename 回退、Windows 路径、歧义、建议测试 |
| `cmd/update.go` | 消费新的索引匹配结果，输出更友好的错误信息 |

---

### Task 1: Repo 输入规范化与配置辅助函数

**Files:**
- Modify: `internal/config/config.go`
- Modify: `internal/config/config_test.go`

- [ ] **Step 1: 先写 repo 规范化与 repo 列表去重测试**

`internal/config/config_test.go` 追加这些测试：

```go
func TestNormalizeRepoInputOwnerRepo(t *testing.T) {
	fullName, err := NormalizeRepoInput("octo/test")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if fullName != "octo/test" {
		t.Fatalf("expected octo/test, got %s", fullName)
	}
}

func TestNormalizeRepoInputGitHubURL(t *testing.T) {
	fullName, err := NormalizeRepoInput("https://github.com/octo/test")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if fullName != "octo/test" {
		t.Fatalf("expected octo/test, got %s", fullName)
	}
}

func TestNormalizeRepoInputRejectsInvalidValue(t *testing.T) {
	_, err := NormalizeRepoInput("https://example.com/octo/test")
	if err == nil {
		t.Fatal("expected invalid repo input error")
	}
}

func TestAddRepoKeepsUniqueList(t *testing.T) {
	cfg := &GlobalConfig{Repos: []string{"octo/test"}}
	cfg.AddRepo("octo/test")
	cfg.AddRepo("octo/blog")

	if len(cfg.Repos) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(cfg.Repos))
	}
	if cfg.Repos[1] != "octo/blog" {
		t.Fatalf("expected octo/blog at index 1, got %s", cfg.Repos[1])
	}
}
```

- [ ] **Step 2: 运行测试，确认新增测试先失败**

Run: `go test ./internal/config -v`

Expected: FAIL with undefined identifiers such as `NormalizeRepoInput` and `AddRepo`.

- [ ] **Step 3: 在配置模块中实现 repo 规范化和 repo 列表辅助函数**

把 `internal/config/config.go` 调整为包含这些新增函数和 import：

```go
package config

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type GlobalConfig struct {
	Token       string   `json:"token"`
	CurrentRepo string   `json:"current_repo"`
	Repos       []string `json:"repos"`
}

type RepoConfig struct {
	Owner    string `json:"owner"`
	Repo     string `json:"repo"`
	FullName string `json:"full_name"`
}

func GetGlobalDataDir() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = os.TempDir()
	}
	return filepath.Join(configDir, "igmeek")
}

func GetRepoDir(globalDir, fullName string) string {
	safe := strings.ReplaceAll(fullName, "/", "_")
	return filepath.Join(globalDir, "repos", safe)
}

func ConfigPath(globalDir string) string {
	return filepath.Join(globalDir, "config.json")
}

func NormalizeRepoInput(input string) (string, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", fmt.Errorf("repository cannot be empty")
	}

	if strings.Contains(trimmed, "://") {
		u, err := url.Parse(trimmed)
		if err != nil {
			return "", fmt.Errorf("invalid repository: %w", err)
		}
		if !strings.EqualFold(u.Hostname(), "github.com") {
			return "", fmt.Errorf("repository must be a GitHub repository URL or owner/repo")
		}
		path := strings.Trim(strings.TrimSpace(u.Path), "/")
		parts := strings.Split(path, "/")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return "", fmt.Errorf("repository must be in owner/repo format")
		}
		return parts[0] + "/" + parts[1], nil
	}

	parts := strings.Split(trimmed, "/")
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return "", fmt.Errorf("repository must be in owner/repo format")
	}

	return strings.TrimSpace(parts[0]) + "/" + strings.TrimSpace(parts[1]), nil
}

func (c *GlobalConfig) AddRepo(fullName string) {
	for _, repo := range c.Repos {
		if repo == fullName {
			return
		}
	}
	c.Repos = append(c.Repos, fullName)
}

func LoadConfig(path string) (*GlobalConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg GlobalConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

func (c *GlobalConfig) Save(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

func LoadRepoConfig(repoDir string) (*RepoConfig, error) {
	path := filepath.Join(repoDir, "repo.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read repo config: %w", err)
	}

	var rc RepoConfig
	if err := json.Unmarshal(data, &rc); err != nil {
		return nil, fmt.Errorf("failed to parse repo config: %w", err)
	}

	return &rc, nil
}

func (r *RepoConfig) Save(repoDir string) error {
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		return fmt.Errorf("failed to create repo dir: %w", err)
	}

	path := filepath.Join(repoDir, "repo.json")
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal repo config: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

func EnsureGlobalDir(globalDir string) error {
	return os.MkdirAll(globalDir, 0755)
}
```

- [ ] **Step 4: 运行配置测试，确认转绿**

Run: `go test ./internal/config -v`

Expected: PASS for `TestNormalizeRepoInputOwnerRepo`, `TestNormalizeRepoInputGitHubURL`, `TestNormalizeRepoInputRejectsInvalidValue`, `TestAddRepoKeepsUniqueList` and existing config tests.

- [ ] **Step 5: 提交这一小步**

```bash
git add internal/config/config.go internal/config/config_test.go
git commit -m "feat: normalize init repo input"
```

---

### Task 2: 让 `init` 保存 repo 并自动执行一次 sync

**Files:**
- Modify: `cmd/init.go`
- Modify: `cmd/sync.go`
- Create: `cmd/init_test.go`

- [ ] **Step 1: 先写 `init` 的成功与失败测试**

创建 `cmd/init_test.go`：

```go
package cmd

import (
	"bytes"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CJSen/igmeek/cli/internal/config"
	innersync "github.com/CJSen/igmeek/cli/internal/sync"
	"github.com/spf13/cobra"
)

func TestRunInitSavesConfigAndRunsSync(t *testing.T) {
	tmpDir := t.TempDir()
	oldGlobalDirFunc := globalDataDirFunc
	oldRunSyncFunc := runSyncForRepoFunc
	defer func() {
		globalDataDirFunc = oldGlobalDirFunc
		runSyncForRepoFunc = oldRunSyncFunc
	}()

	globalDataDirFunc = func() string { return tmpDir }
	runSyncForRepoFunc = func(cmd *cobra.Command, fullName string) (*innersync.SyncResult, string, error) {
		if fullName != "octo/blog" {
			t.Fatalf("expected octo/blog, got %s", fullName)
		}
		return &innersync.SyncResult{IssuesCount: 3, LabelsCount: 2}, filepath.Join(tmpDir, "repos", "octo_blog"), nil
	}

	cmd := &cobra.Command{}
	in := bytes.NewBufferString("token-123\nhttps://github.com/octo/blog\n")
	out := &bytes.Buffer{}
	cmd.SetIn(in)
	cmd.SetOut(out)
	cmd.SetErr(out)

	if err := runInit(cmd, nil); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	cfg, err := config.LoadConfig(config.ConfigPath(tmpDir))
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if cfg.Token != "token-123" {
		t.Fatalf("expected saved token, got %s", cfg.Token)
	}
	if cfg.CurrentRepo != "octo/blog" {
		t.Fatalf("expected current repo octo/blog, got %s", cfg.CurrentRepo)
	}

	printed := out.String()
	if !strings.Contains(printed, "Synced 3 issues, 2 labels from octo/blog") {
		t.Fatalf("expected sync summary in output, got %q", printed)
	}
	if !strings.Contains(printed, config.ConfigPath(tmpDir)) {
		t.Fatalf("expected config path in output, got %q", printed)
	}
	if !strings.Contains(printed, filepath.Join(tmpDir, "repos", "octo_blog")) {
		t.Fatalf("expected repo dir in output, got %q", printed)
	}
}

func TestRunInitKeepsConfigWhenSyncFails(t *testing.T) {
	tmpDir := t.TempDir()
	oldGlobalDirFunc := globalDataDirFunc
	oldRunSyncFunc := runSyncForRepoFunc
	defer func() {
		globalDataDirFunc = oldGlobalDirFunc
		runSyncForRepoFunc = oldRunSyncFunc
	}()

	globalDataDirFunc = func() string { return tmpDir }
	runSyncForRepoFunc = func(cmd *cobra.Command, fullName string) (*innersync.SyncResult, string, error) {
		return nil, filepath.Join(tmpDir, "repos", "octo_blog"), errors.New("boom")
	}

	cmd := &cobra.Command{}
	in := bytes.NewBufferString("token-123\nocto/blog\n")
	out := &bytes.Buffer{}
	cmd.SetIn(in)
	cmd.SetOut(out)
	cmd.SetErr(out)

	err := runInit(cmd, nil)
	if err == nil {
		t.Fatal("expected sync failure")
	}
	if !strings.Contains(err.Error(), "configuration was saved") {
		t.Fatalf("expected preserved-config guidance, got %v", err)
	}

	cfg, loadErr := config.LoadConfig(config.ConfigPath(tmpDir))
	if loadErr != nil {
		t.Fatalf("expected saved config, got load error %v", loadErr)
	}
	if cfg.CurrentRepo != "octo/blog" {
		t.Fatalf("expected saved current repo, got %s", cfg.CurrentRepo)
	}
}
```

- [ ] **Step 2: 运行 `cmd` 测试，确认新增测试先失败**

Run: `go test ./cmd -run TestRunInit -v`

Expected: FAIL with undefined identifiers such as `globalDataDirFunc` and `runSyncForRepoFunc`, or output/assertion failures because `runInit` still reads from `os.Stdin` and does not run sync.

- [ ] **Step 3: 提取同步 helper，避免 `init` 复制 `sync` 逻辑**

把 `cmd/sync.go` 调整为：

```go
package cmd

import (
	"context"
	"fmt"

	"github.com/CJSen/igmeek/cli/internal/api"
	"github.com/CJSen/igmeek/cli/internal/config"
	innersync "github.com/CJSen/igmeek/cli/internal/sync"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync all remote issues and labels to local cache",
	Long:  "Fetch all open and closed issues from the configured repository and update the local index (issues_num_name.json) and tag cache (tags.json). This is a full sync that overwrites local index data.",
	RunE:  runSync,
}

func init() {
	rootCmd.AddCommand(syncCmd)
}

func runSyncForRepo(fullName string) (*innersync.SyncResult, string, error) {
	globalDir := globalDataDirFunc()
	repoDir := config.GetRepoDir(globalDir, fullName)
	owner, repo, err := innersync.ParseOwnerRepo(fullName)
	if err != nil {
		return nil, repoDir, err
	}

	client := api.NewClient(GetToken())
	result, err := innersync.SyncAll(context.Background(), client, owner, repo, repoDir)
	if err != nil {
		return nil, repoDir, err
	}

	return result, repoDir, nil
}

func runSync(cmd *cobra.Command, args []string) error {
	globalDir := globalDataDirFunc()
	cfg, err := config.LoadConfig(config.ConfigPath(globalDir))
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.CurrentRepo == "" {
		return fmt.Errorf("no repository configured. Run 'igmeek repo add' first")
	}

	result, _, err := runSyncForRepo(cfg.CurrentRepo)
	if err != nil {
		return err
	}

	fmt.Printf("Synced %d issues, %d labels from %s\n", result.IssuesCount, result.LabelsCount, cfg.CurrentRepo)
	return nil
}
```

- [ ] **Step 4: 把 `init` 改造成可测试的交互式初始化流程**

把 `cmd/init.go` 调整为：

```go
package cmd

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/CJSen/igmeek/cli/internal/config"
	innersync "github.com/CJSen/igmeek/cli/internal/sync"
	"github.com/spf13/cobra"
)

var (
	globalDataDirFunc = config.GetGlobalDataDir
	runSyncForRepoFunc = func(cmd *cobra.Command, fullName string) (*innersync.SyncResult, string, error) {
		return runSyncForRepo(fullName)
	}
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize igmeek with your GitHub token and initial repository",
	Long:  "Interactively prompt for a GitHub Personal Access Token and an initial repository, save them to the global configuration file, then run a full sync. The token requires the 'repo' scope. Can also be set via the IMGEEK_GITHUB_TOKEN environment variable.",
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	globalDir := globalDataDirFunc()
	if err := config.EnsureGlobalDir(globalDir); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	reader := bufio.NewReader(cmd.InOrStdin())
	out := cmd.OutOrStdout()

	fmt.Fprint(out, "Enter your GitHub Personal Access Token (needs 'repo' scope): ")
	token, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read token: %w", err)
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return fmt.Errorf("token cannot be empty")
	}

	fmt.Fprint(out, "Enter initial repository (owner/repo or GitHub URL): ")
	repoInput, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read repository: %w", err)
	}

	fullName, err := config.NormalizeRepoInput(repoInput)
	if err != nil {
		return err
	}

	cfgPath := config.ConfigPath(globalDir)
	var cfg *config.GlobalConfig
	if existing, loadErr := config.LoadConfig(cfgPath); loadErr == nil {
		cfg = existing
	} else {
		cfg = &config.GlobalConfig{Repos: []string{}}
	}

	cfg.Token = token
	cfg.AddRepo(fullName)
	cfg.CurrentRepo = fullName
	if err := cfg.Save(cfgPath); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	result, repoDir, err := runSyncForRepoFunc(cmd, fullName)
	if err != nil {
		return fmt.Errorf("configuration was saved to %s, but automatic sync failed for %s: %w. Run 'igmeek sync' to retry", cfgPath, fullName, err)
	}

	fmt.Fprintf(out, "Synced %d issues, %d labels from %s\n", result.IssuesCount, result.LabelsCount, fullName)
	fmt.Fprintf(out, "Config saved at: %s\n", cfgPath)
	fmt.Fprintf(out, "Repo data stored at: %s\n", repoDir)
	return nil
}
```

- [ ] **Step 5: 运行 `init` 测试，确认转绿**

Run: `go test ./cmd -run TestRunInit -v`

Expected: PASS for both the success case and the preserved-config-on-sync-failure case.

- [ ] **Step 6: 运行更大范围回归，确认 `sync` 仍可工作**

Run: `go test ./cmd ./internal/config ./internal/sync -v`

Expected: PASS with no regressions in existing sync/config tests.

- [ ] **Step 7: 提交这一小步**

```bash
git add cmd/init.go cmd/init_test.go cmd/sync.go
git commit -m "feat: run initial sync during init"
```

---

### Task 3: 增强索引文件匹配并把新错误体验接入 `update`

**Files:**
- Modify: `internal/index/index.go`
- Modify: `internal/index/index_test.go`
- Modify: `cmd/update.go`

- [ ] **Step 1: 先写索引匹配测试**

在 `internal/index/index_test.go` 追加：

```go
func TestIssueIndexFindByFilePathUsesBasenameFallback(t *testing.T) {
	tmpDir := t.TempDir()
	idx := NewIssueIndex(tmpDir)
	if err := idx.Save([]IssueEntry{{IssueNumber: 7, FilePath: "posts/post.md", Title: "Post"}}); err != nil {
		t.Fatalf("failed to save index: %v", err)
	}

	match, ok := idx.FindByFilePath("/tmp/work/post.md")
	if !ok {
		t.Fatal("expected basename fallback match")
	}
	if match.IssueNumber != 7 {
		t.Fatalf("expected issue 7, got %d", match.IssueNumber)
	}
}

func TestIssueIndexFindByFilePathHandlesWindowsSeparators(t *testing.T) {
	tmpDir := t.TempDir()
	idx := NewIssueIndex(tmpDir)
	if err := idx.Save([]IssueEntry{{IssueNumber: 8, FilePath: `posts\\note.md`, Title: "Note"}}); err != nil {
		t.Fatalf("failed to save index: %v", err)
	}

	match, ok := idx.FindByFilePath(`C:\\work\\note.md`)
	if !ok {
		t.Fatal("expected windows basename fallback match")
	}
	if match.IssueNumber != 8 {
		t.Fatalf("expected issue 8, got %d", match.IssueNumber)
	}
}

func TestIssueIndexFindByFilePathRejectsAmbiguousBasename(t *testing.T) {
	tmpDir := t.TempDir()
	idx := NewIssueIndex(tmpDir)
	if err := idx.Save([]IssueEntry{
		{IssueNumber: 12, FilePath: "posts/foo.md", Title: "Foo"},
		{IssueNumber: 37, FilePath: "drafts/foo.md", Title: "Foo Draft"},
	}); err != nil {
		t.Fatalf("failed to save index: %v", err)
	}

	_, ok := idx.FindByFilePath("/tmp/foo.md")
	if ok {
		t.Fatal("expected ambiguous basename to fail")
	}
}

func TestIssueIndexFindFileMatchesReturnsSuggestions(t *testing.T) {
	tmpDir := t.TempDir()
	idx := NewIssueIndex(tmpDir)
	if err := idx.Save([]IssueEntry{
		{IssueNumber: 1, FilePath: "posts/foo.md", Title: "Foo"},
		{IssueNumber: 2, FilePath: "posts/foo-test.md", Title: "Foo Test"},
	}); err != nil {
		t.Fatalf("failed to save index: %v", err)
	}

	result := idx.FindFileMatches("/tmp/fao.md")
	if result.Found {
		t.Fatal("expected no exact match")
	}
	if len(result.Suggestions) == 0 {
		t.Fatal("expected suggestions for similar filename")
	}
}
```

- [ ] **Step 2: 运行索引测试，确认新增测试先失败**

Run: `go test ./internal/index -v`

Expected: FAIL because `FindByFilePath` still only does exact string comparison and `FindFileMatches` does not exist.

- [ ] **Step 3: 在索引层实现规范化匹配结果与建议**

把 `internal/index/index.go` 调整为下面的核心结构和方法：

```go
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

type FileMatchResult struct {
	Found       bool
	Entry       IssueEntry
	Ambiguous   bool
	Candidates  []IssueEntry
	Suggestions []string
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

func (i *IssueIndex) FindByFilePath(input string) (IssueEntry, bool) {
	result := i.FindFileMatches(input)
	return result.Entry, result.Found
}

func (i *IssueIndex) FindFileMatches(input string) FileMatchResult {
	issues, err := i.Load()
	if err != nil {
		return FileMatchResult{}
	}

	normalizedInput := normalizeStoredPath(input)
	for _, issue := range issues {
		if normalizeStoredPath(issue.FilePath) == normalizedInput {
			return FileMatchResult{Found: true, Entry: issue}
		}
	}

	inputBase := basename(input)
	var candidates []IssueEntry
	for _, issue := range issues {
		if basename(issue.FilePath) == inputBase {
			candidates = append(candidates, issue)
		}
	}

	if len(candidates) == 1 {
		return FileMatchResult{Found: true, Entry: candidates[0]}
	}

	if len(candidates) > 1 {
		return FileMatchResult{
			Ambiguous:   true,
			Candidates:  candidates,
			Suggestions: candidateDescriptions(candidates),
		}
	}

	return FileMatchResult{Suggestions: similarFileNames(inputBase, issues)}
}

func normalizeStoredPath(value string) string {
	replaced := strings.ReplaceAll(strings.TrimSpace(value), `\\`, "/")
	replaced = strings.ReplaceAll(replaced, `\`, "/")
	return path.Clean(replaced)
}

func basename(value string) string {
	normalized := normalizeStoredPath(value)
	return path.Base(normalized)
}

func candidateDescriptions(entries []IssueEntry) []string {
	result := make([]string, 0, len(entries))
	for _, entry := range entries {
		result = append(result, fmt.Sprintf("%s -> #%d", entry.FilePath, entry.IssueNumber))
	}
	return result
}

func similarFileNames(target string, issues []IssueEntry) []string {
	target = strings.ToLower(target)
	type suggestion struct {
		name  string
		score int
	}
	var scored []suggestion
	for _, issue := range issues {
		name := basename(issue.FilePath)
		score := similarityScore(target, strings.ToLower(name))
		if score > 0 {
			scored = append(scored, suggestion{name: name, score: score})
		}
	}
	sort.Slice(scored, func(i, j int) bool {
		if scored[i].score == scored[j].score {
			return scored[i].name < scored[j].name
		}
		return scored[i].score > scored[j].score
	})
	seen := map[string]bool{}
	var result []string
	for _, item := range scored {
		if seen[item.name] {
			continue
		}
		seen[item.name] = true
		result = append(result, item.name)
		if len(result) == 3 {
			break
		}
	}
	return result
}

func similarityScore(a, b string) int {
	if a == "" || b == "" {
		return 0
	}
	if a == b {
		return len(a) + 10
	}
	if strings.Contains(b, a) || strings.Contains(a, b) {
		return min(len(a), len(b)) + 5
	}
	common := 0
	for _, part := range strings.FieldsFunc(a, func(r rune) bool { return r == '-' || r == '_' || r == '.' }) {
		if part != "" && strings.Contains(b, part) {
			common += len(part)
		}
	}
	return common
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
```

- [ ] **Step 4: 在 `update` 命令里接入新的匹配结果与错误提示**

把 `cmd/update.go` 中单参数分支改成：

```go
	} else {
		absPath, err := markdown.NormalizePath(args[0])
		if err != nil {
			return err
		}

		match := issueIndex.FindFileMatches(absPath)
		if !match.Found {
			if match.Ambiguous {
				msg := "存在多个同名文件，请使用 igmeek update <num> <file>"
				if len(match.Suggestions) > 0 {
					msg += "\n候选：" + strings.Join(match.Suggestions, ", ")
				}
				return fmt.Errorf(msg)
			}

			msg := "未找到对应文件名的 issue 映射，先执行 sync 或显式传入 issue_number"
			if len(match.Suggestions) > 0 {
				msg += "\n相近文件：" + strings.Join(match.Suggestions, ", ")
			}
			return fmt.Errorf(msg)
		}

		issueNum = match.Entry.IssueNumber
		filePath = absPath
	}
```

- [ ] **Step 5: 运行索引与更新相关测试，确认转绿**

Run: `go test ./internal/index ./cmd -v`

Expected: PASS for basename fallback, Windows separator handling, ambiguous match rejection, and suggestion generation.

- [ ] **Step 6: 运行完整回归**

Run: `go test ./...`

Expected: PASS across the full repository with no failing packages.

- [ ] **Step 7: 提交这一小步**

```bash
git add internal/index/index.go internal/index/index_test.go cmd/update.go
git commit -m "fix: match issue files by basename fallback"
```

---

## Self-Review

### Spec Coverage

- `init` 同时采集 token + 初始 repo：Task 1 + Task 2 覆盖
- repo 支持 `owner/repo` 和 GitHub URL：Task 1 覆盖
- `init` 自动执行一次 sync：Task 2 覆盖
- `init` 输出结果和数据存储路径：Task 2 覆盖
- sync 失败但保留配置：Task 2 覆盖
- file 匹配截取最后文件名：Task 3 覆盖
- 绝对路径/相对路径兼容：Task 3 覆盖
- Win/Linux/macOS 路径规范差异：Task 3 覆盖
- 找不到时给出建议：Task 3 覆盖
- 同名文件歧义时提示显式 issue number：Task 3 覆盖

### Placeholder Scan

- 没有 `TODO`、`TBD`、`implement later`
- 每个任务都包含明确文件、代码和命令
- 没有“参考前一个任务”这类跳跃描述

### Type Consistency

- repo 规范化函数统一为 `NormalizeRepoInput`
- repo 追加函数统一为 `AddRepo`
- 文件匹配结构统一为 `FileMatchResult`
- `sync` 共用 helper 统一为 `runSyncForRepo`
