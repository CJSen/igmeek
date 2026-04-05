package cmd

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/CJSen/igmeek/cli/internal/config"
	"github.com/CJSen/igmeek/cli/internal/sync"
	"github.com/spf13/cobra"
)

func TestRunInitSavesConfigAndRunsSync(t *testing.T) {
	tmpDir := t.TempDir()

	prevGlobalDataDirFunc := globalDataDirFunc
	prevRunSyncForRepoFunc := runSyncForRepoFunc
	t.Cleanup(func() {
		globalDataDirFunc = prevGlobalDataDirFunc
		runSyncForRepoFunc = prevRunSyncForRepoFunc
	})

	globalDataDirFunc = func() string {
		return tmpDir
	}

	repoDir := config.GetRepoDir(tmpDir, "octo/blog")
	runSyncForRepoFunc = func(cmd *cobra.Command, fullName string, token string) (*sync.SyncResult, string, error) {
		if fullName != "octo/blog" {
			t.Fatalf("expected repo octo/blog, got %s", fullName)
		}
		if token != "token-123" {
			t.Fatalf("expected token token-123, got %s", token)
		}
		return &sync.SyncResult{IssuesCount: 3, LabelsCount: 2}, repoDir, nil
	}

	command := &cobra.Command{}
	command.SetIn(strings.NewReader("token-123\nhttps://github.com/octo/blog\n"))
	var output bytes.Buffer
	command.SetOut(&output)

	if err := runInit(command, nil); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	cfg, err := config.LoadConfig(config.ConfigPath(tmpDir))
	if err != nil {
		t.Fatalf("expected config to be saved, got %v", err)
	}

	if cfg.Token != "token-123" {
		t.Fatalf("expected token token-123, got %s", cfg.Token)
	}
	if cfg.CurrentRepo != "octo/blog" {
		t.Fatalf("expected current repo octo/blog, got %s", cfg.CurrentRepo)
	}
	if len(cfg.Repos) != 1 || cfg.Repos[0] != "octo/blog" {
		t.Fatalf("expected repos [octo/blog], got %v", cfg.Repos)
	}

	outputText := output.String()
	if !strings.Contains(outputText, "Synced 3 issues, 2 labels from octo/blog") {
		t.Fatalf("expected sync output, got %q", outputText)
	}
	if !strings.Contains(outputText, "Config saved at: "+config.ConfigPath(tmpDir)) {
		t.Fatalf("expected config path output, got %q", outputText)
	}
	if !strings.Contains(outputText, "Repo data stored at: "+repoDir) {
		t.Fatalf("expected repo dir output, got %q", outputText)
	}
}

func TestRunInitReturnsErrorWhenSyncFailsAfterSavingConfig(t *testing.T) {
	tmpDir := t.TempDir()
	syncErr := errors.New("sync failed")

	prevGlobalDataDirFunc := globalDataDirFunc
	prevRunSyncForRepoFunc := runSyncForRepoFunc
	t.Cleanup(func() {
		globalDataDirFunc = prevGlobalDataDirFunc
		runSyncForRepoFunc = prevRunSyncForRepoFunc
	})

	globalDataDirFunc = func() string {
		return tmpDir
	}
	runSyncForRepoFunc = func(cmd *cobra.Command, fullName string, token string) (*sync.SyncResult, string, error) {
		if fullName != "octo/blog" {
			t.Fatalf("expected repo octo/blog, got %s", fullName)
		}
		if token != "token-123" {
			t.Fatalf("expected token token-123, got %s", token)
		}
		return nil, "", syncErr
	}

	command := &cobra.Command{}
	command.SetIn(strings.NewReader("token-123\nocto/blog\n"))
	var output bytes.Buffer
	command.SetOut(&output)

	err := runInit(command, nil)
	if err == nil {
		t.Fatal("expected sync error")
	}
	if !strings.Contains(err.Error(), "configuration was saved") {
		t.Fatalf("expected error to mention saved configuration, got %v", err)
	}

	cfg, loadErr := config.LoadConfig(config.ConfigPath(tmpDir))
	if loadErr != nil {
		t.Fatalf("expected config to be saved, got %v", loadErr)
	}
	if cfg.Token != "token-123" {
		t.Fatalf("expected token token-123, got %s", cfg.Token)
	}
	if cfg.CurrentRepo != "octo/blog" {
		t.Fatalf("expected current repo octo/blog, got %s", cfg.CurrentRepo)
	}
	if len(cfg.Repos) != 1 || cfg.Repos[0] != "octo/blog" {
		t.Fatalf("expected repos [octo/blog], got %v", cfg.Repos)
	}
}

