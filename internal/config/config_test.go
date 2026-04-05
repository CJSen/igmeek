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

func TestNormalizeRepoInputRejectsInvalidOwnerRepoValues(t *testing.T) {
	invalidInputs := []string{
		"octo /test",
		"octo/test repo",
		"octo/test?x=1",
		"octo/test#frag",
	}

	for _, input := range invalidInputs {
		_, err := NormalizeRepoInput(input)
		if err == nil {
			t.Fatalf("expected invalid repo input error for %q", input)
		}
	}
}

func TestNormalizeRepoInputRejectsGitHubURLWithQueryOrFragment(t *testing.T) {
	invalidInputs := []string{
		"https://github.com/octo/test?x=1",
		"https://github.com/octo/test#frag",
	}

	for _, input := range invalidInputs {
		_, err := NormalizeRepoInput(input)
		if err == nil {
			t.Fatalf("expected invalid repo input error for %q", input)
		}
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
