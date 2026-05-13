package activities

import (
	"context"
	"fmt"

	"github.com/mclemenceau/argus/internal/buildapi"
)

const composeReplySystem = `You are a helpful Ubuntu build engineer assistant.
Write a concise, human-friendly paragraph (2-4 sentences) summarising the build situation.
Be direct and specific — mention the image name, current status, and key details.
Do not use bullet points or headers. Plain prose only.`

// ComposeReply asks the LLM to write a human-readable summary and packages it into an AgentReply.
// Pass a non-nil analysis for failure diagnosis flows; nil for simple status queries.
func (a *Activities) ComposeReply(ctx context.Context, image buildapi.Image, analysis *LogAnalysis) (buildapi.AgentReply, error) {
	prompt := buildComposePrompt(image, analysis)

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

func buildComposePrompt(image buildapi.Image, analysis *LogAnalysis) string {
	base := fmt.Sprintf("Image: %s\nPackage: %s\nSeries: %s\nArch: %s\nStatus: %s\nStarted: %s",
		image.ID, image.Package, image.Series, image.Arch, image.Status,
		image.StartedAt.Format("2006-01-02 15:04 UTC"))

	if !image.FinishedAt.IsZero() {
		base += fmt.Sprintf("\nFinished: %s", image.FinishedAt.Format("2006-01-02 15:04 UTC"))
	}

	if analysis != nil {
		base += fmt.Sprintf("\n\nRoot cause analysis:\nCategory: %s\nHypothesis: %s\nNext action: %s",
			analysis.Category, analysis.Hypothesis, analysis.NextAction)
	}

	return base
}
