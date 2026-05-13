package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

const fuzzyMatchSystem = `You are a helpful assistant that matches user queries to Ubuntu image IDs.
Given a user query and a list of known image IDs, return the single best match as JSON.
If nothing is a reasonable match, return an empty string for matched_id.
Respond with valid JSON only — no markdown, no extra text:
{"matched_id": "the-matching-id-or-empty-string"}`

// FuzzyMatch asks the LLM to resolve a free-form query to one of the known image IDs.
// Returns an empty string if no match is found.
func (a *Activities) FuzzyMatch(ctx context.Context, query string, imageIDs []string) (string, error) {
	prompt := fmt.Sprintf("User query: %q\n\nKnown image IDs:\n%s",
		query, strings.Join(imageIDs, "\n"))

	raw, err := a.LLM.Complete(ctx, fuzzyMatchSystem, prompt)
	if err != nil {
		return "", fmt.Errorf("FuzzyMatch: %w", err)
	}

	var result struct {
		MatchedID string `json:"matched_id"`
	}
	if err := json.Unmarshal([]byte(stripCodeFence(raw)), &result); err != nil {
		return "", fmt.Errorf("FuzzyMatch: parse response: %w", err)
	}
	return result.MatchedID, nil
}
