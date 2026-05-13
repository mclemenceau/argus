package activities

import (
	"context"
	"fmt"

	"github.com/mclemenceau/argus/internal/buildapi"
)

const composeReplySystem = `You are a helpful Ubuntu release engineer assistant.
Write a concise, human-friendly paragraph (2-4 sentences) summarising the artefact situation.
Be direct and specific — mention the image name, release, version (build date), and current status.
Status values: APPROVED means QA-signed-off, MARKED_AS_FAILED means reviewers flagged a problem,
UNDECIDED means awaiting human review.
Do not use bullet points or headers. Plain prose only.`

// ComposeReply asks the LLM to write a human-readable summary and packages it into an AgentReply.
// Pass a non-nil analysis for failure diagnosis flows; nil for simple status queries.
func (a *Activities) ComposeReply(ctx context.Context, artefact buildapi.Artefact, analysis *LogAnalysis) (buildapi.AgentReply, error) {
	prompt := buildComposePrompt(artefact, analysis)

	summary, err := a.LLM.Complete(ctx, composeReplySystem, prompt)
	if err != nil {
		return buildapi.AgentReply{}, fmt.Errorf("ComposeReply: %w", err)
	}

	reply := buildapi.AgentReply{
		Summary: summary,
	}
	if analysis != nil {
		reply.Category    = analysis.Category
		reply.Hypothesis  = analysis.Hypothesis
		reply.LogExcerpts = analysis.LogExcerpts
		reply.NextAction  = analysis.NextAction
	}
	return reply, nil
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
