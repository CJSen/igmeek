package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/CJSen/igmeek/cli/internal/api"
	"github.com/CJSen/igmeek/cli/internal/config"
	"github.com/CJSen/igmeek/cli/internal/index"
	"github.com/CJSen/igmeek/cli/internal/sync"
	"github.com/spf13/cobra"
)

var labelAddCmd = &cobra.Command{
	Use:   "add <tags>",
	Short: "Create repository labels",
	Long:  "Create one or more labels in the configured repository. Labels are created with the default color (ededed). Multiple labels can be specified as separate arguments.",
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
