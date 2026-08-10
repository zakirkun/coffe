# Coffe

Coffe is a local-first deployment control plane for a single developer. It previews an allowlisted deployment plan, requires explicit approval, executes it over SSH, and exposes operator workflows through a Wails desktop app plus headless CLI, HTTP, and MCP adapters.

## v0.1 Scope

- SQLite local state and deployment history
- Projects with freely named deployment targets
- Hybrid `coffe.yaml` plus local target configuration
- SSH, `git pull --ff-only`, and Docker Compose deployment
- OpenAI-compatible prompt classification
- Mandatory preview and approval before mutation
- Native Wails dashboard, headless CLI, HTTP API, and MCP over one application core
- Durable cancellation, retry, recovery, health checks, and rollback previews
- Docker Swarm stack deployment plus node/service inspection
- Local coding-agent lifecycle for OpenCode, Claude, Codex, Gemini, and Aider
- SQLite/MySQL read-only queries and schema inspection
- Multi-project Git status/diff/branches plus approved pull and push
- Priority-routed OpenAI/Ollama/vLLM/LM Studio providers with model discovery and usage metrics
- Database pagination, explain plans, local query history, and formula-safe CSV export
- Policy controls for Git push, Swarm deploy, agent allowlists, concurrency, and audit retention
- Redacted audit ledger with formula-safe CSV export
- age+scrypt encrypted SQLite backup/import without keyring leakage
- Cross-platform GoReleaser archives plus HTTPS/checksum-verified staged updater
- HTTPS/token-authenticated remote workers with typed task execution and SSH fallback
- SHA-256-pinned JSON-stdio plugins with declared method allowlists
- Cross-platform CI, legacy migration coverage, worker/plugin integration tests, and RC snapshots

Multi-user authentication, arbitrary shell execution, database mutation, and Swarm node mutation remain out of scope.

## Getting Started

Requires Go 1.24.2 or newer. Go's automatic toolchain download can satisfy this when `GOTOOLCHAIN=auto`.

The desktop app also requires Bun 1.3.14+ and the native Wails prerequisites for the host OS. Start it in development mode or create a packaged native build:

```sh
go run github.com/wailsapp/wails/v2/cmd/wails@v2.12.0 dev -dir desktop
go run github.com/wailsapp/wails/v2/cmd/wails@v2.12.0 build -clean -dir desktop
```

The desktop handles project and target registration, deployment approvals, agents, databases, Git, AI providers, policy, workers, plugins, secrets, backups, updates, and audit export. `serve`, `mcp`, and `worker serve` remain headless protocol modes.

Set `COFFE_DB` to an absolute path to launch the desktop against an isolated or non-default workspace database.

Register a repository that contains `coffe.yaml`, then add an SSH target:

```sh
go run ./cmd/coffe project add --name payments --root /absolute/path/to/payments
go run ./cmd/coffe target add \
  --project PROJECT_ID \
  --name office \
  --host server.example.com \
  --user deploy \
  --path /srv/payments
```

Store a hosted-provider key in the OS keyring when needed. Ollama, vLLM, and LM Studio endpoints may not require one. The same operation is available from the desktop Maintenance screen.

```sh
go run ./cmd/coffe secret set openai_api_key
go run ./cmd/coffe ai add --name local --kind ollama --url http://127.0.0.1:11434/v1 --model MODEL_NAME
```

Run the loopback HTTP API and MCP endpoint, or MCP directly over stdio:

```sh
go run ./cmd/coffe serve
go run ./cmd/coffe mcp
```

Register a read-only database from the desktop Databases screen or from the headless CLI:

```sh
go run ./cmd/coffe database add-sqlite --name analytics --path /absolute/path/analytics.db
go run ./cmd/coffe secret set mysql_reader_password
go run ./cmd/coffe database add-mysql \
  --name production-readonly \
  --host db.example.com \
  --user reader \
  --database app \
  --secret mysql_reader_password
```

Use `coffe.swarm.example.yaml` for a Swarm project. Swarm mutations still follow preview, approval, and execute.

Git operations can also be managed explicitly:

```sh
go run ./cmd/coffe git status PROJECT_ID
go run ./cmd/coffe git preview --project PROJECT_ID --kind pull --remote origin --branch main
go run ./cmd/coffe git approve OPERATION_ID
go run ./cmd/coffe git execute OPERATION_ID
```

