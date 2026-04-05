package api

import (
	"context"
	"fmt"

	"github.com/google/go-github/v68/github"
	"golang.org/x/oauth2"
)

type Client struct {
	gh *github.Client
}

func NewClient(token string) *Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	return &Client{
		gh: github.NewClient(tc),
	}
}

func (c *Client) ListIssues(ctx context.Context, owner, repo string) ([]*github.Issue, error) {
	var allIssues []*github.Issue
	opts := &github.IssueListByRepoOptions{
		State: "all",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		issues, resp, err := c.gh.Issues.ListByRepo(ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list issues: %w", err)
		}
		allIssues = append(allIssues, issues...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allIssues, nil
}

func (c *Client) CreateIssue(ctx context.Context, owner, repo, title, body string, labels []string) (*github.Issue, error) {
	issueReq := &github.IssueRequest{
		Title: github.String(title),
		Body:  github.String(body),
	}
	if len(labels) > 0 {
		issueReq.Labels = &labels
	}

	issue, _, err := c.gh.Issues.Create(ctx, owner, repo, issueReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create issue: %w", err)
	}

	return issue, nil
}

func (c *Client) EditIssue(ctx context.Context, owner, repo string, number int, title, body string, labels []string, replaceLabels bool) (*github.Issue, error) {
	issueReq := &github.IssueRequest{}
	if title != "" {
		issueReq.Title = github.String(title)
	}
	if body != "" {
		issueReq.Body = github.String(body)
	}
	if replaceLabels {
		issueReq.Labels = &labels
	}

	issue, _, err := c.gh.Issues.Edit(ctx, owner, repo, number, issueReq)
	if err != nil {
		return nil, fmt.Errorf("failed to edit issue: %w", err)
	}

	return issue, nil
}

func (c *Client) CloseIssue(ctx context.Context, owner, repo string, number int) (*github.Issue, error) {
	issueReq := &github.IssueRequest{
		State: github.String("closed"),
	}

	issue, _, err := c.gh.Issues.Edit(ctx, owner, repo, number, issueReq)
	if err != nil {
		return nil, fmt.Errorf("failed to close issue: %w", err)
	}

	return issue, nil
}

func (c *Client) ReopenIssue(ctx context.Context, owner, repo string, number int) (*github.Issue, error) {
	issueReq := &github.IssueRequest{
		State: github.String("open"),
	}

	issue, _, err := c.gh.Issues.Edit(ctx, owner, repo, number, issueReq)
	if err != nil {
		return nil, fmt.Errorf("failed to reopen issue: %w", err)
	}

	return issue, nil
}

func (c *Client) AddLabels(ctx context.Context, owner, repo string, number int, labels []string) ([]*github.Label, error) {
	result, _, err := c.gh.Issues.AddLabelsToIssue(ctx, owner, repo, number, labels)
	if err != nil {
		return nil, fmt.Errorf("failed to add labels: %w", err)
	}

	return result, nil
}

func (c *Client) ReplaceLabels(ctx context.Context, owner, repo string, number int, labels []string) ([]*github.Label, error) {
	result, _, err := c.gh.Issues.ReplaceLabelsForIssue(ctx, owner, repo, number, labels)
	if err != nil {
		return nil, fmt.Errorf("failed to replace labels: %w", err)
	}

	return result, nil
}

func (c *Client) RemoveLabel(ctx context.Context, owner, repo string, number int, label string) (*github.Response, error) {
	resp, err := c.gh.Issues.RemoveLabelForIssue(ctx, owner, repo, number, label)
	if err != nil {
		return nil, fmt.Errorf("failed to remove label: %w", err)
	}

	return resp, nil
}

func (c *Client) ListLabels(ctx context.Context, owner, repo string) ([]*github.Label, error) {
	var allLabels []*github.Label
	opts := &github.ListOptions{
		PerPage: 100,
	}

	for {
		labels, resp, err := c.gh.Issues.ListLabels(ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list labels: %w", err)
		}
		allLabels = append(allLabels, labels...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allLabels, nil
}

func (c *Client) CreateLabel(ctx context.Context, owner, repo, name string) (*github.Label, error) {
	label := &github.Label{
		Name:  github.String(name),
		Color: github.String("ededed"),
	}

	result, _, err := c.gh.Issues.CreateLabel(ctx, owner, repo, label)
	if err != nil {
		return nil, fmt.Errorf("failed to create label: %w", err)
	}

	return result, nil
}

func (c *Client) GetIssue(ctx context.Context, owner, repo string, number int) (*github.Issue, *github.Response, error) {
	return c.gh.Issues.Get(ctx, owner, repo, number)
}

func (c *Client) VerifyRepo(ctx context.Context, owner, repo string) error {
	_, _, err := c.gh.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return fmt.Errorf("failed to access repository: %w", err)
	}
	return nil
}
