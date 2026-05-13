package workflow

import (
	"time"

	sdk "go.temporal.io/sdk/workflow"

	"github.com/mclemenceau/argus/internal/activities"
	"github.com/mclemenceau/argus/internal/buildapi"
	"github.com/mclemenceau/argus/internal/state"
)

func ChangeWatchWorkflow(ctx sdk.Context) error {
	ctx = sdk.WithActivityOptions(ctx, sdk.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
	})

	var act *activities.Activities

	var fresh []buildapi.Image
	if err := sdk.ExecuteActivity(ctx, act.FetchBuildStatus).Get(ctx, &fresh); err != nil {
		return err
	}

	var old []buildapi.Image
	if err := sdk.ExecuteActivity(ctx, act.LoadSnapshot).Get(ctx, &old); err != nil {
		return err
	}

	report := state.Diff(old, fresh)

	if err := sdk.ExecuteActivity(ctx, act.SaveSnapshot, fresh).Get(ctx, nil); err != nil {
		return err
	}

	if hasChanges(report) {
		sdk.GetLogger(ctx).Info("changes detected",
			"new_failures", len(report.NewFailures),
			"recoveries", len(report.Recoveries),
			"other_changes", len(report.OtherChanges),
			"new_images", len(report.NewImages),
		)
		for _, f := range report.NewFailures {
			sdk.GetLogger(ctx).Warn("new failure", "image", f.Image, "was", f.OldStatus)
		}
		for _, r := range report.Recoveries {
			sdk.GetLogger(ctx).Info("recovery", "image", r.Image, "now", r.NewStatus)
		}
	}

	return nil
}

func hasChanges(r buildapi.ChangeReport) bool {
	return len(r.NewFailures)+len(r.Recoveries)+len(r.OtherChanges)+len(r.NewImages) > 0
}
