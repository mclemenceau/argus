package main

import "go.temporal.io/sdk/workflow"

// Stubs for workflows implemented in later blocks.
func StatusTableWorkflow(ctx workflow.Context) error                    { return nil }
func QueryWorkflow(ctx workflow.Context, _ string) (string, error)     { return "", nil }
