package cmd

import (
	"reflect"
	"testing"
)

func TestParseTagListSupportsChineseAndEnglishCommas(t *testing.T) {
	got := parseTagList("alpha，beta, gamma")
	want := []string{"alpha", "beta", "gamma"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestParseTagListSupportsMultipleArgs(t *testing.T) {
	got := parseTagList("alpha，beta", "gamma, delta", " ")
	want := []string{"alpha", "beta", "gamma", "delta"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}
