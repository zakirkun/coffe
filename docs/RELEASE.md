# Release Process

Releases use GoReleaser v2 and are triggered by tags matching `v*`.

## Local Snapshot

```sh
goreleaser release --snapshot --clean
```

Artifacts cover Linux, macOS, and Windows on amd64 and arm64. User packages are `tar.gz` except Windows `zip`. Raw platform binaries are also published for the staged updater.

Pull requests run tests on Linux, macOS, and Windows. Release-candidate tags matching `v*-rc.*` also produce a GoReleaser snapshot artifact before promotion.

## Tagged Release

1. Run `go test -race -shuffle=on -count=1 ./...`.
2. Create and push a semantic tag such as `v1.2.3`.
3. GitHub Actions runs GoReleaser, creates checksums, and publishes packages.
4. The release workflow generates `manifest.json` from `checksums.txt` and uploads it to the same release.

The manifest maps `GOOS-GOARCH` to an HTTPS asset URL and SHA-256 checksum. Coffe stages the verified asset but does not replace the running executable automatically.

Worker execution tests run fully on POSIX CI. Windows tests assert the worker's explicit unsupported response while validating client, registry, and SSH fallback behavior.
