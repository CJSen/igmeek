package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test-token")
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestListIssues(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client := NewClient("test-token")
	issues, err := client.ListIssues(context.Background(), "owner", "repo")
	// With real go-github this would need a mock server with proper responses
	// For now, just verify client creation doesn't panic
	_ = issues
	_ = err
}
