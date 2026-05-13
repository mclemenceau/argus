package buildapi

import "time"

type Image struct {
	ID         string    `json:"id"`
	Package    string    `json:"package"`
	Series     string    `json:"series"`
	Arch       string    `json:"arch"`
	Status     string    `json:"status"` // BUILDING|SUCCESS|FAILED|CANCELLED
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
	LogURL     string    `json:"log_url"`
}

type ChangeReport struct {
	NewFailures  []ImageDelta `json:"new_failures"`
	Recoveries   []ImageDelta `json:"recoveries"`
	OtherChanges []ImageDelta `json:"other_changes"`
	NewImages    []Image      `json:"new_images"`
}

type ImageDelta struct {
	Image     string    `json:"image"`
	OldStatus string    `json:"old_status"`
	NewStatus string    `json:"new_status"`
	Since     time.Time `json:"since"`
}

type AgentReply struct {
	Summary     string   `json:"summary"`
	Category    string   `json:"category"` // infra|code|dependency|flaky|unknown
	Hypothesis  string   `json:"hypothesis"`
	LogExcerpts []string `json:"log_excerpts"`
	NextAction  string   `json:"next_action"`
	WorkflowID  string   `json:"workflow_id"`
}
