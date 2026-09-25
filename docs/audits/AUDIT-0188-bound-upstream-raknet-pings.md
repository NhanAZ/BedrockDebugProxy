# Audit 0188 - Bound optional upstream RakNet pings

- Commit subject: `Bound optional upstream RakNet pings`
- Change type: `upstream dial timeout fix`
- Runtime impact: `yes`

## Intent

Prevent a lost optional RakNet address-discovery ping from holding the upstream login callback for the full 15-minute upstream dial deadline.

## Scope

- Give initial upstream status pings a five-second deadline, matching the vendored go-raknet `Ping` convenience method.
- Preserve the existing 750 ms timeout for transfer-hop dials.
- Retain the ping result or error in the capture and continue to the original address after a failed ping, as gophertunnel already does.
- Add focused proxy coverage and document the observable timeout behavior.

## Evidence and provenance

Capture [captures/session-20260924T161139Z](https://github.com/BedrockDebugProxy-Captures/session-20260924T161139Z/releases/tag/capture-archive) at revision `29795bbef1f266c9915d194e2d31410abfc5fbbb` ended after `upstream.dial_start` without an upstream transport connection. It ended before a ping-completion event, so it does not independently identify the pending dial sub-step. Source inspection of `third_party/gophertunnel/minecraft/dial.go` confirms that `PingTimeout == 0` inherits the caller's deadline. `third_party/go-raknet/dial.go` warns that `PingContext` can wait indefinitely without a deadline and gives its `Ping` convenience method a five-second timeout. No external source code was copied.

## Validation

- Focused regression: `go test ./internal/proxy -run '^TestConnectUpstreamBoundsAddressDiscoveryPing$' -count=1` passed for both initial and transfer-hop deadline selection. The fake transport confirms the regular dial continues after the optional ping returns an error.
- Project quality gate: `tools/quality.ps1 -LowMemory` passed formatting, PowerShell and viewer JavaScript syntax, module tidiness and integrity, third-party license inventory, the full Go test suite, build, static analysis (0 issues), vulnerability scan (0 reachable vulnerabilities; 4 findings in imported packages and 2 in required modules not called by this code), tracked capture artifact checks, and whitespace validation.
- Live validation: pending. The interrupted Enchanted capture remains incomplete and cannot validate this change. A fresh stamped exact-revision six-server matrix is required before release.
- CI: pending for the commit containing this record.

## Limitations and follow-up

This bounds the optional status ping only. It does not shorten authentication, RakNet connection establishment, upstream login, or resource-pack exchange deadlines. Ping errors remain captured with their existing severity and classification. If a fresh live session still stalls before an upstream transport opens, inspect the newly bounded ping event and continue diagnosis at the next observed dial boundary.
