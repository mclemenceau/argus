package workflow

import (
	"strings"
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

	// 2. Resolve query intent via LLM fuzzy match.
	names := make([]string, len(artefacts))
	for i, a := range artefacts {
		names[i] = a.Name
	}

	var match activities.MatchResult
	if err := sdk.ExecuteActivity(ctx, act.FuzzyMatch, query, names).Get(ctx, &match); err != nil {
		return buildapi.AgentReply{}, err
	}

	// 3a. List intent — filter artefacts and produce a markdown table.
	if match.ListIntent {
		filtered := filterByRelease(artefacts, match.Release)
		var reply buildapi.AgentReply
		if err := sdk.ExecuteActivity(ctx, act.ComposeListReply, query, filtered).Get(ctx, &reply); err != nil {
			return buildapi.AgentReply{}, err
		}
		reply.WorkflowID = sdk.GetInfo(ctx).WorkflowExecution.ID
		return reply, nil
	}

	// 3b. No match for a single-artefact query.
	if match.MatchedID == "" {
		return buildapi.AgentReply{
			Summary:    "I couldn't find an artefact matching that query. Try asking for a specific image name or a release overview.",
			WorkflowID: sdk.GetInfo(ctx).WorkflowExecution.ID,
		}, nil
	}

	// 4. Find the matched artefact struct.
	var artefact buildapi.Artefact
	for _, a := range artefacts {
		if a.Name == match.MatchedID {
			artefact = a
			break
		}
	}

	// 5. Compose a reply based on artefact status.
	var reply buildapi.AgentReply
	if err := sdk.ExecuteActivity(ctx, act.ComposeReply, artefact, (*activities.LogAnalysis)(nil)).Get(ctx, &reply); err != nil {
		return buildapi.AgentReply{}, err
	}
	reply.WorkflowID = sdk.GetInfo(ctx).WorkflowExecution.ID
	return reply, nil
}

func filterByRelease(artefacts []buildapi.Artefact, release string) []buildapi.Artefact {
	if release == "" {
		return artefacts
	}
	r := strings.ToLower(release)
	out := make([]buildapi.Artefact, 0, len(artefacts))
	for _, a := range artefacts {
		if strings.ToLower(a.Release) == r {
			out = append(out, a)
		}
	}
	return out
}
