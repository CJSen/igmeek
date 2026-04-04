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
