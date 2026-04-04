package sync

import (
	"context"
	"fmt"

	"github.com/CJSen/igmeek/internal/api"
	"github.com/CJSen/igmeek/internal/index"
	"github.com/google/go-github/v68/github"
)

type SyncResult struct {
	IssuesCount int
	LabelsCount int
}

func SyncAll(ctx context.Context, client *api.Client, owner, repo string, repoDir string) (*SyncResult, error) {
	issues, err := client.ListIssues(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch issues: %w", err)
	}

	labels, err := client.ListLabels(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch labels: %w", err)
	}

	issueIndex := index.NewIssueIndex(repoDir)
	tagCache := index.NewTagCache(repoDir)

	var issueEntries []index.IssueEntry
	for _, issue := range issues {
		issueEntries = append(issueEntries, convertIssueToEntry(issue))
	}

	if err := issueIndex.Save(issueEntries); err != nil {
		return nil, fmt.Errorf("failed to save issue index: %w", err)
	}

	var tagEntries []index.TagEntry
	for _, label := range labels {
		tagEntries = append(tagEntries, convertLabelToEntry(label))
	}

	if err := tagCache.Save(tagEntries); err != nil {
		return nil, fmt.Errorf("failed to save tag cache: %w", err)
	}

	return &SyncResult{
		IssuesCount: len(issueEntries),
		LabelsCount: len(tagEntries),
	}, nil
}

func convertIssueToEntry(issue *github.Issue) index.IssueEntry {
	var labels []string
	for _, label := range issue.Labels {
		if label.Name != nil {
			labels = append(labels, *label.Name)
		}
	}

	entry := index.IssueEntry{
		IssueNumber: issue.GetNumber(),
		Title:       issue.GetTitle(),
		Labels:      labels,
		State:       issue.GetState(),
		URL:         issue.GetURL(),
		HTMLURL:     issue.GetHTMLURL(),
	}

	if issue.CreatedAt != nil {
		t := issue.CreatedAt.Time
		entry.CreatedAt = &t
	}
	if issue.UpdatedAt != nil {
		t := issue.UpdatedAt.Time
		entry.UpdatedAt = &t
	}
	if issue.ClosedAt != nil {
		t := issue.ClosedAt.Time
		entry.ClosedAt = &t
	}

	return entry
}

func convertLabelToEntry(label *github.Label) index.TagEntry {
	return index.TagEntry{
		Name:  label.GetName(),
		Color: label.GetColor(),
	}
}
