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

- Automated checks: `tools/format.ps1` and `tools/quality.ps1` passed. BedrockPackArchive commit [`902cb09`](https://github.com/NhanAZ-Tools/BedrockPackArchive/commit/902cb0993cc6d573a73814fd753c0e1daa59fca7) updates its synthetic root and transferred-hop fixture for producer `0.6.1`; its exact stamped-producer fixture, full `go test -count=1 ./...`, `go vet ./...`, and clean Windows build passed.
- Live validation: the corrected `0.6.1` OneBlockOnline packs-only session completed and retained three packs. The six-server release matrix later passed at runtime revision `9a30db76d47111344230c167824b9c57a56bad04`; see [AUDIT-0272](AUDIT-0272-full-session-release-gate.md).
- Capture or report: the symptom capture [session-20261120T195739Z](https://github.com/BedrockDebugProxy-Captures/session-20261120T195739Z/releases/tag/capture-archive) and corrected OneBlockOnline capture [session-20261120T210631Z](https://github.com/BedrockDebugProxy-Captures/session-20261120T210631Z/releases/tag/capture-archive) are private, verified archives.

## Limitations and follow-up

This change suppresses only an unsupported optional status ping on transports that explicitly report no support. It does not alter packet exchange, authentication, pack handling, or RakNet transfer diagnostics. The paired BedrockPackArchive compatibility change is recorded by commit `902cb0993cc6d573a73814fd753c0e1daa59fca7`. A fresh 14-Experience pack acquisition remains part of the coordinated release work.
