# Plugin Protocol

Plugins are local executables explicitly registered by absolute path, SHA-256, and a method allowlist.

## Invocation

Coffe starts the executable directly without a shell and sends one JSON line to standard input:

```json
{"method":"incident.lookup","input":{"id":"INC-1"}}
```

The plugin writes one JSON response to standard output:

```json
{"result":{"status":"open"}}
```

Errors use:

```json
{"error":"record not found"}
```

Input is capped at 64 KiB and output at 1 MiB. Coffe verifies the executable hash before every invocation, applies a timeout, and forwards only a minimal non-secret environment. MCP can list plugin manifests but cannot invoke them.

Register and invoke:

```sh
coffe plugin register --name incident-tools --path /absolute/path/plugin --methods incident.lookup
coffe plugin invoke PLUGIN_ID --method incident.lookup --input '{"id":"INC-1"}'
```
