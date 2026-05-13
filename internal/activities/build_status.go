package activities

import (
	"context"
	"fmt"

	"github.com/mclemenceau/argus/internal/buildapi"
	"github.com/mclemenceau/argus/internal/llm"
	"github.com/mclemenceau/argus/internal/state"
)

// Activities holds the dependencies injected at worker startup.
type Activities struct {
	Artefacts buildapi.ArtefactClient
	Snapshot  *state.Snapshot
	LLM       llm.LLMClient
	FeedURL   string // base URL of the HTTP server for SSE push
}

func (a *Activities) FetchBuildStatus(ctx context.Context) ([]buildapi.Artefact, error) {
	artefacts, err := a.Artefacts.FetchArtefacts(ctx)
	if err != nil {
		return nil, fmt.Errorf("FetchBuildStatus: %w", err)
	}
	return artefacts, nil
}

func (a *Activities) LoadSnapshot(_ context.Context) ([]buildapi.Artefact, error) {
	artefacts, err := a.Snapshot.Read()
	if err != nil {
		return nil, fmt.Errorf("LoadSnapshot: %w", err)
	}
	return artefacts, nil
}

func (a *Activities) SaveSnapshot(_ context.Context, artefacts []buildapi.Artefact) error {
	if err := a.Snapshot.Write(artefacts); err != nil {
		return fmt.Errorf("SaveSnapshot: %w", err)
	}
	return nil
}
