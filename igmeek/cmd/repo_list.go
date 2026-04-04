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
