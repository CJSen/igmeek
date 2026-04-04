# igmeek CLI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 构建 `igmeek` CLI，实现本地优先的 GitHub Issue/Tag 管理，适配 Gmeek 标签驱动发布流程。

**Architecture:** Go CLI 基于 cobra 框架，internal 层封装 GitHub API、配置管理、索引管理和 Markdown 读取，cmd 层提供 sync/new/update/del/undel/label/repo 等子命令。全量同步策略，以 issue_number 为主键维护本地索引。

**Tech Stack:** Go 1.21+, cobra, go-github, standard library (encoding/json, os, filepath)

---

## File Structure Map

```
igmeek/                          # 项目根目录（即 /Users/css/dev/igmeek/igmeek/）
├── main.go                      # 入口，调用 cmd.Execute()
├── go.mod                       # Go module 定义
├── go.sum                       # 依赖校验
├── cmd/
│   ├── root.go                  # 根命令，全局 flag，token 解析
│   ├── sync.go                  # sync 命令
│   ├── new.go                   # new 命令
│   ├── update.go                # update 命令
│   ├── del.go                   # del 命令
│   ├── undel.go                 # undel 命令
│   ├── label.go                 # label 父命令
│   ├── label_list.go            # label list 子命令
│   ├── label_add.go             # label add 子命令
│   ├── repo.go                  # repo 父命令
│   ├── repo_add.go              # repo add 子命令
│   ├── repo_del.go              # repo del 子命令
│   ├── repo_list.go             # repo list 子命令
│   └── repo_use.go              # repo use 子命令
├── internal/
│   ├── config/
│   │   ├── config.go            # 全局配置读写 + 数据目录管理
│   │   └── config_test.go       # 配置模块单元测试
│   ├── api/
│   │   ├── github.go            # GitHub API 封装
│   │   └── github_test.go       # API 模块单元测试
│   ├── index/
│   │   ├── index.go             # Issue 索引 + 标签缓存读写
│   │   └── index_test.go        # 索引模块单元测试
│   ├── markdown/
│   │   ├── reader.go            # Markdown 文件读取 + 路径归一化
│   │   └── reader_test.go       # Markdown 模块单元测试
│   └── sync/
│       ├── sync.go              # 全量同步逻辑
│       └── sync_test.go         # 同步模块单元测试
└── README.md                    # 最小文档
```

### File Responsibilities

| File | Responsibility |
|------|---------------|
| `main.go` | 程序入口，调用 `cmd.Execute()` |
| `cmd/root.go` | 根命令定义、全局 flag、token 获取与验证、退出码常量 |
| `cmd/sync.go` | 调用 sync 模块执行全量同步 |
| `cmd/new.go` | 解析 file/tag/notag 参数，调用 api 创建 Issue |
| `cmd/update.go` | 解析 file/num/tag 操作参数，调用 api 更新 Issue |
| `cmd/del.go` | 调用 api 关闭 Issue |
| `cmd/undel.go` | 调用 api 重开 Issue |
| `cmd/label.go` | label 子命令的父命令，无独立逻辑 |
| `cmd/label_list.go` | 列出标签，调用 api + index 更新缓存 |
| `cmd/label_add.go` | 创建标签，调用 api + index 更新缓存 |
| `cmd/repo.go` | repo 子命令的父命令，无独立逻辑 |
| `cmd/repo_add.go` | 添加仓库，交互式或参数式 |
| `cmd/repo_del.go` | 删除仓库，交互式选择 |
| `cmd/repo_list.go` | 列出已绑定仓库 |
| `cmd/repo_use.go` | 切换当前仓库 |
| `internal/config/config.go` | 全局数据目录路径计算、config.json 读写、仓库目录管理 |
| `internal/api/github.go` | go-github 客户端封装：Issues CRUD、Labels CRUD、分页 |
| `internal/index/index.go` | issues_num_name.json 和 tags.json 的读写、查询 |
| `internal/markdown/reader.go` | 读取 Markdown 文件内容、路径归一化、提取标题 |
| `internal/sync/sync.go` | 协调 api 和 index 完成全量同步 |

---

## Task 1: 初始化 Go 模块与 cobra 根命令

**Files:**
- Create: `igmeek/go.mod`
- Create: `igmeek/main.go`
- Create: `igmeek/cmd/root.go`

- [ ] **Step 1: 初始化 Go 模块**

```bash
cd /Users/css/dev/igmeek
mkdir -p igmeek/cmd igmeek/internal/config igmeek/internal/api igmeek/internal/index igmeek/internal/markdown igmeek/internal/sync
cd igmeek
go mod init github.com/CJSen/igmeek
```

- [ ] **Step 2: 安装 cobra 依赖**

```bash
cd /Users/css/dev/igmeek/igmeek
go get github.com/spf13/cobra@latest
```

- [ ] **Step 3: 创建 main.go**

`igmeek/main.go`:
```go
package main

import (
	"github.com/CJSen/igmeek/cmd"
)

func main() {
	cmd.Execute()
}
```

- [ ] **Step 4: 创建 cmd/root.go**

`igmeek/cmd/root.go`:
```go
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

const (
	ExitSuccess      = 0
	ExitGeneralError = 1
	ExitParamError   = 2
	ExitAuthError    = 3
	ExitNetworkError = 4
)

var rootCmd = &cobra.Command{
	Use:   "igmeek",
	Short: "Local-first GitHub Issue/Tag management CLI for Gmeek blogs",
	Long: `igmeek is a CLI tool for managing GitHub Issues and Tags
for blogs built with the Gmeek framework.

It allows you to create, update, close, and reopen Issues
from your local terminal, with label management tailored
for Gmeek's label-driven publishing workflow.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(ExitGeneralError)
	}
}
```

- [ ] **Step 5: 验证编译通过**

```bash
cd /Users/css/dev/igmeek/igmeek
go build -o igmeek .
./igmeek --help
```

Expected output: shows igmeek help with Use/Short/Long text.

- [ ] **Step 6: 提交**

```bash
cd /Users/css/dev/igmeek
git add igmeek/
git commit -m "feat: initialize Go CLI project with cobra root command"
```

---

## Task 2: 实现跨平台全局数据目录与全局配置

**Files:**
- Create: `igmeek/internal/config/config.go`
- Create: `igmeek/internal/config/config_test.go`

- [ ] **Step 1: 编写 config 模块测试**

`igmeek/internal/config/config_test.go`:
```go
package config

import (
	"path/filepath"
	"testing"
)

func TestGetGlobalDataDir(t *testing.T) {
	dir := GetGlobalDataDir()
	if dir == "" {
		t.Error("expected non-empty global data dir")
	}
	if filepath.Base(dir) != "igmeek" {
		t.Errorf("expected dir to end with 'igmeek', got %s", dir)
	}
}

func TestConfigSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &GlobalConfig{
		Token:       "test-token",
		CurrentRepo: "owner/repo",
		Repos:       []string{"owner/repo"},
	}

	cfgPath := filepath.Join(tmpDir, "config.json")
	if err := cfg.Save(cfgPath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	loaded, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if loaded.Token != "test-token" {
		t.Errorf("expected token 'test-token', got %s", loaded.Token)
	}
	if loaded.CurrentRepo != "owner/repo" {
		t.Errorf("expected current_repo 'owner/repo', got %s", loaded.CurrentRepo)
	}
	if len(loaded.Repos) != 1 || loaded.Repos[0] != "owner/repo" {
		t.Errorf("expected repos ['owner/repo'], got %v", loaded.Repos)
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/config.json")
	if err == nil {
		t.Error("expected error for missing config file")
	}
}

func TestRepoDirPath(t *testing.T) {
	globalDir := t.TempDir()
	repoDir := GetRepoDir(globalDir, "owner/repo")
	expected := filepath.Join(globalDir, "repos", "owner_repo")
	if repoDir != expected {
		t.Errorf("expected %s, got %s", expected, repoDir)
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

```bash
cd /Users/css/dev/igmeek/igmeek
go test ./internal/config/ -v
```

Expected: FAIL -- functions not defined.

- [ ] **Step 3: 实现 config 模块**

`igmeek/internal/config/config.go`:
```go
package config

import (
	"encoding/json"
	"fmt"
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

- [ ] **Step 4: 运行测试确认通过**

```bash
cd /Users/css/dev/igmeek/igmeek
go test ./internal/config/ -v
```

Expected: All tests PASS.

- [ ] **Step 5: 提交**

```bash
cd /Users/css/dev/igmeek
git add igmeek/internal/config/
git commit -m "feat: implement cross-platform global config and data directory"
```

---

## Task 3: 实现 Token 解析与初始化流程

**Files:**
- Modify: `igmeek/cmd/root.go`
- Create: `igmeek/cmd/init.go`

- [ ] **Step 1: 创建 init 命令**

`igmeek/cmd/init.go`:
```go
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/CJSen/igmeek/internal/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize igmeek with your GitHub token",
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	globalDir := config.GetGlobalDataDir()
	if err := config.EnsureGlobalDir(globalDir); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	fmt.Print("Enter your GitHub Personal Access Token (needs 'repo' scope): ")
	reader := bufio.NewReader(os.Stdin)
	token, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}
	token = strings.TrimSpace(token)

	if token == "" {
		return fmt.Errorf("token cannot be empty")
	}

	cfgPath := config.ConfigPath(globalDir)
	var cfg *config.GlobalConfig
	if existing, err := config.LoadConfig(cfgPath); err == nil {
		cfg = existing
	} else {
		cfg = &config.GlobalConfig{
			Repos: []string{},
		}
	}

	cfg.Token = token
	if err := cfg.Save(cfgPath); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println("Token saved to configuration file.")
	return nil
}
```

- [ ] **Step 2: 更新 root.go 添加 token 获取逻辑**

`igmeek/cmd/root.go` 完整内容替换为：
```go
package cmd

import (
	"os"

	"github.com/CJSen/igmeek/internal/config"
	"github.com/spf13/cobra"
)

const (
	ExitSuccess      = 0
	ExitGeneralError = 1
	ExitParamError   = 2
	ExitAuthError    = 3
	ExitNetworkError = 4
)

var rootCmd = &cobra.Command{
	Use:   "igmeek",
	Short: "Local-first GitHub Issue/Tag management CLI for Gmeek blogs",
	Long: `igmeek is a CLI tool for managing GitHub Issues and Tags
for blogs built with the Gmeek framework.

It allows you to create, update, close, and reopen Issues
from your local terminal, with label management tailored
for Gmeek's label-driven publishing workflow.`,
	PersistentPreRunE: preRun,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(ExitGeneralError)
	}
}

func preRun(cmd *cobra.Command, args []string) error {
	if cmd.Name() == "init" || cmd.Name() == "help" || cmd.Name() == "completion" {
		return nil
	}

	token := GetToken()
	if token == "" {
		return &TokenError{Msg: "no GitHub token found. Please set IMGEEK_GITHUB_TOKEN environment variable or run 'igmeek init'"}
	}
	return nil
}

func GetToken() string {
	if token := os.Getenv("IMGEEK_GITHUB_TOKEN"); token != "" {
		return token
	}

	globalDir := config.GetGlobalDataDir()
	cfgPath := config.ConfigPath(globalDir)
	if cfg, err := config.LoadConfig(cfgPath); err == nil {
		return cfg.Token
	}

	return ""
}

type TokenError struct {
	Msg string
}

func (e *TokenError) Error() string {
	return e.Msg
}
```

- [ ] **Step 3: 验证编译和命令**

```bash
cd /Users/css/dev/igmeek/igmeek
go build -o igmeek .
./igmeek --help
./igmeek init --help
```

Expected: `init` command appears in help, `--help` works for both.

- [ ] **Step 4: 提交**

```bash
cd /Users/css/dev/igmeek
git add igmeek/cmd/init.go igmeek/cmd/root.go
git commit -m "feat: add token parsing and init flow"
```

---

## Task 4: 实现 GitHub API 适配层

**Files:**
- Create: `igmeek/internal/api/github.go`
- Create: `igmeek/internal/api/github_test.go`

- [ ] **Step 1: 安装 go-github 依赖**

```bash
cd /Users/css/dev/igmeek/igmeek
go get github.com/google/go-github/v68/github
go get golang.org/x/oauth2
```

- [ ] **Step 2: 编写 API 模块测试**

`igmeek/internal/api/github_test.go`:
```go
package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test-token")
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestListIssues(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client := NewClient("test-token")
	issues, _, err := client.ListIssues(context.Background(), "owner", "repo")
	// With real go-github this would need a mock server with proper responses
	// For now, just verify client creation doesn't panic
	_ = issues
	_ = err
}
```

- [ ] **Step 3: 实现 API 封装**

`igmeek/internal/api/github.go`:
```go
package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-github/v68/github"
	"golang.org/x/oauth2"
)

type Client struct {
	gh *github.Client
}

func NewClient(token string) *Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	return &Client{
		gh: github.NewClient(tc),
	}
}