func TestRunInitReturnsErrorWhenExistingConfigCannotBeLoaded(t *testing.T) {
	tmpDir := t.TempDir()
	invalidConfig := []byte("{invalid json")
	cfgPath := config.ConfigPath(tmpDir)
	if err := os.WriteFile(cfgPath, invalidConfig, 0600); err != nil {
		t.Fatalf("failed to write invalid config: %v", err)
	}

	prevGlobalDataDirFunc := globalDataDirFunc
	prevRunSyncForRepoFunc := runSyncForRepoFunc
	t.Cleanup(func() {
		globalDataDirFunc = prevGlobalDataDirFunc
		runSyncForRepoFunc = prevRunSyncForRepoFunc
	})

	globalDataDirFunc = func() string {
		return tmpDir
	}
	runSyncForRepoFunc = func(cmd *cobra.Command, fullName string, token string) (*sync.SyncResult, string, error) {
		t.Fatal("sync should not run when existing config cannot be loaded")
		return nil, "", nil
	}

	command := &cobra.Command{}
	command.SetIn(strings.NewReader("token-123\nocto/blog\n"))
	var output bytes.Buffer
	command.SetOut(&output)

	err := runInit(command, nil)
	if err == nil {
		t.Fatal("expected config load error")
	}
	if !strings.Contains(err.Error(), "failed to load existing config") {
		t.Fatalf("expected config load error, got %v", err)
	}

	data, readErr := os.ReadFile(cfgPath)
	if readErr != nil {
		t.Fatalf("failed to read config after error: %v", readErr)
	}
	if string(data) != string(invalidConfig) {
		t.Fatalf("expected invalid config to remain unchanged, got %q", string(data))
	}
}

func TestRunSyncWritesToCommandOutput(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.GlobalConfig{Token: "token-123", CurrentRepo: "octo/blog", Repos: []string{"octo/blog"}}
	if err := cfg.Save(config.ConfigPath(tmpDir)); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	prevGlobalDataDirFunc := globalDataDirFunc
	prevRunSyncForRepoFunc := runSyncForRepoFunc
	t.Cleanup(func() {
		globalDataDirFunc = prevGlobalDataDirFunc
		runSyncForRepoFunc = prevRunSyncForRepoFunc
	})

	globalDataDirFunc = func() string {
		return tmpDir
	}
	runSyncForRepoFunc = func(cmd *cobra.Command, fullName string, token string) (*sync.SyncResult, string, error) {
		if fullName != "octo/blog" {
			t.Fatalf("expected repo octo/blog, got %s", fullName)
		}
		if token != "" {
			t.Fatalf("expected empty explicit token for sync command, got %s", token)
		}
		return &sync.SyncResult{IssuesCount: 3, LabelsCount: 2}, config.GetRepoDir(tmpDir, fullName), nil
	}

	command := &cobra.Command{}
	var output bytes.Buffer
	command.SetOut(&output)

	if err := runSync(command, nil); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !strings.Contains(output.String(), "Synced 3 issues, 2 labels from octo/blog") {
		t.Fatalf("expected sync output in cobra writer, got %q", output.String())
	}
}

func TestRunInitCreatesRepoDirBeforeSync(t *testing.T) {
	tmpDir := t.TempDir()

	prevGlobalDataDirFunc := globalDataDirFunc
	prevRunSyncForRepoFunc := runSyncForRepoFunc
	t.Cleanup(func() {
		globalDataDirFunc = prevGlobalDataDirFunc
		runSyncForRepoFunc = prevRunSyncForRepoFunc
	})

	globalDataDirFunc = func() string {
		return tmpDir
	}

	repoDir := config.GetRepoDir(tmpDir, "octo/blog")
	runSyncForRepoFunc = func(cmd *cobra.Command, fullName string, token string) (*sync.SyncResult, string, error) {
		info, err := os.Stat(repoDir)
		if err != nil {
			t.Fatalf("expected repo dir to exist before sync, got %v", err)
		}
		if !info.IsDir() {
			t.Fatalf("expected repo dir path to be a directory")
		}
		return &sync.SyncResult{}, repoDir, nil
	}

	command := &cobra.Command{}
	command.SetIn(strings.NewReader("token-123\nocto/blog\n"))
	command.SetOut(&bytes.Buffer{})

	if err := runInit(command, nil); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestRunInitSyncUsesPromptedTokenEvenWhenEnvVarSet(t *testing.T) {
	tmpDir := t.TempDir()

	prevGlobalDataDirFunc := globalDataDirFunc
	prevRunSyncForRepoFunc := runSyncForRepoFunc
	prevEnvToken, hadEnvToken := os.LookupEnv("IMGEEK_GITHUB_TOKEN")
	t.Cleanup(func() {
		globalDataDirFunc = prevGlobalDataDirFunc
		runSyncForRepoFunc = prevRunSyncForRepoFunc
		if hadEnvToken {
			if err := os.Setenv("IMGEEK_GITHUB_TOKEN", prevEnvToken); err != nil {
				t.Fatalf("failed to restore env token: %v", err)
			}
			return
		}
		if err := os.Unsetenv("IMGEEK_GITHUB_TOKEN"); err != nil {
			t.Fatalf("failed to clear env token: %v", err)
		}
	})

	if err := os.Setenv("IMGEEK_GITHUB_TOKEN", "stale-env-token"); err != nil {
		t.Fatalf("failed to set env token: %v", err)
	}

	globalDataDirFunc = func() string {
		return tmpDir
	}

	runSyncForRepoFunc = func(cmd *cobra.Command, fullName string, token string) (*sync.SyncResult, string, error) {
		if token != "fresh-token" {
			t.Fatalf("expected prompted token to be used for init sync, got %s", token)
		}
		return &sync.SyncResult{}, config.GetRepoDir(tmpDir, fullName), nil
	}

	command := &cobra.Command{}
	command.SetIn(strings.NewReader("fresh-token\nocto/blog\n"))
	command.SetOut(&bytes.Buffer{})

	if err := runInit(command, nil); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
