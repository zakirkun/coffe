# Coffe

Go 1.24.2+ local-first deployment control plane.

## Commands
- `go test -race -shuffle=on -count=1 ./...` - run tests
- `go build ./cmd/coffe` - build the binary
- `go run ./cmd/coffe tui` - open the dashboard
- `go run ./cmd/coffe serve` - run HTTP and MCP endpoints
- `go run ./cmd/coffe mcp` - run MCP over stdio

## Architecture
- `internal/domain` owns valid project, target, plan, and event values.
- `internal/service` owns application workflows and approval policy.
- `internal/store/sqlite` owns local persistence.
- `internal/deploy` renders allowlisted remote commands.
- `internal/agentexec` runs allowlisted local coding-agent binaries.
- `internal/dbtool` owns capped read-only SQLite/MySQL inspection.
- `internal/gittool` owns fixed-binary Git inspection and approved pull/push execution.
- `internal/ai` owns OpenAI-compatible clients and priority fallback routing.
- `internal/backup` owns age-encrypted SQLite snapshot portability.
- `internal/update` owns HTTPS manifest checks and checksum-verified staging.
- `internal/worker` owns authenticated remote task protocol and SSH fallback routing.
- `internal/pluginexec` owns hash-pinned JSON-stdio plugin execution.
- `internal/httpapi`, `internal/mcpserver`, and `internal/tui` are adapters.

## Conventions
- AI output never becomes shell input.
- Every remote mutation requires an approved ActionPlan.
- Secrets never enter SQLite or logs.
- Coding-agent prompts never enter SQLite or logs.
- AI prompts never enter SQLite usage records or logs.
- Git pull and push require an approved GitOperation.
- Policy cannot disable mandatory approval for remote or Git mutations.
- Backup archives never contain OS keyring values.
- Worker and plugin APIs never accept arbitrary shell commands.
- Keep Go files below 250 pure LOC.
