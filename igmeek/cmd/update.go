package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/CJSen/igmeek/internal/api"
	"github.com/CJSen/igmeek/internal/config"
	"github.com/CJSen/igmeek/internal/index"
	"github.com/CJSen/igmeek/internal/markdown"
	"github.com/CJSen/igmeek/internal/sync"
	"github.com/google/go-github/v68/github"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update <file|num> [file]",
	Short: "Update an existing issue",
	Args:  cobra.RangeArgs(1, 2),
	RunE:  runUpdate,
}

var (
	updateAddTag    string
	updateRemoveTag string
	updateSetTag    string
)

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().StringVar(&updateAddTag, "add-tag", "", "Add labels (comma-separated)")
	updateCmd.Flags().StringVar(&updateRemoveTag, "remove-tag", "", "Remove labels (comma-separated)")
	updateCmd.Flags().StringVar(&updateSetTag, "set-tag", "", "Replace all labels (comma-separated)")
}

func runUpdate(cmd *cobra.Command, args []string) error {
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
	client := api.NewClient(GetToken())

	var issueNum int
	var filePath string

	if len(args) == 2 {
		num, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("first argument must be an issue number when providing two arguments")
		}
		issueNum = num
		filePath = args[1]
	} else {
		absPath, err := markdown.NormalizePath(args[0])
		if err != nil {
			return err
		}
		entry, ok := issueIndex.FindByFilePath(absPath)
		if !ok {
			return fmt.Errorf("issue not found for file: %s. Run 'igmeek sync' to refresh index", absPath)
		}
		issueNum = entry.IssueNumber
		filePath = absPath
	}

	_, found := issueIndex.FindByNumber(issueNum)
	if !found {
		return fmt.Errorf("issue #%d not found in index. Run 'igmeek sync' to refresh", issueNum)
	}

	var body string
	title := ""
	if filePath != "" {
		mdFile, err := markdown.ReadFile(filePath)
		if err != nil {
			return err
		}
		body = mdFile.Content
		title = mdFile.Title
	}

	issue, err := client.EditIssue(context.Background(), owner, repo, issueNum, title, body)
	if err != nil {
		return err
	}

	if updateAddTag != "" {
		tags := parseTags(updateAddTag)
		existing, _ := issueIndex.FindByNumber(issueNum)
		allTags := append(existing.Labels, tags...)
		allTags = uniqueStrings(allTags)
		_, err = client.ReplaceLabels(context.Background(), owner, repo, issueNum, allTags)
		if err != nil {
			return fmt.Errorf("failed to add labels: %w", err)
		}
		issue.Labels = makeLabelSlice(allTags)
	}

	if updateRemoveTag != "" {
		tags := parseTags(updateRemoveTag)
		existing, _ := issueIndex.FindByNumber(issueNum)
		var remaining []string
		removeSet := make(map[string]bool)
		for _, t := range tags {
			removeSet[t] = true
		}
		for _, t := range existing.Labels {
			if !removeSet[t] {
				remaining = append(remaining, t)
			}
		}
		_, err = client.ReplaceLabels(context.Background(), owner, repo, issueNum, remaining)
		if err != nil {
			return fmt.Errorf("failed to remove labels: %w", err)
		}
		issue.Labels = makeLabelSlice(remaining)
	}

	if updateSetTag != "" {
		tags := parseTags(updateSetTag)
		_, err = client.ReplaceLabels(context.Background(), owner, repo, issueNum, tags)
		if err != nil {
			return fmt.Errorf("failed to set labels: %w", err)
		}
		issue.Labels = makeLabelSlice(tags)
	}

	entries, _ := issueIndex.Load()
	for i, entry := range entries {
		if entry.IssueNumber == issueNum {
			entries[i].FilePath = filePath
			if filePath != "" {
				entries[i].FilePath = filePath
			}
			entries[i].Title = issue.GetTitle()
			entries[i].State = issue.GetState()
			var labels []string
			for _, l := range issue.Labels {
				labels = append(labels, l.GetName())
			}
			entries[i].Labels = labels
			if issue.UpdatedAt != nil {
				t := issue.UpdatedAt.Time
				entries[i].UpdatedAt = &t
			}
			break
		}
	}
	issueIndex.Save(entries)

	fmt.Printf("Updated issue #%d: %s\n", issue.GetNumber(), issue.GetTitle())
	return nil
}

func parseTags(s string) []string {
	var result []string
	for _, t := range strings.Split(s, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			result = append(result, t)
		}
	}
	return result
}

func uniqueStrings(s []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, v := range s {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}

func makeLabelSlice(names []string) []*github.Label {
	var labels []*github.Label
	for _, name := range names {
		labels = append(labels, &github.Label{Name: github.String(name)})
	}
	return labels
}
