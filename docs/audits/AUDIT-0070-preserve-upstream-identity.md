# AUDIT-0070 - Preserve downstream identity on unauthenticated upstream dials

- Commit subject: `fix: preserve downstream identity on unauthenticated upstream dials`
- Change type: `runtime`
- Runtime impact: `yes`

## Intent

Make an unauthenticated local upstream observe the same identity supplied by the downstream client instead of gophertunnel's default `Steve` identity.

## Scope

Pass `login.IdentityData` from the listener's resource-pack callback through `connectUpstream` into `minecraft.Dialer.IdentityData`. Add an upstream fake-server regression assertion and document the local observation. Authenticated Xbox identity selection, packet forwarding, capture storage, and transfer routing remain unchanged.

## Evidence and provenance

- Before the change, a local Axolotl-PM server with `xbox-auth=off` logged the proxied synthetic client as `Steve` while the direct connection was `LocalProbe`.
- After the change, the same loopback run logged `LocalProbe` for both direct and proxied connections.
- The gophertunnel dialer source at `third_party/gophertunnel/minecraft/dial.go` defaults the identity only when the caller does not supply one and replaces it from the Xbox chain when authentication is enabled.
- Detailed notes are in [`docs/research/local-identity-parity.md`](../research/local-identity-parity.md). No public server, account token, or captured private data was used.

## Validation

- Automated checks: `gofmt`, `go test ./internal/proxy -run '^TestRunnerForwardsBidirectionalPacketsAndCapturesSession$' -count=1 -v`, the full `tools/quality.ps1` gate before this change, and `git diff --check` after the change.
- Live validation: pass on the local PocketMine loopback server using a synthetic offline gophertunnel client. Direct and proxied login, resource-pack exchange, spawn, and report capture completed. Public-server validation: not observed.
- Capture or report: local synthetic reports under the ignored PocketMine plugin data directory. They are not part of this commit.

## Limitations and follow-up

The regression covers the unauthenticated identity path. A real Xbox-authenticated vanilla-client run is still required before claiming full vanilla parity. This change is not an anti-cheat bypass and does not establish how any public server detects intermediaries.
