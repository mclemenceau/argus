package activities

import (
	"context"
	"fmt"
	"strings"

	"github.com/mclemenceau/argus/internal/buildapi"
)

const composeReplySystem = `You are a helpful Ubuntu release engineer assistant.
Write a concise, human-friendly response (2-4 sentences) summarising the artefact situation.
Be direct and specific — mention the image name, release, version (build date), and current status.
Status values: APPROVED means QA-signed-off, MARKED_AS_FAILED means reviewers flagged a problem,
UNDECIDED means awaiting human review.
You may use Markdown: **bold** and inline code. No headers or bullet points.`

const composeListSystem = `You are a helpful Ubuntu release engineer assistant.
Given a list of Ubuntu image artefacts and the user's query, respond with:
1. One brief sentence summarising the overview (e.g. how many artefacts, any failures).
2. A Markdown table with columns: | Name | Release | Version | Status | Stage |
Sort and filter the table according to the user's intent — put failures first unless specified otherwise.
Status rendering: APPROVED → ✅ approved, MARKED_AS_FAILED → ❌ failed, UNDECIDED → ⏳ pending.
Use only Markdown. No HTML.`

// ComposeReply asks the LLM to write a human-readable summary and packages it into an AgentReply.
// Pass a non-nil analysis for failure diagnosis flows; nil for simple status queries.
func (a *Activities) ComposeReply(ctx context.Context, artefact buildapi.Artefact, analysis *LogAnalysis) (buildapi.AgentReply, error) {
	prompt := buildComposePrompt(artefact, analysis)

	summary, err := a.LLM.Complete(ctx, composeReplySystem, prompt)
	if err != nil {
		return buildapi.AgentReply{}, fmt.Errorf("ComposeReply: %w", err)
	}

	reply := buildapi.AgentReply{Summary: summary}
	if analysis != nil {
		reply.Category    = analysis.Category
		reply.Hypothesis  = analysis.Hypothesis
		reply.LogExcerpts = analysis.LogExcerpts
		reply.NextAction  = analysis.NextAction
	} else {
		reply.Category = categoryFromStatus(artefact.Status)
	}
	return reply, nil
}

// ComposeListReply asks the LLM to produce a markdown table for a list of artefacts.
func (a *Activities) ComposeListReply(ctx context.Context, query string, artefacts []buildapi.Artefact) (buildapi.AgentReply, error) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("User query: %q\n\nArtefacts (%d):\n", query, len(artefacts)))
	sb.WriteString("Name | Release | Version | Status | Stage\n")
	for _, art := range artefacts {
		sb.WriteString(fmt.Sprintf("%s | %s | %s | %s | %s\n",
			art.Name, art.Release, art.Version, art.Status, art.Stage))
	}

	summary, err := a.LLM.Complete(ctx, composeListSystem, sb.String())
	if err != nil {
		return buildapi.AgentReply{}, fmt.Errorf("ComposeListReply: %w", err)
	}
	return buildapi.AgentReply{Summary: summary}, nil
}

func buildComposePrompt(artefact buildapi.Artefact, analysis *LogAnalysis) string {
	base := fmt.Sprintf(
		"Artefact: %s\nRelease: %s\nVersion: %s\nOS: %s\nStage: %s\nStatus: %s",
		artefact.Name, artefact.Release, artefact.Version,
		artefact.OS, artefact.Stage, artefact.Status,
	)

	if artefact.ImageURL != "" {
		base += fmt.Sprintf("\nImage URL: %s", artefact.ImageURL)
	}

	if analysis != nil {
		base += fmt.Sprintf("\n\nRoot cause analysis:\nCategory: %s\nHypothesis: %s\nNext action: %s",
			analysis.Category, analysis.Hypothesis, analysis.NextAction)
	}

	return base
}

func categoryFromStatus(status string) string {
	switch status {
	case "APPROVED":
		return "approved"
	case "MARKED_AS_FAILED":
		return "failed"
	case "UNDECIDED":
		return "pending"
	default:
		return ""
	}
}
