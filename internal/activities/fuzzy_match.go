package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// MatchResult is the structured output of FuzzyMatch.
type MatchResult struct {
	MatchedID  string // non-empty when a single artefact matched
	ListIntent bool   // true when the query asks for a list/overview
	Release    string // release filter for list queries (may be empty = all)
}

const fuzzyMatchSystem = `You are a helpful assistant that matches user queries to Ubuntu artefact names.
Given a user query and a list of known artefact names, respond with JSON only — no markdown, no extra text.

Rules:
- If the query targets a single specific artefact, return {"matched_id":"<name>","list":false,"release":""}.
- If the query asks for a list, overview, or status of multiple artefacts (e.g. "list all builds",
  "status of noble", "what failed", "show me everything"), return {"matched_id":"","list":true,"release":"<ubuntu-release-or-empty>"}.
- If nothing matches and it is not a list query, return {"matched_id":"","list":false,"release":""}.`

// FuzzyMatch asks the LLM to resolve a free-form query to a MatchResult.
func (a *Activities) FuzzyMatch(ctx context.Context, query string, imageIDs []string) (MatchResult, error) {
	prompt := fmt.Sprintf("User query: %q\n\nKnown artefact names:\n%s",
		query, strings.Join(imageIDs, "\n"))

	raw, err := a.LLM.Complete(ctx, fuzzyMatchSystem, prompt)
	if err != nil {
		return MatchResult{}, fmt.Errorf("FuzzyMatch: %w", err)
	}

	var result struct {
		MatchedID string `json:"matched_id"`
		List      bool   `json:"list"`
		Release   string `json:"release"`
	}
	if err := json.Unmarshal([]byte(stripCodeFence(raw)), &result); err != nil {
		return MatchResult{}, fmt.Errorf("FuzzyMatch: parse response: %w", err)
	}
	return MatchResult{
		MatchedID:  result.MatchedID,
		ListIntent: result.List,
		Release:    result.Release,
	}, nil
}
