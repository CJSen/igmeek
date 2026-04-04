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
	Long:  "Fetch all open and closed issues from the configured repository and update the local index (issues_num_name.json) and tag cache (tags.json). This is a full sync that overwrites local index data.",
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