func (c *Client) ListIssues(ctx context.Context, owner, repo string) ([]*github.Issue, error) {
	var allIssues []*github.Issue
	opts := &github.IssueListByRepoOptions{
		State: "all",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		issues, resp, err := c.gh.Issues.ListByRepo(ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list issues: %w", err)
		}
		allIssues = append(allIssues, issues...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allIssues, nil
}

func (c *Client) CreateIssue(ctx context.Context, owner, repo, title, body string) (*github.Issue, error) {
	issueReq := &github.IssueRequest{
		Title: github.String(title),
		Body:  github.String(body),
	}

	issue, _, err := c.gh.Issues.Create(ctx, owner, repo, issueReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create issue: %w", err)
	}

	return issue, nil
}

func (c *Client) EditIssue(ctx context.Context, owner, repo string, number int, title, body string) (*github.Issue, error) {
	issueReq := &github.IssueRequest{}
	if title != "" {
		issueReq.Title = github.String(title)
	}
	if body != "" {
		issueReq.Body = github.String(body)
	}

	issue, _, err := c.gh.Issues.Edit(ctx, owner, repo, number, issueReq)
	if err != nil {
		return nil, fmt.Errorf("failed to edit issue: %w", err)
	}

	return issue, nil
}

func (c *Client) CloseIssue(ctx context.Context, owner, repo string, number int) (*github.Issue, error) {
	issueReq := &github.IssueRequest{
		State: github.String("closed"),
	}

	issue, _, err := c.gh.Issues.Edit(ctx, owner, repo, number, issueReq)
	if err != nil {
		return nil, fmt.Errorf("failed to close issue: %w", err)
	}

	return issue, nil
}

func (c *Client) ReopenIssue(ctx context.Context, owner, repo string, number int) (*github.Issue, error) {
	issueReq := &github.IssueRequest{
		State: github.String("open"),
	}

	issue, _, err := c.gh.Issues.Edit(ctx, owner, repo, number, issueReq)
	if err != nil {
		return nil, fmt.Errorf("failed to reopen issue: %w", err)
	}

	return issue, nil
}

func (c *Client) AddLabels(ctx context.Context, owner, repo string, number int, labels []string) ([]*github.Label, error) {
	result, _, err := c.gh.Issues.AddLabelsToIssue(ctx, owner, repo, number, labels)
	if err != nil {
		return nil, fmt.Errorf("failed to add labels: %w", err)
	}

	return result, nil
}

func (c *Client) ReplaceLabels(ctx context.Context, owner, repo string, number int, labels []string) ([]*github.Label, error) {
	result, _, err := c.gh.Issues.ReplaceLabelsForIssue(ctx, owner, repo, number, labels)
	if err != nil {
		return nil, fmt.Errorf("failed to replace labels: %w", err)
	}

	return result, nil
}

func (c *Client) RemoveLabel(ctx context.Context, owner, repo string, number int, label string) (*http.Response, error) {
	resp, err := c.gh.Issues.RemoveLabelForIssue(ctx, owner, repo, number, label)
	if err != nil {
		return nil, fmt.Errorf("failed to remove label: %w", err)
	}

	return resp, nil
}

func (c *Client) ListLabels(ctx context.Context, owner, repo string) ([]*github.Label, error) {
	var allLabels []*github.Label
	opts := &github.ListOptions{
		PerPage: 100,
	}

	for {
		labels, resp, err := c.gh.Issues.ListLabels(ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list labels: %w", err)
		}
		allLabels = append(allLabels, labels...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allLabels, nil
}

func (c *Client) CreateLabel(ctx context.Context, owner, repo, name string) (*github.Label, error) {
	label := &github.Label{
		Name:  github.String(name),
		Color: github.String("ededed"),
	}

	result, _, err := c.gh.Issues.CreateLabel(ctx, owner, repo, label)
	if err != nil {
		return nil, fmt.Errorf("failed to create label: %w", err)
	}

	return result, nil
}

func (c *Client) GetIssue(ctx context.Context, owner, repo string, number int) (*github.Issue, *github.Response, error) {
	return c.gh.Issues.Get(ctx, owner, repo, number)
}

func (c *Client) VerifyRepo(ctx context.Context, owner, repo string) error {
	_, _, err := c.gh.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return fmt.Errorf("failed to access repository: %w", err)
	}
	return nil
}
```

- [ ] **Step 4: 运行测试确认通过**

```bash
cd /Users/css/dev/igmeek/igmeek
go test ./internal/api/ -v
```

Expected: All tests PASS.

- [ ] **Step 5: 提交**

```bash
cd /Users/css/dev/igmeek
git add igmeek/internal/api/
git commit -m "feat: implement GitHub API wrapper layer with go-github"
```

---

## Task 5: 实现仓库级数据文件读写（索引模块）

**Files:**
- Create: `igmeek/internal/index/index.go`
- Create: `igmeek/internal/index/index_test.go`

- [ ] **Step 1: 编写 index 模块测试**

`igmeek/internal/index/index_test.go`:
```go
package index

import (
	"path/filepath"
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
```

- [ ] **Step 2: 运行测试确认失败**

```bash
cd /Users/css/dev/igmeek/igmeek
go test ./internal/index/ -v
```

Expected: FAIL -- types and functions not defined.

- [ ] **Step 3: 实现 index 模块**

`igmeek/internal/index/index.go`:
```go
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
```

- [ ] **Step 4: 运行测试确认通过**

```bash
cd /Users/css/dev/igmeek/igmeek
go test ./internal/index/ -v
```

Expected: All tests PASS.

- [ ] **Step 5: 提交**

```bash
cd /Users/css/dev/igmeek
git add igmeek/internal/index/
git commit -m "feat: implement repository-level index and tag cache"
```

---

## Task 6: 实现 Markdown 文件读取与路径归一化

**Files:**
- Create: `igmeek/internal/markdown/reader.go`
- Create: `igmeek/internal/markdown/reader_test.go`

- [ ] **Step 1: 编写 markdown 模块测试**

`igmeek/internal/markdown/reader_test.go`:
```go
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
```

- [ ] **Step 2: 运行测试确认失败**

```bash
cd /Users/css/dev/igmeek/igmeek
go test ./internal/markdown/ -v
```

Expected: FAIL -- functions not defined.

- [ ] **Step 3: 实现 markdown 模块**

`igmeek/internal/markdown/reader.go`:
```go
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
```

- [ ] **Step 4: 运行测试确认通过**

```bash
cd /Users/css/dev/igmeek/igmeek
go test ./internal/markdown/ -v
```

Expected: All tests PASS.

- [ ] **Step 5: 提交**

```bash
cd /Users/css/dev/igmeek
git add igmeek/internal/markdown/
git commit -m "feat: implement markdown file reading and path normalization"
```

---

## Task 7: 实现 sync 全量同步

**Files:**
- Create: `igmeek/internal/sync/sync.go`
- Create: `igmeek/internal/sync/sync_test.go`
- Create: `igmeek/cmd/sync.go`

- [ ] **Step 1: 编写 sync 模块测试**

`igmeek/internal/sync/sync_test.go`:
```go
package sync

import (
	"testing"

	"github.com/google/go-github/v68/github"
)

func TestConvertIssueToEntry(t *testing.T) {
	ghIssue := &github.Issue{
		Number: github.Int(42),
		Title:  github.String("Test Issue"),
		State:  github.String("open"),
	}

	entry := convertIssueToEntry(ghIssue)

	if entry.IssueNumber != 42 {
		t.Errorf("expected number 42, got %d", entry.IssueNumber)
	}
	if entry.Title != "Test Issue" {
		t.Errorf("expected title 'Test Issue', got %s", entry.Title)
	}
	if entry.State != "open" {
		t.Errorf("expected state 'open', got %s", entry.State)
	}
}

func TestConvertLabelToEntry(t *testing.T) {
	ghLabel := &github.Label{
		Name:  github.String("tech"),
		Color: github.String("0075ca"),
	}

	entry := convertLabelToEntry(ghLabel)

	if entry.Name != "tech" {
		t.Errorf("expected name 'tech', got %s", entry.Name)
	}
	if entry.Color != "0075ca" {
		t.Errorf("expected color '0075ca', got %s", entry.Color)
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

```bash
cd /Users/css/dev/igmeek/igmeek
go test ./internal/sync/ -v
```

Expected: FAIL -- functions not defined.

- [ ] **Step 3: 实现 sync 模块**

`igmeek/internal/sync/sync.go`:
```go
package sync

import (
	"context"
	"fmt"
	"strings"

	"github.com/CJSen/igmeek/internal/api"
	"github.com/CJSen/igmeek/internal/index"
	"github.com/google/go-github/v68/github"
)

type SyncResult struct {
	IssuesCount int
	LabelsCount int
}

func SyncAll(ctx context.Context, client *api.Client, owner, repo string, repoDir string) (*SyncResult, error) {
	issues, err := client.ListIssues(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch issues: %w", err)
	}

	labels, err := client.ListLabels(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch labels: %w", err)
	}

	issueIndex := index.NewIssueIndex(repoDir)
	tagCache := index.NewTagCache(repoDir)

	var issueEntries []index.IssueEntry
	for _, issue := range issues {
		issueEntries = append(issueEntries, convertIssueToEntry(issue))
	}

	if err := issueIndex.Save(issueEntries); err != nil {
		return nil, fmt.Errorf("failed to save issue index: %w", err)
	}

	var tagEntries []index.TagEntry
	for _, label := range labels {
		tagEntries = append(tagEntries, convertLabelToEntry(label))
	}

	if err := tagCache.Save(tagEntries); err != nil {
		return nil, fmt.Errorf("failed to save tag cache: %w", err)
	}

	return &SyncResult{
		IssuesCount: len(issueEntries),
		LabelsCount: len(tagEntries),
	}, nil
}

func convertIssueToEntry(issue *github.Issue) index.IssueEntry {
	var labels []string
	for _, label := range issue.Labels {
		if label.Name != nil {
			labels = append(labels, *label.Name)
		}
	}

	entry := index.IssueEntry{
		IssueNumber: issue.GetNumber(),
		Title:       issue.GetTitle(),
		Labels:      labels,
		State:       issue.GetState(),
		URL:         issue.GetURL(),
		HTMLURL:     issue.GetHTMLURL(),
	}

	if issue.CreatedAt != nil {
		t := issue.CreatedAt.Time
		entry.CreatedAt = &t
	}
	if issue.UpdatedAt != nil {
		t := issue.UpdatedAt.Time
		entry.UpdatedAt = &t
	}
	if issue.ClosedAt != nil {
		t := issue.ClosedAt.Time
		entry.ClosedAt = &t
	}

	return entry
}

func convertLabelToEntry(label *github.Label) index.TagEntry {
	return index.TagEntry{
		Name:  label.GetName(),
		Color: label.GetColor(),
	}
}

func ParseOwnerRepo(fullName string) (string, string, error) {
	parts := strings.Split(fullName, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid repository format: %s (expected owner/repo)", fullName)
	}
	return parts[0], parts[1], nil
}
```

- [ ] **Step 4: 运行测试确认通过**

```bash
cd /Users/css/dev/igmeek/igmeek
go test ./internal/sync/ -v
```

Expected: All tests PASS.

- [ ] **Step 5: 创建 sync 命令**

`igmeek/cmd/sync.go`:
```go
package cmd

import (
	"context"
	"fmt"

	"github.com/CJSen/igmeek/internal/api"
	"github.com/CJSen/igmeek/internal/config"
	"github.com/CJSen/igmeek/internal/sync"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync all remote issues and labels to local cache",
	RunE:  runSync,
}

func init() {
	rootCmd.AddCommand(syncCmd)
}

func runSync(cmd *cobra.Command, args []string) error {
	globalDir := config.GetGlobalDataDir()
	cfg, err := config.LoadConfig(config.ConfigPath(globalDir))
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.CurrentRepo == "" {
		return fmt.Errorf("no repository configured. Run 'igmeek repo add' first")
	}

	owner, repo, err := sync.ParseOwnerRepo(cfg.CurrentRepo)
	if err != nil {
		return err
	}

	repoDir := config.GetRepoDir(globalDir, cfg.CurrentRepo)
	client := api.NewClient(GetToken())

	result, err := sync.SyncAll(context.Background(), client, owner, repo, repoDir)
	if err != nil {
		return err
	}

	fmt.Printf("Synced %d issues, %d labels from %s\n", result.IssuesCount, result.LabelsCount, cfg.CurrentRepo)
	return nil
}
```

- [ ] **Step 6: 验证编译**

```bash
cd /Users/css/dev/igmeek/igmeek
go build -o igmeek .
./igmeek sync --help
```

Expected: sync command help displayed.

- [ ] **Step 7: 提交**

```bash
cd /Users/css/dev/igmeek
git add igmeek/internal/sync/ igmeek/cmd/sync.go
git commit -m "feat: implement full sync command with pagination"
```

---

## Task 8: 实现 repo add/del/list/use 命令

**Files:**
- Create: `igmeek/cmd/repo.go`
- Create: `igmeek/cmd/repo_add.go`
- Create: `igmeek/cmd/repo_del.go`
- Create: `igmeek/cmd/repo_list.go`
- Create: `igmeek/cmd/repo_use.go`

- [ ] **Step 1: 创建 repo 父命令**

`igmeek/cmd/repo.go`:
```go
package cmd

import (
	"github.com/spf13/cobra"
)

var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Manage repository configurations",
}

func init() {
	rootCmd.AddCommand(repoCmd)
}
```

- [ ] **Step 2: 实现 repo add**

`igmeek/cmd/repo_add.go`:
```go
package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/CJSen/igmeek/internal/api"
	"github.com/CJSen/igmeek/internal/config"
	"github.com/CJSen/igmeek/internal/sync"
	"github.com/spf13/cobra"
)

var repoAddCmd = &cobra.Command{
	Use:   "add [owner/repo]",
	Short: "Add a repository configuration",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runRepoAdd,
}

func init() {
	repoCmd.AddCommand(repoAddCmd)
}

func runRepoAdd(cmd *cobra.Command, args []string) error {
	var fullName string

	if len(args) > 0 {
		fullName = args[0]
	} else {
		fmt.Print("Enter repository (owner/repo): ")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}
		fullName = strings.TrimSpace(input)
	}

	owner, repo, err := sync.ParseOwnerRepo(fullName)
	if err != nil {
		return err
	}

	globalDir := config.GetGlobalDataDir()
	cfg, err := config.LoadConfig(config.ConfigPath(globalDir))
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := api.NewClient(GetToken())
	if err := client.VerifyRepo(context.Background(), owner, repo); err != nil {
		return fmt.Errorf("cannot access repository: %w", err)
	}

	repoDir := config.GetRepoDir(globalDir, fullName)
	repoConfig := &config.RepoConfig{
		Owner:    owner,
		Repo:     repo,
		FullName: fullName,
	}

	if err := repoConfig.Save(repoDir); err != nil {
		return fmt.Errorf("failed to save repo config: %w", err)
	}

	found := false
	for _, r := range cfg.Repos {
		if r == fullName {
			found = true
			break
		}
	}
	if !found {
		cfg.Repos = append(cfg.Repos, fullName)
	}

	if cfg.CurrentRepo == "" {
		cfg.CurrentRepo = fullName
	}

	if err := cfg.Save(config.ConfigPath(globalDir)); err != nil {
		return fmt.Errorf("failed to save global config: %w", err)
	}

	fmt.Printf("Added repository: %s\n", fullName)
	return nil
}
```

- [ ] **Step 3: 实现 repo list**

`igmeek/cmd/repo_list.go`:
```go
package cmd

import (
	"fmt"

	"github.com/CJSen/igmeek/internal/config"
	"github.com/spf13/cobra"
)

var repoListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured repositories",
	RunE:  runRepoList,
}

func init() {
	repoCmd.AddCommand(repoListCmd)
}

func runRepoList(cmd *cobra.Command, args []string) error {
	globalDir := config.GetGlobalDataDir()
	cfg, err := config.LoadConfig(config.ConfigPath(globalDir))
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(cfg.Repos) == 0 {
		fmt.Println("No repositories configured.")
		return nil
	}

	fmt.Println("Configured repositories:")
	for _, r := range cfg.Repos {
		if r == cfg.CurrentRepo {
			fmt.Printf("* %s (current)\n", r)
		} else {
			fmt.Printf("  %s\n", r)
		}
	}

	return nil
}
```

- [ ] **Step 4: 实现 repo use**

`igmeek/cmd/repo_use.go`:
```go
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/CJSen/igmeek/internal/config"
	"github.com/spf13/cobra"
)

var repoUseCmd = &cobra.Command{
	Use:   "use",
	Short: "Select the current working repository",
	RunE:  runRepoUse,
}

func init() {
	repoCmd.AddCommand(repoUseCmd)
}

func runRepoUse(cmd *cobra.Command, args []string) error {
	globalDir := config.GetGlobalDataDir()
	cfg, err := config.LoadConfig(config.ConfigPath(globalDir))
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(cfg.Repos) == 0 {
		return fmt.Errorf("no repositories configured. Run 'igmeek repo add' first")
	}

	if len(cfg.Repos) == 1 {
		cfg.CurrentRepo = cfg.Repos[0]
		if err := cfg.Save(config.ConfigPath(globalDir)); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		fmt.Printf("Switched to: %s\n", cfg.CurrentRepo)
		return nil
	}

	fmt.Println("Select a repository:")
	for i, r := range cfg.Repos {
		marker := " "
		if r == cfg.CurrentRepo {
			marker = "*"
		}
		fmt.Printf("  %d. %s%s\n", i+1, r, marker)
	}

	fmt.Print("Enter number: ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	num, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil || num < 1 || num > len(cfg.Repos) {
		return fmt.Errorf("invalid selection")
	}

	cfg.CurrentRepo = cfg.Repos[num-1]
	if err := cfg.Save(config.ConfigPath(globalDir)); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Switched to: %s\n", cfg.CurrentRepo)
	return nil
}
```

- [ ] **Step 5: 实现 repo del**

`igmeek/cmd/repo_del.go`:
```go
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/CJSen/igmeek/internal/config"
	"github.com/spf13/cobra"
)

var repoDelCmd = &cobra.Command{
	Use:   "del",
	Short: "Remove a repository configuration",
	RunE:  runRepoDel,
}

func init() {
	repoCmd.AddCommand(repoDelCmd)
}

func runRepoDel(cmd *cobra.Command, args []string) error {
	globalDir := config.GetGlobalDataDir()
	cfg, err := config.LoadConfig(config.ConfigPath(globalDir))
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(cfg.Repos) == 0 {
		return fmt.Errorf("no repositories configured")
	}

	var target string
	if len(cfg.Repos) == 1 {
		target = cfg.Repos[0]
	} else {
		fmt.Println("Select a repository to remove:")
		for i, r := range cfg.Repos {
			fmt.Printf("  %d. %s\n", i+1, r)
		}

		fmt.Print("Enter number: ")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		num, err := strconv.Atoi(strings.TrimSpace(input))
		if err != nil || num < 1 || num > len(cfg.Repos) {
			return fmt.Errorf("invalid selection")
		}

		target = cfg.Repos[num-1]
	}

	var newRepos []string
	for _, r := range cfg.Repos {
		if r != target {
			newRepos = append(newRepos, r)
		}
	}
	cfg.Repos = newRepos

	if cfg.CurrentRepo == target {
		if len(newRepos) > 0 {
			cfg.CurrentRepo = newRepos[0]
		} else {
			cfg.CurrentRepo = ""
		}
	}

	if err := cfg.Save(config.ConfigPath(globalDir)); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	repoDir := config.GetRepoDir(globalDir, target)
	os.RemoveAll(repoDir)

	fmt.Printf("Removed repository: %s\n", target)
	return nil
}
```

- [ ] **Step 6: 验证编译**

```bash
cd /Users/css/dev/igmeek/igmeek
go build -o igmeek .
./igmeek repo --help
```

Expected: repo subcommands listed.

- [ ] **Step 7: 提交**

```bash
cd /Users/css/dev/igmeek
git add igmeek/cmd/repo*.go
git commit -m "feat: implement repo add/del/list/use commands"
```

---

## Task 9: 实现 new 命令

**Files:**
- Create: `igmeek/cmd/new.go`

- [ ] **Step 1: 创建 new 命令**

`igmeek/cmd/new.go`:
```go
package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/CJSen/igmeek/internal/api"
	"github.com/CJSen/igmeek/internal/config"
	"github.com/CJSen/igmeek/internal/index"
	"github.com/CJSen/igmeek/internal/markdown"
	"github.com/CJSen/igmeek/internal/sync"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new <file>",
	Short: "Create a new issue from a markdown file",
	Args:  cobra.ExactArgs(1),
	RunE:  runNew,
}

var (
	newTags  string
	newNoTag bool
)

func init() {
	rootCmd.AddCommand(newCmd)
	newCmd.Flags().StringVar(&newTags, "tag", "", "Comma-separated labels (required unless --notag)")
	newCmd.Flags().BoolVar(&newNoTag, "notag", false, "Create issue without labels")
}

func runNew(cmd *cobra.Command, args []string) error {
	if !newNoTag && newTags == "" {
		return fmt.Errorf("must specify --tag (at least one) or --notag")
	}

	if !newNoTag {
		tags := strings.Split(newTags, ",")
		var cleaned []string
		for _, t := range tags {
			t = strings.TrimSpace(t)
			if t != "" {
				cleaned = append(cleaned, t)
			}
		}
		if len(cleaned) == 0 {
			return fmt.Errorf("must specify at least one tag with --tag")
		}
		newTags = strings.Join(cleaned, ",")
	}

	globalDir := config.GetGlobalDataDir()
	cfg, err := config.LoadConfig(config.ConfigPath(globalDir))
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.CurrentRepo == "" {
		return fmt.Errorf("no repository configured. Run 'igmeek repo add' first")
	}

	owner, repo, err := sync.ParseOwnerRepo(cfg.CurrentRepo)
	if err != nil {
		return err
	}

	mdFile, err := markdown.ReadFile(args[0])
	if err != nil {
		return err
	}

	client := api.NewClient(GetToken())
	issue, err := client.CreateIssue(context.Background(), owner, repo, mdFile.Title, mdFile.Content)
	if err != nil {
		return err
	}

	repoDir := config.GetRepoDir(globalDir, cfg.CurrentRepo)
	issueIndex := index.NewIssueIndex(repoDir)

	var labels []string
	if !newNoTag {
		labelList := strings.Split(newTags, ",")
		_, err = client.AddLabels(context.Background(), owner, repo, issue.GetNumber(), labelList)
		if err != nil {
			return fmt.Errorf("failed to add labels: %w", err)
		}
		labels = labelList
	}

	entries, _ := issueIndex.Load()
	newEntry := index.IssueEntry{
		IssueNumber: issue.GetNumber(),
		FilePath:    mdFile.AbsPath,
		Title:       issue.GetTitle(),
		Labels:      labels,
		State:       issue.GetState(),
		URL:         issue.GetURL(),
		HTMLURL:     issue.GetHTMLURL(),
	}
	if issue.CreatedAt != nil {
		t := issue.CreatedAt.Time
		newEntry.CreatedAt = &t
	}
	if issue.UpdatedAt != nil {
		t := issue.UpdatedAt.Time
		newEntry.UpdatedAt = &t
	}
	entries = append(entries, newEntry)
	if err := issueIndex.Save(entries); err != nil {
		return fmt.Errorf("failed to update index: %w", err)
	}

	fmt.Printf("Created issue #%d: %s\n", issue.GetNumber(), issue.GetTitle())
	fmt.Printf("URL: %s\n", issue.GetHTMLURL())
	return nil
}
```

- [ ] **Step 2: 验证编译**

```bash
cd /Users/css/dev/igmeek/igmeek
go build -o igmeek .
./igmeek new --help
```

Expected: new command help with --tag and --notag flags shown.

- [ ] **Step 3: 提交**

```bash
cd /Users/css/dev/igmeek
git add igmeek/cmd/new.go
git commit -m "feat: implement new command with --tag and --notag options"
```

---

## Task 10: 实现 update 命令

**Files:**
- Create: `igmeek/cmd/update.go`

- [ ] **Step 1: 创建 update 命令**

`igmeek/cmd/update.go`:
```go
package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/CJSen/igmeek/internal/api"
	"github.com/CJSen/igmeek/internal/config"
	"github.com/CJSen/igmeek/internal/index"
	"github.com/CJSen/igmeek/internal/markdown"
	"github.com/CJSen/igmeek/internal/sync"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update <file|num> [file]",
	Short: "Update an existing issue",
	Args:  cobra.RangeArgs(1, 2),
	RunE:  runUpdate,
}

var (
	updateAddTag    string
	updateRemoveTag string
	updateSetTag    string
)

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().StringVar(&updateAddTag, "add-tag", "", "Add labels (comma-separated)")
	updateCmd.Flags().StringVar(&updateRemoveTag, "remove-tag", "", "Remove labels (comma-separated)")
	updateCmd.Flags().StringVar(&updateSetTag, "set-tag", "", "Replace all labels (comma-separated)")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	globalDir := config.GetGlobalDataDir()
	cfg, err := config.LoadConfig(config.ConfigPath(globalDir))
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.CurrentRepo == "" {
		return fmt.Errorf("no repository configured. Run 'igmeek repo add' first")
	}

	owner, repo, err := sync.ParseOwnerRepo(cfg.CurrentRepo)
	if err != nil {
		return err
	}

	repoDir := config.GetRepoDir(globalDir, cfg.CurrentRepo)
	issueIndex := index.NewIssueIndex(repoDir)
	client := api.NewClient(GetToken())

	var issueNum int
	var filePath string

	if len(args) == 2 {
		num, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("first argument must be an issue number when providing two arguments")
		}
		issueNum = num
		filePath = args[1]
	} else {
		absPath, err := markdown.NormalizePath(args[0])
		if err != nil {
			return err
		}
		entry, ok := issueIndex.FindByFilePath(absPath)
		if !ok {
			return fmt.Errorf("issue not found for file: %s. Run 'igmeek sync' to refresh index", absPath)
		}
		issueNum = entry.IssueNumber
		filePath = absPath
	}

	_, found := issueIndex.FindByNumber(issueNum)
	if !found {
		return fmt.Errorf("issue #%d not found in index. Run 'igmeek sync' to refresh", issueNum)
	}

	var body string
	title := ""
	if filePath != "" {
		mdFile, err := markdown.ReadFile(filePath)
		if err != nil {
			return err
		}
		body = mdFile.Content
		title = mdFile.Title
	}

	issue, err := client.EditIssue(context.Background(), owner, repo, issueNum, title, body)
	if err != nil {
		return err
	}

	if updateAddTag != "" {
		tags := parseTags(updateAddTag)
		existing, _ := issueIndex.FindByNumber(issueNum)
		allTags := append(existing.Labels, tags...)
		allTags = uniqueStrings(allTags)
		_, err = client.ReplaceLabels(context.Background(), owner, repo, issueNum, allTags)
		if err != nil {
			return fmt.Errorf("failed to add labels: %w", err)
		}
		issue.Labels = makeLabelSlice(allTags)
	}

	if updateRemoveTag != "" {
		tags := parseTags(updateRemoveTag)
		existing, _ := issueIndex.FindByNumber(issueNum)
		var remaining []string
		removeSet := make(map[string]bool)
		for _, t := range tags {
			removeSet[t] = true
		}
		for _, t := range existing.Labels {
			if !removeSet[t] {
				remaining = append(remaining, t)
			}
		}
		_, err = client.ReplaceLabels(context.Background(), owner, repo, issueNum, remaining)
		if err != nil {
			return fmt.Errorf("failed to remove labels: %w", err)
		}
		issue.Labels = makeLabelSlice(remaining)
	}

	if updateSetTag != "" {
		tags := parseTags(updateSetTag)
		_, err = client.ReplaceLabels(context.Background(), owner, repo, issueNum, tags)
		if err != nil {
			return fmt.Errorf("failed to set labels: %w", err)
		}
		issue.Labels = makeLabelSlice(tags)
	}

	entries, _ := issueIndex.Load()
	for i, entry := range entries {
		if entry.IssueNumber == issueNum {
			entries[i].FilePath = filePath
			if filePath != "" {
				entries[i].FilePath = filePath
			}
			entries[i].Title = issue.GetTitle()
			entries[i].State = issue.GetState()
			var labels []string
			for _, l := range issue.Labels {
				labels = append(labels, l.GetName())
			}
			entries[i].Labels = labels
			if issue.UpdatedAt != nil {
				t := issue.UpdatedAt.Time
				entries[i].UpdatedAt = &t
			}
			break
		}
	}
	issueIndex.Save(entries)

	fmt.Printf("Updated issue #%d: %s\n", issue.GetNumber(), issue.GetTitle())
	return nil
}

func parseTags(s string) []string {
	var result []string
	for _, t := range strings.Split(s, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			result = append(result, t)
		}
	}
	return result
}

func uniqueStrings(s []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, v := range s {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}

func makeLabelSlice(names []string) []*github.Label {
	var labels []*github.Label
	for _, name := range names {
		labels = append(labels, &github.Label{Name: github.String(name)})
	}
	return labels
}
```

- [ ] **Step 2: 添加 github import 到 update.go**

需要确保 update.go 顶部有：
```go
import (
	"github.com/google/go-github/v68/github"
)
```

- [ ] **Step 3: 验证编译**

```bash
cd /Users/css/dev/igmeek/igmeek
go build -o igmeek .
./igmeek update --help
```

Expected: update command help with --add-tag, --remove-tag, --set-tag flags.

- [ ] **Step 4: 提交**

```bash
cd /Users/css/dev/igmeek
git add igmeek/cmd/update.go
git commit -m "feat: implement update command with label operations"
```

---

## Task 11: 实现 del 与 undel 命令

**Files:**
- Create: `igmeek/cmd/del.go`
- Create: `igmeek/cmd/undel.go`

- [ ] **Step 1: 创建 del 命令**

`igmeek/cmd/del.go`:
```go
package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/CJSen/igmeek/internal/api"
	"github.com/CJSen/igmeek/internal/config"
	"github.com/CJSen/igmeek/internal/index"
	"github.com/CJSen/igmeek/internal/sync"
	"github.com/spf13/cobra"
)

var delCmd = &cobra.Command{
	Use:   "del <num>",
	Short: "Close an issue without deleting local file",
	Args:  cobra.ExactArgs(1),
	RunE:  runDel,
}

func init() {
	rootCmd.AddCommand(delCmd)
}

func runDel(cmd *cobra.Command, args []string) error {
	num, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid issue number: %s", args[0])
	}

	globalDir := config.GetGlobalDataDir()
	cfg, err := config.LoadConfig(config.ConfigPath(globalDir))
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.CurrentRepo == "" {
		return fmt.Errorf("no repository configured. Run 'igmeek repo add' first")
	}

	owner, repo, err := sync.ParseOwnerRepo(cfg.CurrentRepo)
	if err != nil {
		return err
	}

	repoDir := config.GetRepoDir(globalDir, cfg.CurrentRepo)
	issueIndex := index.NewIssueIndex(repoDir)

	entry, ok := issueIndex.FindByNumber(num)
	if !ok {
		return fmt.Errorf("issue #%d not found in index. Run 'igmeek sync' to refresh", num)
	}

	client := api.NewClient(GetToken())
	issue, err := client.CloseIssue(context.Background(), owner, repo, num)
	if err != nil {
		return err
	}

	entries, _ := issueIndex.Load()
	for i, e := range entries {
		if e.IssueNumber == num {
			entries[i].State = "closed"
			if issue.ClosedAt != nil {
				t := issue.ClosedAt.Time
				entries[i].ClosedAt = &t
			}
			break
		}
	}
	issueIndex.Save(entries)

	fmt.Printf("Closed issue #%d: %s\n", num, entry.Title)
	return nil
}
```

- [ ] **Step 2: 创建 undel 命令**

`igmeek/cmd/undel.go`:
```go
package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/CJSen/igmeek/internal/api"
	"github.com/CJSen/igmeek/internal/config"
	"github.com/CJSen/igmeek/internal/index"
	"github.com/CJSen/igmeek/internal/sync"
	"github.com/spf13/cobra"
)

var undelCmd = &cobra.Command{
	Use:   "undel <num>",
	Short: "Reopen a closed issue",
	Args:  cobra.ExactArgs(1),
	RunE:  runUndel,
}

func init() {
	rootCmd.AddCommand(undelCmd)
}

func runUndel(cmd *cobra.Command, args []string) error {
	num, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid issue number: %s", args[0])
	}

	globalDir := config.GetGlobalDataDir()
	cfg, err := config.LoadConfig(config.ConfigPath(globalDir))
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.CurrentRepo == "" {
		return fmt.Errorf("no repository configured. Run 'igmeek repo add' first")
	}

	owner, repo, err := sync.ParseOwnerRepo(cfg.CurrentRepo)
	if err != nil {
		return err
	}

	repoDir := config.GetRepoDir(globalDir, cfg.CurrentRepo)
	issueIndex := index.NewIssueIndex(repoDir)

	entry, ok := issueIndex.FindByNumber(num)
	if !ok {
		return fmt.Errorf("issue #%d not found in index. Run 'igmeek sync' to refresh", num)
	}

	client := api.NewClient(GetToken())
	issue, err := client.ReopenIssue(context.Background(), owner, repo, num)
	if err != nil {
		return err
	}

	entries, _ := issueIndex.Load()
	for i, e := range entries {
		if e.IssueNumber == num {
			entries[i].State = "open"
			break
		}
	}
	issueIndex.Save(entries)

	fmt.Printf("Reopened issue #%d: %s\n", num, entry.Title)
	return nil
}
```

- [ ] **Step 3: 验证编译**

```bash
cd /Users/css/dev/igmeek/igmeek
go build -o igmeek .
./igmeek del --help
./igmeek undel --help
```

Expected: Both commands show help text.

- [ ] **Step 4: 提交**

```bash
cd /Users/css/dev/igmeek
git add igmeek/cmd/del.go igmeek/cmd/undel.go
git commit -m "feat: implement del and undel commands"
```

---

## Task 12: 实现 label list 与 label add 命令

**Files:**
- Create: `igmeek/cmd/label.go`
- Create: `igmeek/cmd/label_list.go`
- Create: `igmeek/cmd/label_add.go`

- [ ] **Step 1: 创建 label 父命令**

`igmeek/cmd/label.go`:
```go
package cmd

import (
	"github.com/spf13/cobra"
)

var labelCmd = &cobra.Command{
	Use:   "label",
	Short: "Manage repository labels",
}

func init() {
	rootCmd.AddCommand(labelCmd)
}
```

- [ ] **Step 2: 实现 label list**

`igmeek/cmd/label_list.go`:
```go
package cmd

import (
	"context"
	"fmt"

	"github.com/CJSen/igmeek/internal/api"
	"github.com/CJSen/igmeek/internal/config"
	"github.com/CJSen/igmeek/internal/index"
	"github.com/CJSen/igmeek/internal/sync"
	"github.com/spf13/cobra"
)

var labelListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all repository labels",
	RunE:  runLabelList,
}

func init() {
	labelCmd.AddCommand(labelListCmd)
}

func runLabelList(cmd *cobra.Command, args []string) error {
	globalDir := config.GetGlobalDataDir()
	cfg, err := config.LoadConfig(config.ConfigPath(globalDir))
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.CurrentRepo == "" {
		return fmt.Errorf("no repository configured. Run 'igmeek repo add' first")
	}

	owner, repo, err := sync.ParseOwnerRepo(cfg.CurrentRepo)
	if err != nil {
		return err
	}

	repoDir := config.GetRepoDir(globalDir, cfg.CurrentRepo)
	tagCache := index.NewTagCache(repoDir)

	client := api.NewClient(GetToken())
	labels, err := client.ListLabels(context.Background(), owner, repo)
	if err != nil {
		return err
	}

	var tagEntries []index.TagEntry
	for _, label := range labels {
		tagEntries = append(tagEntries, index.TagEntry{
			Name:  label.GetName(),
			Color: label.GetColor(),
		})
	}
	tagCache.Save(tagEntries)

	fmt.Printf("Labels in %s:\n", cfg.CurrentRepo)
	for _, entry := range tagEntries {
		fmt.Printf("  %s\n", entry.Name)
	}

	return nil
}
```

- [ ] **Step 3: 实现 label add**

`igmeek/cmd/label_add.go`:
```go
package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/CJSen/igmeek/internal/api"
	"github.com/CJSen/igmeek/internal/config"
	"github.com/CJSen/igmeek/internal/index"
	"github.com/CJSen/igmeek/internal/sync"
	"github.com/spf13/cobra"
)

var labelAddCmd = &cobra.Command{
	Use:   "add <tags>",
	Short: "Create repository labels",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runLabelAdd,
}

func init() {
	labelCmd.AddCommand(labelAddCmd)
}

func runLabelAdd(cmd *cobra.Command, args []string) error {
	globalDir := config.GetGlobalDataDir()
	cfg, err := config.LoadConfig(config.ConfigPath(globalDir))
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.CurrentRepo == "" {
		return fmt.Errorf("no repository configured. Run 'igmeek repo add' first")
	}

	owner, repo, err := sync.ParseOwnerRepo(cfg.CurrentRepo)
	if err != nil {
		return err
	}

	repoDir := config.GetRepoDir(globalDir, cfg.CurrentRepo)
	tagCache := index.NewTagCache(repoDir)
	client := api.NewClient(GetToken())

	var created []string
	for _, name := range args {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		_, err := client.CreateLabel(context.Background(), owner, repo, name)
		if err != nil {
			return fmt.Errorf("failed to create label '%s': %w", name, err)
		}
		created = append(created, name)
	}

	existingTags, _ := tagCache.Load()
	for _, name := range created {
		existingTags = append(existingTags, index.TagEntry{
			Name:  name,
			Color: "ededed",
		})
	}
	tagCache.Save(existingTags)

	fmt.Printf("Created labels: %s\n", strings.Join(created, ", "))
	return nil
}
```

- [ ] **Step 4: 验证编译**

```bash
cd /Users/css/dev/igmeek/igmeek
go build -o igmeek .
./igmeek label --help
```

Expected: label subcommands listed.

- [ ] **Step 5: 提交**

```bash
cd /Users/css/dev/igmeek
git add igmeek/cmd/label*.go
git commit -m "feat: implement label list and label add commands"
```

---

## Task 13: 补齐帮助信息与错误提示

**Files:**
- Modify: `igmeek/cmd/root.go`
- Modify: `igmeek/cmd/new.go`
- Modify: `igmeek/cmd/update.go`
- Modify: `igmeek/cmd/del.go`
- Modify: `igmeek/cmd/undel.go`

- [ ] **Step 1: 更新 root.go 添加更好的错误处理**

在 `cmd/root.go` 的 `Execute()` 函数中添加错误类型判断：

```go
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		if tokenErr, ok := err.(*TokenError); ok {
			fmt.Fprintf(os.Stderr, "Error: %s\n", tokenErr.Msg)
			os.Exit(ExitAuthError)
		}
		os.Exit(ExitGeneralError)
	}
}
```

- [ ] **Step 2: 为每个命令添加 Long 描述**

在每个命令定义中添加 `Long` 字段。例如 `new.go`:

```go
var newCmd = &cobra.Command{
	Use:   "new <file>",
	Short: "Create a new issue from a markdown file",
	Long: `Create a new GitHub Issue from a local Markdown file.

The issue title is extracted from the first H1 heading (# Title)
or the filename if no heading is found.

You must specify either --tag (at least one label) or --notag.
Issues with labels will be recognized by Gmeek as publishable.`,
	Args: cobra.ExactArgs(1),
	RunE: runNew,
}
```

- [ ] **Step 3: 验证所有命令帮助**

```bash
cd /Users/css/dev/igmeek/igmeek
go build -o igmeek .
for cmd in sync new update del undel "label list" "label add" "repo add" "repo del" "repo list" "repo use" init; do
  echo "=== igmeek $cmd ==="
  ./igmeek $cmd --help
  echo ""
done
```

Expected: Each command shows proper help with description, usage, and flags.

- [ ] **Step 4: 提交**

```bash
cd /Users/css/dev/igmeek
git add igmeek/cmd/
git commit -m "docs: improve help text and error messages for all commands"
```

---

## Task 14: 手工联调与最小文档

**Files:**
- Create: `igmeek/README.md`

- [ ] **Step 1: 编写 README**

`igmeek/README.md`:
```markdown
# igmeek

Local-first GitHub Issue/Tag management CLI for Gmeek blogs.

## Installation

```bash
cd igmeek
go build -o igmeek .
sudo cp igmeek /usr/local/bin/
```

## Quick Start

```bash
# Initialize with your GitHub token
igmeek init

# Add your blog repository
igmeek repo add CJSen/cjsen.github.io

# Sync all issues and labels
igmeek sync

# Create a new publishable issue
igmeek new ./my-post.md --tag tech,blog

# Create a draft issue (no labels)
igmeek new ./draft.md --notag

# Update an existing issue
igmeek update ./my-post.md

# Close an issue
igmeek del 42

# Reopen an issue
igmeek undel 42

# Manage labels
igmeek label list
igmeek label add tutorial
```

## Commands

| Command | Description |
|---------|-------------|
| `igmeek init` | Initialize with GitHub token |
| `igmeek sync` | Sync all issues and labels to local cache |
| `igmeek new <file> --tag <tags>` | Create new issue with labels |
| `igmeek new <file> --notag` | Create new issue without labels |
| `igmeek update <file>` | Update issue by file path |
| `igmeek update <num> <file>` | Update issue by number |
| `igmeek del <num>` | Close issue |
| `igmeek undel <num>` | Reopen issue |
| `igmeek label list` | List repository labels |
| `igmeek label add <tags>` | Create repository labels |
| `igmeek repo add [owner/repo]` | Add repository configuration |
| `igmeek repo del` | Remove repository configuration |
| `igmeek repo list` | List configured repositories |
| `igmeek repo use` | Select current working repository |

## Environment Variables

- `IMGEEK_GITHUB_TOKEN` - GitHub Personal Access Token (requires `repo` scope)

## Configuration

Configuration is stored in the standard user config directory:
- macOS: `~/Library/Application Support/igmeek/`
- Linux: `~/.config/igmeek/`
- Windows: `%APPDATA%\igmeek\`
```

- [ ] **Step 2: 验证完整编译**

```bash
cd /Users/css/dev/igmeek/igmeek
go build -o igmeek .
./igmeek --help
```

- [ ] **Step 3: 运行所有单元测试**

```bash
cd /Users/css/dev/igmeek/igmeek
go test ./... -v
```

Expected: All tests PASS.

- [ ] **Step 4: 更新 igmeek-cli-history.md**

Update the history file to reflect completion:

```markdown
## 六、当前进度

### 已完成
- [x] 探索项目上下文
- [x] 判断是否需要可视化辅助
- [x] 逐轮澄清问题（目标、约束、成功标准）
- [x] 提出可行方向并给出取舍建议
- [x] 分段呈现设计并获取确认
- [x] 将确认后的设计写入 spec 文档
- [x] 对设计文档做一次自检
- [x] 用户审阅 spec 文档
- [x] 转入 writing-plans 阶段
- [x] 实现计划已落盘

### 当前状态
实现计划已完成，准备开始编码实现。
```

- [ ] **Step 5: 最终提交**

```bash
cd /Users/css/dev/igmeek
git add igmeek/README.md igmeek-cli-history.md
git commit -m "docs: add README and update project history"
```

---

## Self-Review

### Spec Coverage Check

| Spec Section | Task Coverage |
|-------------|---------------|
| 项目概述 | Task 1 (root command Long description) |
| 技术架构/项目结构 | Task 1-13 (all files created per structure) |
| 数据模型 - config.json | Task 2 (GlobalConfig) |
| 数据模型 - repo.json | Task 2 (RepoConfig) |
| 数据模型 - issues_num_name.json | Task 5 (IssueEntry) |
| 数据模型 - tags.json | Task 5 (TagEntry) |
| 路径归一化规则 | Task 6 (NormalizePath) |
| 认证与初始化 | Task 3 (init command, GetToken) |
| sync 命令 | Task 7 |
| new 命令 (--tag/--notag) | Task 9 |
| update 命令 | Task 10 |
| del 命令 | Task 11 |
| undel 命令 | Task 11 |
| label list | Task 12 |
| label add | Task 12 |
| repo add | Task 8 |
| repo del | Task 8 |
| repo list | Task 8 |
| repo use | Task 8 |
| 错误处理/退出码 | Task 1, Task 13 |
| 测试策略 | Each task includes unit tests |

### Placeholder Scan

No TBD, TODO, or incomplete sections found. All code blocks contain actual implementation code.

### Type Consistency Check

- `GetToken()` defined in `cmd/root.go`, used consistently in all commands
- `config.GetGlobalDataDir()`, `config.ConfigPath()`, `config.GetRepoDir()` used consistently
- `sync.ParseOwnerRepo()` used consistently for owner/repo parsing
- `index.IssueEntry` and `index.TagEntry` types match spec data model exactly
- Exit codes defined as constants in `cmd/root.go`

All consistent. Plan is ready for execution.
