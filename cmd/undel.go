package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/CJSen/igmeek/cli/internal/api"
	"github.com/CJSen/igmeek/cli/internal/config"
	"github.com/CJSen/igmeek/cli/internal/index"
	"github.com/CJSen/igmeek/cli/internal/sync"
	"github.com/spf13/cobra"
)

var undelCmd = &cobra.Command{
	Use:   "undel <num>",
	Short: "Reopen a closed issue",
	Long:  "Reopen a previously closed GitHub Issue by number. Updates the local index state to 'open'.",
	Args:  cobra.ExactArgs(1),
	RunE:  runUndel,
}

func init() {
	rootCmd.AddCommand(undelCmd)
}

func runUndel(cmd *cobra.Command, args []string) error {
	num, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid issue number: %s", args[0])
	}

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
	issueIndex := index.NewIssueIndex(repoDir)

	entry, ok := issueIndex.FindByNumber(num)
	if !ok {
		return fmt.Errorf("issue #%d not found in index. Run 'igmeek sync' to refresh", num)
	}

	client := api.NewClient(GetToken())
	_, err = client.ReopenIssue(context.Background(), owner, repo, num)
	if err != nil {
		return err
	}

	entries, _ := issueIndex.Load()
	for i, e := range entries {
		if e.IssueNumber == num {
			entries[i].State = "open"
			break
		}
	}
	issueIndex.Save(entries)

	fmt.Printf("Reopened issue #%d: %s\n", num, entry.Title)
	return nil
}
