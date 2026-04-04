package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/CJSen/igmeek/internal/api"
	"github.com/CJSen/igmeek/internal/config"
	"github.com/CJSen/igmeek/internal/index"
	"github.com/CJSen/igmeek/internal/sync"
	"github.com/spf13/cobra"
)

var delCmd = &cobra.Command{
	Use:   "del <num>",
	Short: "Close an issue without deleting local file",
	Long:  "Close a GitHub Issue by number. The local Markdown file and index entry are preserved, only the issue state is changed to 'closed'. Use 'igmeek undel <num>' to reopen.",
	Args:  cobra.ExactArgs(1),
	RunE:  runDel,
}

func init() {
	rootCmd.AddCommand(delCmd)
}

func runDel(cmd *cobra.Command, args []string) error {
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
	issue, err := client.CloseIssue(context.Background(), owner, repo, num)
	if err != nil {
		return err
	}

	entries, _ := issueIndex.Load()
	for i, e := range entries {
		if e.IssueNumber == num {
			entries[i].State = "closed"
			if issue.ClosedAt != nil {
				t := issue.ClosedAt.Time
				entries[i].ClosedAt = &t
			}
			break
		}
	}
	issueIndex.Save(entries)

	fmt.Printf("Closed issue #%d: %s\n", num, entry.Title)
	return nil
}
