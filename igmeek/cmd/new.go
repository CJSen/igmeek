package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/CJSen/igmeek/internal/api"
	"github.com/CJSen/igmeek/internal/config"
	"github.com/CJSen/igmeek/internal/index"
	"github.com/CJSen/igmeek/internal/markdown"
	"github.com/CJSen/igmeek/internal/sync"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new <file>",
	Short: "Create a new issue from a markdown file",
	Long:  "Create a new GitHub Issue from a local Markdown file. The file's first H1 heading becomes the issue title, and the rest becomes the body. Requires --tag (at least one label) or --notag (create without labels, useful for drafts).",
	Args:  cobra.ExactArgs(1),
	RunE:  runNew,
}

var (
	newTags  string
	newNoTag bool
)

func init() {
	rootCmd.AddCommand(newCmd)
	newCmd.Flags().StringVar(&newTags, "tag", "", "Comma-separated labels (required unless --notag)")
	newCmd.Flags().BoolVar(&newNoTag, "notag", false, "Create issue without labels")
}

func runNew(cmd *cobra.Command, args []string) error {
	if !newNoTag && newTags == "" {
		return fmt.Errorf("must specify --tag (at least one) or --notag")
	}

	if !newNoTag {
		tags := strings.Split(newTags, ",")
		var cleaned []string
		for _, t := range tags {
			t = strings.TrimSpace(t)
			if t != "" {
				cleaned = append(cleaned, t)
			}
		}
		if len(cleaned) == 0 {
			return fmt.Errorf("must specify at least one tag with --tag")
		}
		newTags = strings.Join(cleaned, ",")
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

	mdFile, err := markdown.ReadFile(args[0])
	if err != nil {
		return err
	}

	client := api.NewClient(GetToken())
	issue, err := client.CreateIssue(context.Background(), owner, repo, mdFile.Title, mdFile.Content)
	if err != nil {
		return err
	}

	repoDir := config.GetRepoDir(globalDir, cfg.CurrentRepo)
	issueIndex := index.NewIssueIndex(repoDir)

	var labels []string
	if !newNoTag {
		labelList := strings.Split(newTags, ",")
		_, err = client.AddLabels(context.Background(), owner, repo, issue.GetNumber(), labelList)
		if err != nil {
			return fmt.Errorf("failed to add labels: %w", err)
		}
		labels = labelList
	}

	entries, _ := issueIndex.Load()
	newEntry := index.IssueEntry{
		IssueNumber: issue.GetNumber(),
		FilePath:    mdFile.AbsPath,
		Title:       issue.GetTitle(),
		Labels:      labels,
		State:       issue.GetState(),
		URL:         issue.GetURL(),
		HTMLURL:     issue.GetHTMLURL(),
	}
	if issue.CreatedAt != nil {
		t := issue.CreatedAt.Time
		newEntry.CreatedAt = &t
	}
	if issue.UpdatedAt != nil {
		t := issue.UpdatedAt.Time
		newEntry.UpdatedAt = &t
	}
	entries = append(entries, newEntry)
	if err := issueIndex.Save(entries); err != nil {
		return fmt.Errorf("failed to update index: %w", err)
	}

	fmt.Printf("Created issue #%d: %s\n", issue.GetNumber(), issue.GetTitle())
	fmt.Printf("URL: %s\n", issue.GetHTMLURL())
	return nil
}
