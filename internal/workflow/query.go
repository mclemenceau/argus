package workflow

import (
	"fmt"
	"time"

	sdk "go.temporal.io/sdk/workflow"

	"github.com/mclemenceau/argus/internal/activities"
	"github.com/mclemenceau/argus/internal/buildapi"
)

func QueryWorkflow(ctx sdk.Context, query string) (buildapi.AgentReply, error) {
	ctx = sdk.WithActivityOptions(ctx, sdk.ActivityOptions{
		StartToCloseTimeout: 60 * time.Second,
	})

	var act *activities.Activities

	// 1. Fetch current artefact list.
	var artefacts []buildapi.Artefact
	if err := sdk.ExecuteActivity(ctx, act.FetchBuildStatus).Get(ctx, &artefacts); err != nil {
		return buildapi.AgentReply{}, err
	}

	// 2. Resolve the query to an artefact name via LLM fuzzy match.
	names := make([]string, len(artefacts))
	for i, a := range artefacts {
		names[i] = a.Name
	}

	var matchedName string
	if err := sdk.ExecuteActivity(ctx, act.FuzzyMatch, query, names).Get(ctx, &matchedName); err != nil {
		return buildapi.AgentReply{}, err
	}

	if matchedName == "" {
		return buildapi.AgentReply{
			Summary:    fmt.Sprintf("No artefact found matching %q. Known artefacts: %v", query, names),
			Category:   "unknown",
			WorkflowID: sdk.GetInfo(ctx).WorkflowExecution.ID,
		}, nil
	}

	// 3. Find the matched artefact struct.
	var artefact buildapi.Artefact
	for _, a := range artefacts {
		if a.Name == matchedName {
			artefact = a
			break
		}
	}

	// 4. Compose a reply based on artefact status.
	// Log analysis is not yet wired (no build log URLs from Test Observer).
	var reply buildapi.AgentReply
	if err := sdk.ExecuteActivity(ctx, act.ComposeReply, artefact, (*activities.LogAnalysis)(nil)).Get(ctx, &reply); err != nil {
		return buildapi.AgentReply{}, err
	}
	reply.WorkflowID = sdk.GetInfo(ctx).WorkflowExecution.ID
	return reply, nil
}
