package cmd

import (
	"reflect"
	"strings"
	"testing"

	"github.com/CJSen/igmeek/cli/internal/index"
	"github.com/google/go-github/v68/github"
)

func TestLabelsAfterAddUsesRemoteIssueLabels(t *testing.T) {
	issue := &github.Issue{
		Labels: []*github.Label{
			{Name: github.String("remote")},
			{Name: github.String("shared")},
		},
	}

	got := labelsAfterAdd(issue, []string{"new", "shared"})
	want := []string{"remote", "shared", "new"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestLabelsAfterRemoveUsesRemoteIssueLabels(t *testing.T) {
	issue := &github.Issue{
		Labels: []*github.Label{
			{Name: github.String("remote")},
			{Name: github.String("stale")},
		},
	}

	got := labelsAfterRemove(issue, []string{"stale"})
	want := []string{"remote"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestResolveUpdatedLabelsReturnsNilWhenNoFlags(t *testing.T) {
	issue := &github.Issue{Labels: []*github.Label{{Name: github.String("remote")}}}

	got, replace := resolveUpdatedLabels(issue, "", "", "")
	if replace {
		t.Fatal("expected replace=false")
	}
	if got != nil {
		t.Fatalf("expected nil labels, got %v", got)
	}
}

func TestResolveUpdatedLabelsAppliesFlagsInCurrentOrder(t *testing.T) {
	issue := &github.Issue{
		Labels: []*github.Label{
			{Name: github.String("remote")},
			{Name: github.String("shared")},
		},
	}

	got, replace := resolveUpdatedLabels(issue, "new，shared", "remote", "final,done")
	want := []string{"final", "done"}
	if !replace {
		t.Fatal("expected replace=true")
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestFormatIssueEntriesIncludesNumberTitleAndPath(t *testing.T) {
	formatted := formatIssueEntries([]index.IssueEntry{{
		IssueNumber: 12,
		Title:       "shared title",
		FilePath:    "posts/shared.md",
	}})

	if !strings.Contains(formatted, "#12") {
		t.Fatalf("expected issue number in output, got %q", formatted)
	}
	if !strings.Contains(formatted, "shared title") {
		t.Fatalf("expected title in output, got %q", formatted)
	}
	if !strings.Contains(formatted, "posts/shared.md") {
		t.Fatalf("expected file path in output, got %q", formatted)
	}
}
