# AUDIT-0271 - Skip unsupported NetherNet status pings

- Commit subject: `Skip unsupported NetherNet status pings`
- Change type: `runtime`
- Runtime impact: `yes`

## Intent

An authenticated NetherNet session completed its resource-pack exchange but reported an unsupported RakNet status ping as a transport error. Avoid recording an operation the selected transport does not support.

## Scope

Mark the authenticated NetherNet transport as not supporting status pings and apply that capability when configuring the upstream dialer. Keep HTTPS endpoint checks and best-effort RakNet transfer-route pings unchanged. Add proxy and transport regression coverage. Update the adjacent BedrockPackArchive fixture to accept a root capture without an optional status-ping event while retaining its child-route ping assertion.

## Evidence and provenance

The pinned local `third_party/gophertunnel/minecraft/nethernet.go` implementation returns `not supported` from `NetherNet.PingContext`. The closed OneBlockOnline capture [session-20261120T195739Z](https://github.com/BedrockDebugProxy-Captures/session-20261120T195739Z/releases/tag/capture-archive), produced by `0.6.0` at `84ab313`, completed a three-pack exchange and recorded `transport.ping` with `minecraft: NetherNet.PingContext: not supported`. Capture integrity passed with zero write errors, dropped events, truncation, or decode errors. Its packs-only scope does not prove spawn or gameplay.

## Validation

- Automated checks: `tools/format.ps1`, `tools/quality.ps1`, and the focused proxy and Experience regression tests passed. The paired BedrockPackArchive fixture test passed with synthetic root and transferred-hop captures. Exact stamped-producer verification and six-server validation remain pending for `0.6.1`.
- Live validation: `not_observed` for the corrected revision. The earlier `0.6.0` capture establishes the unsupported ping symptom only.
- Capture or report: private verified capture [session-20261120T195739Z](https://github.com/BedrockDebugProxy-Captures/session-20261120T195739Z/releases/tag/capture-archive).

## Limitations and follow-up

This change suppresses only an unsupported optional status ping on transports that explicitly report no support. It does not alter packet exchange, authentication, pack handling, or RakNet transfer diagnostics. Update the paired BedrockPackArchive compatibility record with the exact producer commit and rerun its fixture against the stamped producer. Build and validate `0.6.1` on OneBlockOnline, then complete the six-server release gate before publication.
