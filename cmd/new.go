package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/CJSen/igmeek/cli/internal/api"
	"github.com/CJSen/igmeek/cli/internal/config"
	"github.com/CJSen/igmeek/cli/internal/index"
	"github.com/CJSen/igmeek/cli/internal/markdown"
	"github.com/CJSen/igmeek/cli/internal/sync"
	"github.com/google/go-github/v68/github"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new <file>",
	Short: "Create a new issue from a markdown file",
	Long:  "Create a new GitHub Issue from a local Markdown file. Requires --tag (at least one label) or --notag (create without labels, useful for drafts).",
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
		cleaned := parseTagList(newTags)
		if len(cleaned) == 0 {
			return fmt.Errorf("must specify at least one tag with --tag")
		}
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
	issueTitle := issueTitleForNew(mdFile.AbsPath)

	client := api.NewClient(GetToken())
	var labels []string
	if !newNoTag {
		labels = parseTagList(newTags)
		remoteLabels, err := client.ListLabels(context.Background(), owner, repo)
		if err != nil {
			return fmt.Errorf("failed to list labels: %w", err)
		}

		missing := missingLabelNames(labels, remoteLabels)
		if len(missing) > 0 {
			return fmt.Errorf("labels do not exist in %s: %s. Create them first with 'igmeek label add %s'", cfg.CurrentRepo, strings.Join(missing, ", "), strings.Join(missing, ","))
		}
	}

	issue, err := client.CreateIssue(context.Background(), owner, repo, issueTitle, mdFile.Content, labels)
	if err != nil {
		return err
	}

	repoDir := config.GetRepoDir(globalDir, cfg.CurrentRepo)
	issueIndex := index.NewIssueIndex(repoDir)

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

func issueTitleForNew(absPath string) string {
	return markdown.ExtractTitleFromFileName(absPath)
}

func missingLabelNames(wanted []string, remoteLabels []*github.Label) []string {
	existing := make(map[string]struct{}, len(remoteLabels))
	for _, label := range remoteLabels {
		existing[label.GetName()] = struct{}{}
	}

	var missing []string
	for _, label := range wanted {
		if _, ok := existing[label]; !ok {
			missing = append(missing, label)
		}
	}
	return uniqueStrings(missing)
}
