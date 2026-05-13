package workflow

import (
	"time"

	sdk "go.temporal.io/sdk/workflow"

	"github.com/mclemenceau/argus/internal/activities"
	"github.com/mclemenceau/argus/internal/buildapi"
)

func StatusTableWorkflow(ctx sdk.Context) error {
	ctx = sdk.WithActivityOptions(ctx, sdk.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
	})

	var act *activities.Activities

	var artefacts []buildapi.Artefact
	if err := sdk.ExecuteActivity(ctx, act.FetchBuildStatus).Get(ctx, &artefacts); err != nil {
		return err
	}

	var table string
	if err := sdk.ExecuteActivity(ctx, act.FormatStatusTable, artefacts).Get(ctx, &table); err != nil {
		return err
	}

	if err := sdk.ExecuteActivity(ctx, act.PushToFeed, table).Get(ctx, nil); err != nil {
		sdk.GetLogger(ctx).Warn("StatusTableWorkflow: PushToFeed failed", "error", err)
	}
	return nil
}
