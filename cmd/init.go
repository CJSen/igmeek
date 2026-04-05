package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/CJSen/igmeek/cli/internal/config"
	"github.com/CJSen/igmeek/cli/internal/sync"
	"github.com/spf13/cobra"
)

var runSyncForRepoFunc = func(cmd *cobra.Command, fullName string, token string) (*sync.SyncResult, string, error) {
	return runSyncForRepo(fullName, token)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize igmeek with your GitHub token",
	Long:  "Interactively prompt for a GitHub Personal Access Token and save it to the global configuration file. The token requires the 'repo' scope. Can also be set via the IMGEEK_GITHUB_TOKEN environment variable.",
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

	token, err := promptLine(reader, out, "Enter your GitHub Personal Access Token (needs 'repo' scope): ")
	if err != nil {
		return err
	}

	if token == "" {
		return fmt.Errorf("token cannot be empty")
	}

	repoInput, err := promptLine(reader, out, "Enter the repository (owner/repo or GitHub URL): ")
	if err != nil {
		return err
	}

	fullName, err := config.NormalizeRepoInput(repoInput)
	if err != nil {
		return err
	}

	cfgPath := config.ConfigPath(globalDir)
	var cfg *config.GlobalConfig
	existing, err := config.LoadConfig(cfgPath)
	if err == nil {
		cfg = existing
	} else if errors.Is(err, os.ErrNotExist) {
		cfg = &config.GlobalConfig{
			Repos: []string{},
		}
	} else {
		return fmt.Errorf("failed to load existing config: %w", err)
	}

	cfg.Token = token
	cfg.AddRepo(fullName)
	cfg.CurrentRepo = fullName
	if err := cfg.Save(cfgPath); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	repoDir := config.GetRepoDir(globalDir, fullName)
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		return fmt.Errorf("failed to create repo data directory: %w", err)
	}

	result, repoDir, err := runSyncForRepoFunc(cmd, fullName, token)
	if err != nil {
		return fmt.Errorf("sync failed after init, but configuration was saved. You can retry with 'igmeek sync': %w", err)
	}

	fmt.Fprintf(out, "Synced %d issues, %d labels from %s\n", result.IssuesCount, result.LabelsCount, fullName)
	fmt.Fprintf(out, "Config saved at: %s\n", cfgPath)
	fmt.Fprintf(out, "Repo data stored at: %s\n", repoDir)
	return nil
}

func promptLine(reader *bufio.Reader, out io.Writer, prompt string) (string, error) {
	if _, err := fmt.Fprint(out, prompt); err != nil {
		return "", fmt.Errorf("failed to write prompt: %w", err)
	}

	value, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	return strings.TrimSpace(value), nil
}
