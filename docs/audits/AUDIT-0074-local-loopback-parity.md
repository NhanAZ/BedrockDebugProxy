# Audit 0074 - Validate rebuilt local identity parity

## Intent

Record the local-only validation that confirmed the identity-forwarding fix with a binary rebuilt from the current revision and classified the PocketMine parser warnings observed during the run.

## Scope

- Updated `docs/research/local-identity-parity.md`.
- No runtime source, capture schema, or protocol behavior changed.
- Validation used only the operator-controlled local PocketMine/ProxyFingerprint server and a synthetic offline client.

## Evidence and provenance

The rebuilt binary was produced from the current BedrockDebugProxy revision after the stale pre-fix binary had produced a false mismatch. Labeled direct and proxied reports both reached `LocalProbe`, completed login and spawn, and compared with zero field differences and zero packet-count differences. PocketMine's `ClientData` and self-signed JWT model classes were inspected to classify the unknown-property warnings.

## Automated validation

- `git diff --check`
- `go test ./internal/proxy`
- `go build -o bin/bedrock-debug-proxy.exe ./cmd/bedrock-debug-proxy`

## Live validation

No public server was contacted. The local server temporarily used `xbox-auth=off` for the synthetic loopback fixture and was restored to `xbox-auth=on` afterward. The temporary client was removed.

## Known limitations and follow-up

This evidence covers the unauthenticated local path only. It does not characterize any public server's enforcement or replace a real vanilla-client validation of an authenticated path.
