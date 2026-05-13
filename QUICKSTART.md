# ARGUS — Local Dev Quickstart

## Prerequisites
- Docker + Docker Compose
- An [OpenRouter](https://openrouter.ai) API key

## Step 1 — Configure environment

```
cp .env.example .env
```

Edit `.env` and replace the placeholder with your real key:

```
OPENROUTER_API_KEY=your-openrouter-api-key-here
```

## Step 2 — Start the stack

```
make up
```

This builds and starts four containers:
- **temporal** — Temporal server with SQLite (`temporalio/auto-setup`)
- **temporal-ui** — Temporal Web UI (`temporalio/ui`)
- **worker** — Temporal worker (runs workflows and activities)
- **server** — Gin HTTP gateway (Web UI + `/query` + `/feed` SSE at :8080)

Allow ~60 seconds on first start — `temporal` runs schema migrations before
the healthcheck turns green, and `worker`/`server` wait for it.

## Step 3 — Open the UI

| URL                       | What                        |
|---------------------------|-----------------------------|
| http://localhost:8080     | ARGUS Web UI                |
| http://localhost:8233     | Temporal dashboard          |

The left feed panel will show **Connecting…** and turn green once the SSE
connection is established.

## Step 4 — Watch logs (optional)

```
docker compose logs -f
```

## Stopping

```
make down
```

The `state/` volume (snapshot.json) is preserved across restarts.

## Manually triggering a workflow

Open the Temporal dashboard at http://localhost:8233, navigate to
**Workflows → Start Workflow**, and fill in:

| Field | Value |
|---|---|
| Task Queue | `argus` |
| Workflow Type | `StatusTableWorkflow` or `ChangeWatchWorkflow` |

This is the easiest way to trigger a feed update without waiting for the
cron timer (6 h for status table, 10 min for change watch).
