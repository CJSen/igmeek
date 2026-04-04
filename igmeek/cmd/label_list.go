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
