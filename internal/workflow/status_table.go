package workflow

import (
	"fmt"
	"strings"
	"time"

	sdk "go.temporal.io/sdk/workflow"

	"github.com/mclemenceau/argus/internal/activities"
	"github.com/mclemenceau/argus/internal/buildapi"
	"github.com/mclemenceau/argus/internal/state"
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

	table := formatStatusTable(artefacts)

	if err := sdk.ExecuteActivity(ctx, act.PushToFeed, table).Get(ctx, nil); err != nil {
		sdk.GetLogger(ctx).Warn("StatusTableWorkflow: PushToFeed failed", "error", err)
	}
	return nil
}

// formatStatusTable renders artefacts for the latest release only.
// The full list is available to the LLM via queries.
func formatStatusTable(artefacts []buildapi.Artefact) string {
	latest := state.LatestRelease(artefacts)

	var filtered []buildapi.Artefact
	for _, a := range artefacts {
		if a.Release == latest {
			filtered = append(filtered, a)
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("=== Build Status — %s (%s) ===\n",
		latest, time.Now().UTC().Format("2006-01-02 15:04 UTC")))
	sb.WriteString(fmt.Sprintf("%-45s %-10s %-20s\n", "IMAGE", "VERSION", "STATUS"))
	sb.WriteString(strings.Repeat("─", 78) + "\n")
	for _, a := range filtered {
		sb.WriteString(fmt.Sprintf("%-45s %-10s %-20s\n", a.Name, a.Version, a.Status))
	}
	return sb.String()
}
