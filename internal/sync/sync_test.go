package sync

import (
	"testing"

	"github.com/google/go-github/v68/github"
)

func TestConvertIssueToEntry(t *testing.T) {
	ghIssue := &github.Issue{
		Number: github.Int(42),
		Title:  github.String("Test Issue"),
		State:  github.String("open"),
	}

	entry := convertIssueToEntry(ghIssue)

	if entry.IssueNumber != 42 {
		t.Errorf("expected number 42, got %d", entry.IssueNumber)
	}
	if entry.Title != "Test Issue" {
		t.Errorf("expected title 'Test Issue', got %s", entry.Title)
	}
	if entry.State != "open" {
		t.Errorf("expected state 'open', got %s", entry.State)
	}
}

func TestConvertLabelToEntry(t *testing.T) {
	ghLabel := &github.Label{
		Name:  github.String("tech"),
		Color: github.String("0075ca"),
	}

	entry := convertLabelToEntry(ghLabel)

	if entry.Name != "tech" {
		t.Errorf("expected name 'tech', got %s", entry.Name)
	}
	if entry.Color != "0075ca" {
		t.Errorf("expected color '0075ca', got %s", entry.Color)
	}
}
