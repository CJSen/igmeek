package cmd

import "testing"

func TestIssueTitleForNewUsesFileName(t *testing.T) {
	got := issueTitleForNew("/tmp/posts/hello-world.md")
	if got != "hello-world" {
		t.Fatalf("expected hello-world, got %s", got)
	}
}
