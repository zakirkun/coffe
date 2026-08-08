# Threat Model

## Assets

- Deployment and Git mutation authority
- Worker bearer tokens and provider/database credentials in the OS keyring
- Local SQLite control-plane state
- Coding-agent prompts and database query results
- Plugin execution authority

## Remote Worker

- Registry URLs require HTTPS.
- Tokens remain in the OS keyring and are compared constant-time by workers.
- Worker requests carry typed plan/step values; arbitrary commands are rejected.
- Commands are rendered through the same deployment allowlist used by SSH execution.
- Worker output is capped and redacted before entering control-plane events.
- A configured worker failure does not fall back to SSH, preventing duplicate mutations. SSH fallback occurs only when no enabled worker is registered.

## Plugins

- Registration is an explicit trust decision.
- Executables require absolute paths and pinned SHA-256 hashes.
- Hashes are verified before every invocation.
- Plugins run directly without a shell, receive bounded JSON, and expose only declared methods.
- Plugin processes receive a minimal environment without Coffe secrets.
- Plugin invocation remains operator-only; MCP exposes manifest listing only.

## Residual Risk

- A trusted plugin executes with the local user's OS permissions and may access resources available to that user.
- Worker bearer tokens are shared secrets; deployment should use short-lived rotation and TLS certificates managed by the operator.
- Coffe v0.x remains single-user and loopback HTTP has no multi-user authentication.
