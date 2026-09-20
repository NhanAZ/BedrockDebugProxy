# AUDIT-0071 - Assert downstream client metadata forwarding

- Commit subject: `test: assert upstream client metadata forwarding`
- Change type: `test`
- Runtime impact: `no`

## Intent

Extend the local proxy regression to verify that the upstream connection receives the downstream skin and cape metadata used by the login request.

## Scope

Add assertions for the synthetic integration client's skin and cape identifiers and encoded data in `internal/proxy/proxy_test.go`. No production code, capture format, network behavior, or authentication behavior changes.

## Evidence and provenance

The assertion covers the same `minecraft.Dialer.ClientData` path used by the local identity-parity fix in commit `915d351`. It uses synthetic values already defined by the existing integration fixture and does not read private captures or credentials.

## Validation

- Automated: `gofmt -w internal/proxy/proxy_test.go`.
- Automated: `go test ./internal/proxy -run '^TestRunnerForwardsBidirectionalPacketsAndCapturesSession$' -count=1 -v` passed.
- Live validation: unchanged. The existing local PocketMine loopback evidence remains documented in `docs/research/local-identity-parity.md`; no public server was contacted.

## Limitations and follow-up

This test proves forwarding of selected login metadata through the unauthenticated local path. It does not prove authenticated Xbox identity or cryptographic key parity with a vanilla client.