Register local or hosted OpenAI-compatible providers. Lower priority values are attempted first:

```sh
go run ./cmd/coffe ai add \
  --name local-ollama \
  --kind ollama \
  --url http://127.0.0.1:11434/v1 \
  --model qwen3 \
  --priority 10
go run ./cmd/coffe ai list
go run ./cmd/coffe ai models PROVIDER_ID
```

Inspect or replace the local policy:

```sh
go run ./cmd/coffe policy show
go run ./cmd/coffe policy set \
  --git-push=true \
  --swarm=true \
  --agent-tools=opencode,claude,codex,gemini,aider \
  --max-agents=4 \
  --audit-days=365
go run ./cmd/coffe audit list
go run ./cmd/coffe audit export --out /absolute/path/audit.csv
```

Export or import encrypted workspace state. Import requires explicit confirmation and preserves the previous database beside the restored one:

```sh
go run ./cmd/coffe backup export --out /absolute/path/coffe-backup.age
go run ./cmd/coffe backup import --in /absolute/path/coffe-backup.age --confirm
```

Keyring credentials are intentionally excluded from backup archives and must be registered again on the restored machine.

Check and stage an update without replacing the running executable:

```sh
go run ./cmd/coffe update check --manifest https://example.invalid/manifest.json --current v1.0.0
go run ./cmd/coffe update download \
  --manifest https://example.invalid/manifest.json \
  --current v1.0.0 \
  --out /absolute/path/coffe.new
```

See `docs/RELEASE.md` for cross-platform release generation.

Run a remote worker behind TLS. The token is read from an environment variable and is never stored in worker arguments:

```sh
export COFFE_WORKER_TOKEN='a-long-random-worker-token'
coffe worker serve \
  --addr 0.0.0.0:4444 \
  --tls-cert /etc/coffe/tls.crt \
  --tls-key /etc/coffe/tls.key
```

On the control plane, store the same token in the keyring and map the worker to a target:

```sh
coffe secret set office-worker-token
coffe worker register \
  --target TARGET_ID \
  --name office-worker \
  --url https://worker.example.com:4444 \
  --secret office-worker-token
coffe worker heartbeat WORKER_ID
```

When no enabled worker is registered, Coffe uses the existing SSH executor. A configured worker failure does not retry over SSH, preventing duplicate mutation.

Plugins are operator-trusted executables. See `docs/PLUGIN.md` for the bounded JSON protocol:

```sh
coffe plugin register \
  --name incident-tools \
  --path /absolute/path/plugin \
  --methods incident.lookup,incident.timeline
coffe plugin invoke PLUGIN_ID --method incident.lookup --input '{"id":"INC-1"}'
```

`POST /api/v1/plans/preview` creates a draft. Remote execution remains unavailable until `POST /api/v1/plans/{id}/approve` records explicit approval.

## Development

```sh
go test -race -shuffle=on -count=1 ./...
go build ./cmd/coffe
cd desktop/frontend && bun install --frozen-lockfile && bun test && bun run check && bun run build
go run github.com/wailsapp/wails/v2/cmd/wails@v2.12.0 build -clean -dir desktop
```

See `coffe.example.yaml` for repository-owned deployment configuration.

## Safety Model

- AI can classify only the `deploy` intent.
- Command strings are rendered from five typed steps and never from model output.
- SSH runs with batch mode and strict host-key checking.
- MCP can inspect projects and create previews, but cannot approve or execute them.
- MCP can inspect Swarm and agents and can run only capped read-only database statements.
- SQLite stores no credentials or unredacted subprocess output.
- Agent prompts are passed to the selected process but are never persisted.
- AI routing persists token/latency metrics, never prompt or response content.
- Git uses the fixed `git` binary with parsed remote/branch values; MCP cannot approve or execute.
- CSV export prefixes spreadsheet-formula cells and remains capped by the query reader.
- Update manifests and assets must use HTTPS; downloaded assets must match SHA-256 before staging.
- Policy changes, approvals, agent starts, provider registration, database queries, backups, and update downloads produce redacted audit metadata.
- Worker tasks reuse the deployment allowlist; plugin binaries are verified against pinned SHA-256 before every invocation.
