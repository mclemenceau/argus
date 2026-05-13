# ARGUS — Design Reference

## What this is

An AI-powered release monitoring agent for Ubuntu image build pipelines.
ARGUS runs two concurrent modes:

**Proactive (Temporal cron workflows — no human trigger):**
- Every 10 min: fetch artefacts → diff against local snapshot → push change report to chat if anything changed
- Every 6 h: fetch artefacts → format markdown status table → push to chat

**Reactive (human-triggered via Web UI chat):**
- Natural language Q&A against the live snapshot
- The LLM receives the full artefact table and the question in a single call — it decides whether to answer in prose (single image) or produce a sorted markdown table (list/overview)

## Tech stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.21+ |
| Workflow orchestration | Temporal (`temporalio/auto-setup`) |
| LLM | OpenRouter API — model configurable via `LLM_MODEL` env var |
| Pipeline data | Ubuntu Test Observer API (`https://tests-api.ubuntu.com`) |
| Web server | Gin |
| UI | Single `index.html` — Vanilla Framework CSS, plain JS |
| State | `state/snapshot.json` — atomic write (tmp → rename), no database |

## Project structure

```
cmd/
  server/main.go          Gin HTTP gateway: POST /query, GET /feed (SSE),
                          POST /internal/push, GET / (static UI)
  worker/main.go          Temporal worker entrypoint; registers all workflows
                          and activities; starts cron workflows on boot
internal/
  workflow/
    change_watch.go       10-min cron: fetch → diff snapshot → push if changed
    status_table.go       6-h cron: fetch → markdown table → push to feed
    query.go              On-demand: load snapshot → AnswerQuery (single LLM call)
  activities/
    build_status.go       FetchBuildStatus (GET Test Observer API → []Artefact)
                          FormatStatusTable (markdown table with age + emoji status)
                          LoadSnapshot / SaveSnapshot (read/write snapshot.json)
    compose_reply.go      AnswerQuery (full snapshot + query → LLM → markdown reply)
                          imageAge (YYYYMMDD[.N] → human-readable age string)
    analyze_log.go        AnalyzeLog (build log → LLM → root-cause JSON)
    fetch_log.go          FetchLog (GET log URL → last 200 lines)
    push_feed.go          PushToFeed (POST /internal/push to fan out to SSE clients)
  buildapi/
    client.go             ArtefactClient interface + HTTPClient (Test Observer)
    types.go              Shared data types (Artefact, ChangeReport, AgentReply, …)
  llm/
    openrouter.go         LLMClient interface + OpenRouterClient + MockLLMClient
  state/
    snapshot.go           Atomic JSON read/write; Diff logic; LatestRelease helper
  config/
    config.go             Env var loading with defaults; fails fast if key missing
web/
  index.html              Single-page chat UI (dark Ubuntu nav, markdown renderer,
                          SSE feed messages inline with chat replies)
```

## Core data types

```go
// Artefact mirrors the Test Observer API response for the image family.
type Artefact struct {
    ID       int    `json:"id"`
    Name     string `json:"name"`
    Version  string `json:"version"` // YYYYMMDD or YYYYMMDD.N (respin)
    OS       string `json:"os"`      // product / Ubuntu variant
    Release  string `json:"release"` // e.g. "noble", "oracular"
    Stage    string `json:"stage"`   // pending | current
    Status   string `json:"status"`  // APPROVED | MARKED_AS_FAILED | UNDECIDED
    Archived bool   `json:"archived"`
    ImageURL string `json:"image_url"`
}

type ChangeReport struct {
    NewFailures  []ArtefactDelta `json:"new_failures"`
    Recoveries   []ArtefactDelta `json:"recoveries"`
    OtherChanges []ArtefactDelta `json:"other_changes"`
    NewArtefacts []Artefact      `json:"new_artefacts"`
}

type ArtefactDelta struct {
    Name      string `json:"name"`
    Release   string `json:"release"`
    Version   string `json:"version"`
    OldStatus string `json:"old_status"`
    NewStatus string `json:"new_status"`
}

type AgentReply struct {
    Summary     string   `json:"summary"`      // markdown; rendered in the UI
    Category    string   `json:"category"`     // approved|failed|pending|infra|code|…
    Hypothesis  string   `json:"hypothesis"`   // set by AnalyzeLog flow
    LogExcerpts []string `json:"log_excerpts"` // set by AnalyzeLog flow
    NextAction  string   `json:"next_action"`  // set by AnalyzeLog flow
    WorkflowID  string   `json:"workflow_id"`
}
```

## Workflow data flows

### ChangeWatchWorkflow (every 10 min)
```
FetchBuildStatus → LoadSnapshot → Diff → SaveSnapshot
                                       └─ if changes → PushToFeed (markdown change report)
```

### StatusTableWorkflow (every 6 h)
```
FetchBuildStatus → FormatStatusTable → PushToFeed (markdown table)
```

### QueryWorkflow (on demand)
```
LoadSnapshot ──► (empty? → FetchBuildStatus) ──► AnswerQuery (LLM) → AgentReply
```
The snapshot is the single source of truth for queries. `AnswerQuery` sends
the full artefact table to the LLM in one call; the model decides how to answer
based on the question (prose for a single image, sorted table for an overview,
free-form for cross-cutting queries).

## Web UI

Single-page chat (`web/index.html`) with no build step or external JS dependencies.

- **Nav**: ubuntu.com/desktop style — dark charcoal bar, orange Ubuntu CoF badge, "Argus" wordmark
- **Chat panel**: full-width, SSE feed messages and query replies share the same thread
- **Markdown renderer**: inline `md()` function handles tables, bold, italic, inline code
- **SSE reconnect**: automatic 3-second retry on connection loss; status dot in header
- **Greeting**: rendered immediately on load before any SSE connection

Feed messages (from `PushToFeed`) and query replies (from `POST /query`) both
render through the same `md()` pipeline, so status tables, change reports, and
LLM answers all use consistent formatting.

## Key design decisions

**Snapshot as query source** — `QueryWorkflow` reads the local snapshot rather
than hitting the Test Observer API on every query. This eliminates a redundant
API call per query, keeps latency low, and gives the LLM consistent data that
matches what the change-watch cycle is monitoring.

**Single LLM call per query** — the full artefact table + question go to the
model in one shot. This is simpler, faster, and more capable than the previous
multi-step intent-detection approach (FuzzyMatch → branch → ComposeReply /
ComposeListReply), which struggled with cross-cutting questions and list queries.

**No database** — `state/snapshot.json` written atomically (write to `.tmp`,
rename) is sufficient for the monitoring use case and avoids infrastructure
complexity.

**Interface-driven testing** — `ArtefactClient` and `LLMClient` are interfaces
with mock implementations, allowing unit tests for all LLM-driven activities
without real API calls.
