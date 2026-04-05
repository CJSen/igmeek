package cmd

import (
	"reflect"
	"testing"

	"github.com/google/go-github/v68/github"
)

func TestIssueTitleForNewUsesFileName(t *testing.T) {
	got := issueTitleForNew("/tmp/posts/hello-world.md")
	if got != "hello-world" {
		t.Fatalf("expected hello-world, got %s", got)
	}
}

func TestMissingLabelNamesReturnsOnlyMissingLabels(t *testing.T) {
	remote := []*github.Label{
		{Name: github.String("blog")},
		{Name: github.String("tech")},
	}

	got := missingLabelNames([]string{"blog", "draft", "draft", "weekly"}, remote)
	want := []string{"draft", "weekly"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}
