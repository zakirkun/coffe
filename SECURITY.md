# Security

## Trust Boundary

Coffe v0.x is a local single-user control plane. The HTTP server binds to loopback by default and does not provide multi-user authentication.

## Invariants

- AI output never becomes shell input.
- Remote deployment and Git mutations require explicit approval.
- Process binaries and argument shapes are selected from closed allowlists.
- Credentials remain in the operating-system keyring.
- Backup archives contain SQLite state but never keyring values.
- Database tools accept only conservative read-only statements and cap returned rows.
- Update manifests and assets require HTTPS and SHA-256 verification.
- Audit records contain identifiers and outcomes, not prompts, credentials, query results, or raw unredacted process output.
- Remote workers require bearer authentication and TLS outside explicit loopback development mode.
- Plugin executables require absolute paths, pinned SHA-256, declared methods, capped JSON I/O, and direct execution without a shell.

## Backup Restore

Encrypted backups use age passphrase encryption with scrypt. Use a unique passphrase with at least 12 characters. Import requires `--confirm`, validates SQLite integrity, and preserves the previous database as a timestamped backup.

## Reporting

Do not include production credentials, private keys, customer data, prompts, or database contents in a vulnerability report.
