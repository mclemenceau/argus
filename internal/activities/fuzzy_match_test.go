package activities

import (
	"context"
	"testing"

	"github.com/mclemenceau/argus/internal/llm"
)

func TestFuzzyMatchReturnsID(t *testing.T) {
	act := &Activities{LLM: &llm.MockLLMClient{Response: `{"matched_id":"ubuntu-desktop-amd64","list":false,"release":""}`}}

	got, err := act.FuzzyMatch(context.Background(), "ubuntu desktop", []string{"ubuntu-desktop-amd64", "ubuntu-server-amd64"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.MatchedID != "ubuntu-desktop-amd64" {
		t.Errorf("got %q, want %q", got.MatchedID, "ubuntu-desktop-amd64")
	}
	if got.ListIntent {
		t.Error("expected ListIntent=false")
	}
}

func TestFuzzyMatchNoMatch(t *testing.T) {
	act := &Activities{LLM: &llm.MockLLMClient{Response: `{"matched_id":"","list":false,"release":""}`}}

	got, err := act.FuzzyMatch(context.Background(), "totally unknown image", []string{"ubuntu-desktop-amd64"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.MatchedID != "" {
		t.Errorf("expected empty MatchedID, got %q", got.MatchedID)
	}
	if got.ListIntent {
		t.Error("expected ListIntent=false")
	}
}

func TestFuzzyMatchListIntent(t *testing.T) {
	act := &Activities{LLM: &llm.MockLLMClient{Response: `{"matched_id":"","list":true,"release":"noble"}`}}

	got, err := act.FuzzyMatch(context.Background(), "list all noble builds", []string{"ubuntu-desktop-amd64"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.MatchedID != "" {
		t.Errorf("expected empty MatchedID, got %q", got.MatchedID)
	}
	if !got.ListIntent {
		t.Error("expected ListIntent=true")
	}
	if got.Release != "noble" {
		t.Errorf("got release %q, want %q", got.Release, "noble")
	}
}

func TestFuzzyMatchStripsCodeFence(t *testing.T) {
	act := &Activities{LLM: &llm.MockLLMClient{Response: "```json\n{\"matched_id\":\"ubuntu-server-amd64\",\"list\":false,\"release\":\"\"}\n```"}}

	got, err := act.FuzzyMatch(context.Background(), "server", []string{"ubuntu-server-amd64"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.MatchedID != "ubuntu-server-amd64" {
		t.Errorf("got %q, want %q", got.MatchedID, "ubuntu-server-amd64")
	}
}

func TestFuzzyMatchInvalidJSON(t *testing.T) {
	act := &Activities{LLM: &llm.MockLLMClient{Response: "not json"}}

	_, err := act.FuzzyMatch(context.Background(), "anything", []string{"ubuntu-desktop-amd64"})
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}
