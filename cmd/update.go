package cmd

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/CJSen/igmeek/cli/internal/api"
	"github.com/CJSen/igmeek/cli/internal/config"
	"github.com/CJSen/igmeek/cli/internal/index"
	"github.com/CJSen/igmeek/cli/internal/markdown"
	"github.com/CJSen/igmeek/cli/internal/sync"
	"github.com/google/go-github/v68/github"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update <file|num> [file]",
	Short: "Update an existing issue",
	Long: `Update an existing issue's body and/or labels.

Usage:
  igmeek update <file>          Update issue linked to this file
  igmeek update <num> <file>    Update issue #num with file content

Label options (can be combined with file update):
  --add-tag <tags>              Add labels (comma-separated)
  --remove-tag <tags>           Remove labels (comma-separated)
  --set-tag <tags>              Replace all labels`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runUpdate,
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
		absPath, err := markdown.NormalizePath(args[1])
		if err != nil {
			return err
		}
		issueNum = num
		filePath = absPath
	} else {
		absPath, err := markdown.NormalizePath(args[0])
		if err != nil {
			return err
		}
		result, err := issueIndex.FindFileMatches(absPath)
		if err != nil {
			return err
		}
		if result.Ambiguous {
			message := "存在多个同名文件，请使用 igmeek update <num> <file>"
			if len(result.Candidates) > 0 {
				message += "\n候选 issue：\n" + formatIssueEntries(result.Candidates)
			}
			message += "\n如果你是要新增同名文章，请使用 igmeek new <file>；否则请选择一个要更新的 issue 编号。"
			return errors.New(message)
		}
		if !result.Found {
			message := "未找到对应文件名的 issue 映射，先执行 sync 或显式传入 issue_number"
			if len(result.Suggestions) > 0 {
				message += "\n相近 issue：\n" + formatIssueEntries(result.Suggestions)
			}
			return errors.New(message)
		}
		issueNum = result.Entry.IssueNumber
		filePath = absPath
	}

	_, found := issueIndex.FindByNumber(issueNum)
	if err := issueIndex.LastError(); err != nil {
		return err
	}
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

	remoteIssue, _, err := client.GetIssue(context.Background(), owner, repo, issueNum)
	if err != nil {
		return fmt.Errorf("failed to fetch issue #%d: %w", issueNum, err)
	}

	finalLabels, replaceLabels := resolveUpdatedLabels(remoteIssue, updateAddTag, updateRemoveTag, updateSetTag)

	issue, err := client.EditIssue(context.Background(), owner, repo, issueNum, title, body, finalLabels, replaceLabels)
	if err != nil {
		return err
	}

	entries, err := issueIndex.Load()
	if err != nil {
		return err
	}
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
	if err := issueIndex.Save(entries); err != nil {
		return err
	}

	fmt.Printf("Updated issue #%d: %s\n", issue.GetNumber(), issue.GetTitle())
	return nil
}

func parseTags(s string) []string {
	return parseTagList(s)
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

func labelsAfterAdd(issue *github.Issue, add []string) []string {
	labels := append(issueLabelNames(issue), add...)
	return uniqueStrings(labels)
}

func resolveUpdatedLabels(issue *github.Issue, addRaw, removeRaw, setRaw string) ([]string, bool) {
	replaceLabels := addRaw != "" || removeRaw != "" || setRaw != ""
	if !replaceLabels {
		return nil, false
	}

	current := makeLabelSlice(issueLabelNames(issue))
	working := &github.Issue{Labels: current}

	if addRaw != "" {
		working.Labels = makeLabelSlice(labelsAfterAdd(working, parseTags(addRaw)))
	}
	if removeRaw != "" {
		working.Labels = makeLabelSlice(labelsAfterRemove(working, parseTags(removeRaw)))
	}
	if setRaw != "" {
		working.Labels = makeLabelSlice(parseTags(setRaw))
	}

	return issueLabelNames(working), true
}

func labelsAfterRemove(issue *github.Issue, remove []string) []string {
	removeSet := make(map[string]bool)
	for _, label := range remove {
		removeSet[label] = true
	}

	var remaining []string
	for _, label := range issueLabelNames(issue) {
		if !removeSet[label] {
			remaining = append(remaining, label)
		}
	}
	return remaining
}

func issueLabelNames(issue *github.Issue) []string {
	labels := make([]string, 0, len(issue.Labels))
	for _, label := range issue.Labels {
		labels = append(labels, label.GetName())
	}
	return labels
}

func formatIssueEntries(entries []index.IssueEntry) string {
	lines := make([]string, 0, len(entries))
	for _, entry := range entries {
		line := fmt.Sprintf("- #%d %s", entry.IssueNumber, entry.Title)
		if entry.FilePath != "" {
			line += fmt.Sprintf(" (%s)", entry.FilePath)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}
