package activities

import (
	"context"
	"testing"

	"github.com/mclemenceau/argus/internal/llm"
)

func TestFuzzyMatchReturnsID(t *testing.T) {
	act := &Activities{LLM: &llm.MockLLMClient{Response: `{"matched_id":"ubuntu-desktop-amd64"}`}}

	got, err := act.FuzzyMatch(context.Background(), "ubuntu desktop", []string{"ubuntu-desktop-amd64", "ubuntu-server-amd64"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "ubuntu-desktop-amd64" {
		t.Errorf("got %q, want %q", got, "ubuntu-desktop-amd64")
	}
}

func TestFuzzyMatchNoMatch(t *testing.T) {
	act := &Activities{LLM: &llm.MockLLMClient{Response: `{"matched_id":""}`}}

	got, err := act.FuzzyMatch(context.Background(), "totally unknown image", []string{"ubuntu-desktop-amd64"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestFuzzyMatchStripsCodeFence(t *testing.T) {
	act := &Activities{LLM: &llm.MockLLMClient{Response: "```json\n{\"matched_id\":\"ubuntu-server-amd64\"}\n```"}}

	got, err := act.FuzzyMatch(context.Background(), "server", []string{"ubuntu-server-amd64"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "ubuntu-server-amd64" {
		t.Errorf("got %q, want %q", got, "ubuntu-server-amd64")
	}
}

func TestFuzzyMatchInvalidJSON(t *testing.T) {
	act := &Activities{LLM: &llm.MockLLMClient{Response: "not json"}}

	_, err := act.FuzzyMatch(context.Background(), "anything", []string{"ubuntu-desktop-amd64"})
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}
