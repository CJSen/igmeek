package cmd

import (
	"context"
	"fmt"

	"github.com/CJSen/igmeek/cli/internal/api"
	"github.com/CJSen/igmeek/cli/internal/config"
	"github.com/CJSen/igmeek/cli/internal/sync"
	"github.com/spf13/cobra"
)

var globalDataDirFunc = config.GetGlobalDataDir

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync all remote issues and labels to local cache",
	Long:  "Fetch all open and closed issues from the configured repository and update the local index (issues_num_name.json) and tag cache (tags.json). This is a full sync that overwrites local index data.",
	RunE:  runSync,
}

func init() {
	rootCmd.AddCommand(syncCmd)
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

	result, _, err := runSyncForRepoFunc(cmd, cfg.CurrentRepo, "")
	if err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Synced %d issues, %d labels from %s\n", result.IssuesCount, result.LabelsCount, cfg.CurrentRepo)
	return nil
}

func runSyncForRepo(fullName string, token string) (*sync.SyncResult, string, error) {
	globalDir := globalDataDirFunc()

	owner, repo, err := sync.ParseOwnerRepo(fullName)
	if err != nil {
		return nil, "", err
	}

	repoDir := config.GetRepoDir(globalDir, fullName)
	if token == "" {
		token = GetToken()
	}
	client := api.NewClient(token)

	result, err := sync.SyncAll(context.Background(), client, owner, repo, repoDir)
	if err != nil {
		return nil, "", err
	}

	return result, repoDir, nil
}
