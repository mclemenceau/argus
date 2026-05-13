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

	// 1. Fetch current build list.
	var images []buildapi.Image
	if err := sdk.ExecuteActivity(ctx, act.FetchBuildStatus).Get(ctx, &images); err != nil {
		return buildapi.AgentReply{}, err
	}

	// 2. Resolve the query to an image ID via LLM fuzzy match.
	ids := make([]string, len(images))
	for i, img := range images {
		ids[i] = img.ID
	}

	var matchedID string
	if err := sdk.ExecuteActivity(ctx, act.FuzzyMatch, query, ids).Get(ctx, &matchedID); err != nil {
		return buildapi.AgentReply{}, err
	}

	if matchedID == "" {
		return buildapi.AgentReply{
			Summary:    fmt.Sprintf("No image found matching %q. Known images: %v", query, ids),
			Category:   "unknown",
			WorkflowID: sdk.GetInfo(ctx).WorkflowExecution.ID,
		}, nil
	}

	// 3. Find the matched image struct.
	var image buildapi.Image
	for _, img := range images {
		if img.ID == matchedID {
			image = img
			break
		}
	}

	// 4. For failed images: fetch log, analyse, then compose reply.
	if image.Status == "FAILED" {
		var logContent string
		if err := sdk.ExecuteActivity(ctx, act.FetchLog, image.LogURL).Get(ctx, &logContent); err != nil {
			return buildapi.AgentReply{}, err
		}

		var analysis activities.LogAnalysis
		if err := sdk.ExecuteActivity(ctx, act.AnalyzeLog, image.ID, logContent).Get(ctx, &analysis); err != nil {
			return buildapi.AgentReply{}, err
		}

		var reply buildapi.AgentReply
		if err := sdk.ExecuteActivity(ctx, act.ComposeReply, image, &analysis).Get(ctx, &reply); err != nil {
			return buildapi.AgentReply{}, err
		}
		reply.WorkflowID = sdk.GetInfo(ctx).WorkflowExecution.ID
		return reply, nil
	}

	// 5. Non-failed: compose a plain status reply.
	var reply buildapi.AgentReply
	if err := sdk.ExecuteActivity(ctx, act.ComposeReply, image, (*activities.LogAnalysis)(nil)).Get(ctx, &reply); err != nil {
		return buildapi.AgentReply{}, err
	}
	reply.WorkflowID = sdk.GetInfo(ctx).WorkflowExecution.ID
	return reply, nil
}
