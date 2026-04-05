package cmd

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/CJSen/igmeek/cli/internal/config"
	"github.com/CJSen/igmeek/cli/internal/sync"
	"github.com/spf13/cobra"
)

func TestRunRepoAddNormalizesGitHubURLArgument(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.GlobalConfig{Token: "token-123", Repos: []string{}}
	if err := cfg.Save(config.ConfigPath(tmpDir)); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	prevGlobalDataDirFunc := globalDataDirFunc
	prevVerifyRepoAccessFunc := verifyRepoAccessFunc
	prevRunSyncForRepoFunc := runSyncForRepoFunc
	prevEnvToken, hadEnvToken := os.LookupEnv("IMGEEK_GITHUB_TOKEN")
	t.Cleanup(func() {
		globalDataDirFunc = prevGlobalDataDirFunc
		verifyRepoAccessFunc = prevVerifyRepoAccessFunc
		runSyncForRepoFunc = prevRunSyncForRepoFunc
		if hadEnvToken {
			if err := os.Setenv("IMGEEK_GITHUB_TOKEN", prevEnvToken); err != nil {
				t.Fatalf("failed to restore env token: %v", err)
			}
		} else {
			if err := os.Unsetenv("IMGEEK_GITHUB_TOKEN"); err != nil {
				t.Fatalf("failed to unset env token: %v", err)
			}
		}
	})
	if err := os.Unsetenv("IMGEEK_GITHUB_TOKEN"); err != nil {
		t.Fatalf("failed to unset env token: %v", err)
	}

	globalDataDirFunc = func() string { return tmpDir }
	verifyRepoAccessFunc = func(ctx context.Context, token, owner, repo string) error {
		if owner != "octo" || repo != "blog" {
			t.Fatalf("expected octo/blog, got %s/%s", owner, repo)
		}
		return nil
	}
	runSyncForRepoFunc = func(cmd *cobra.Command, fullName string, token string) (*sync.SyncResult, string, error) {
		if fullName != "octo/blog" {
			t.Fatalf("expected octo/blog sync, got %s", fullName)
		}
		return &sync.SyncResult{IssuesCount: 4, LabelsCount: 2}, config.GetRepoDir(tmpDir, fullName), nil
	}

	command := &cobra.Command{}
	var output bytes.Buffer
	command.SetOut(&output)

	if err := runRepoAdd(command, []string{"https://github.com/octo/blog"}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	loaded, err := config.LoadConfig(config.ConfigPath(tmpDir))
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if len(loaded.Repos) != 1 || loaded.Repos[0] != "octo/blog" {
		t.Fatalf("expected repos [octo/blog], got %v", loaded.Repos)
	}
	if loaded.CurrentRepo != "octo/blog" {
		t.Fatalf("expected current repo octo/blog, got %s", loaded.CurrentRepo)
	}
	if !strings.Contains(output.String(), "Added repository: octo/blog") {
		t.Fatalf("expected normalized output, got %q", output.String())
	}
	if !strings.Contains(output.String(), "Synced 4 issues, 2 labels from octo/blog") {
		t.Fatalf("expected sync output, got %q", output.String())
	}
	if _, err := config.LoadRepoConfig(config.GetRepoDir(tmpDir, "octo/blog")); err != nil {
		t.Fatalf("expected repo config to be saved, got %v", err)
	}
}

func TestRunRepoAddReadsInteractiveGitHubURL(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := &config.GlobalConfig{Token: "token-123", Repos: []string{"octo/existing"}, CurrentRepo: "octo/existing"}
	if err := cfg.Save(config.ConfigPath(tmpDir)); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	prevGlobalDataDirFunc := globalDataDirFunc
	prevVerifyRepoAccessFunc := verifyRepoAccessFunc
	prevRunSyncForRepoFunc := runSyncForRepoFunc
	prevEnvToken, hadEnvToken := os.LookupEnv("IMGEEK_GITHUB_TOKEN")
	t.Cleanup(func() {
		globalDataDirFunc = prevGlobalDataDirFunc
		verifyRepoAccessFunc = prevVerifyRepoAccessFunc
		runSyncForRepoFunc = prevRunSyncForRepoFunc
		if hadEnvToken {
			if err := os.Setenv("IMGEEK_GITHUB_TOKEN", prevEnvToken); err != nil {
				t.Fatalf("failed to restore env token: %v", err)
			}
		} else {
			if err := os.Unsetenv("IMGEEK_GITHUB_TOKEN"); err != nil {
				t.Fatalf("failed to unset env token: %v", err)
			}
		}
	})
	if err := os.Unsetenv("IMGEEK_GITHUB_TOKEN"); err != nil {
		t.Fatalf("failed to unset env token: %v", err)
	}

	globalDataDirFunc = func() string { return tmpDir }
	verifyRepoAccessFunc = func(ctx context.Context, token, owner, repo string) error {
		if owner != "octo" || repo != "notes" {
			t.Fatalf("expected octo/notes, got %s/%s", owner, repo)
		}
		return nil
	}
	runSyncForRepoFunc = func(cmd *cobra.Command, fullName string, token string) (*sync.SyncResult, string, error) {
		if fullName != "octo/notes" {
			t.Fatalf("expected octo/notes sync, got %s", fullName)
		}
		return &sync.SyncResult{IssuesCount: 1, LabelsCount: 1}, config.GetRepoDir(tmpDir, fullName), nil
	}

	command := &cobra.Command{}
	command.SetIn(strings.NewReader("https://github.com/octo/notes\n"))
	var output bytes.Buffer
	command.SetOut(&output)

	if err := runRepoAdd(command, nil); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	loaded, err := config.LoadConfig(config.ConfigPath(tmpDir))
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if len(loaded.Repos) != 2 || loaded.Repos[1] != "octo/notes" {
		t.Fatalf("expected octo/notes appended, got %v", loaded.Repos)
	}
	if loaded.CurrentRepo != "octo/existing" {
		t.Fatalf("expected current repo to stay octo/existing, got %s", loaded.CurrentRepo)
	}
	printed := output.String()
	if !strings.Contains(printed, "Enter repository (owner/repo or GitHub URL): ") {
		t.Fatalf("expected prompt in output, got %q", printed)
	}
	if !strings.Contains(printed, "Added repository: octo/notes") {
		t.Fatalf("expected normalized output, got %q", printed)
	}
	if !strings.Contains(printed, "Synced 1 issues, 1 labels from octo/notes") {
		t.Fatalf("expected sync output, got %q", printed)
	}
}
