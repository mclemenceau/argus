package main

import (
	"context"
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/mclemenceau/argus/internal/activities"
	"github.com/mclemenceau/argus/internal/buildapi"
	"github.com/mclemenceau/argus/internal/config"
	"github.com/mclemenceau/argus/internal/llm"
	"github.com/mclemenceau/argus/internal/state"
	argusworkflow "github.com/mclemenceau/argus/internal/workflow"
)

const taskQueue = "argus"

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	c, err := client.Dial(client.Options{
		HostPort: cfg.TemporalHost,
	})
	if err != nil {
		log.Fatalf("worker: dial temporal: %v", err)
	}
	defer c.Close()

	act := &activities.Activities{
		Builds:   buildapi.NewMockClient(),
		Snapshot: state.New("state/snapshot.json"),
		LLM:      llm.NewOpenRouterClient(cfg.OpenRouterAPIKey),
	}

	w := worker.New(c, taskQueue, worker.Options{})

	w.RegisterWorkflow(argusworkflow.ChangeWatchWorkflow)
	w.RegisterWorkflow(argusworkflow.QueryWorkflow)
	w.RegisterWorkflow(StatusTableWorkflow)

	w.RegisterActivity(act)

	startCronWorkflows(c)

	log.Printf("worker started on task queue %q (temporal: %s)", taskQueue, cfg.TemporalHost)
	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalf("worker: %v", err)
	}
}

func startCronWorkflows(c client.Client) {
	_, err := c.ExecuteWorkflow(
		context.Background(),
		client.StartWorkflowOptions{
			ID:           "change-watch",
			TaskQueue:    taskQueue,
			CronSchedule: "*/10 * * * *",
		},
		argusworkflow.ChangeWatchWorkflow,
	)
	if err != nil {
		// Cron may already be running from a previous worker start — not fatal.
		log.Printf("note: change-watch cron start: %v", err)
	} else {
		log.Println("change-watch cron scheduled (every 10 min)")
	}
}
