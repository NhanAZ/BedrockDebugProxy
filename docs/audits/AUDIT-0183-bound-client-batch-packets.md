# AUDIT-0183 - Bound larger client packet batches

- Commit subject: `Accept larger client packet batches`
- Change type: `bug fix`
- Runtime impact: `yes`

## Intent

Avoid closing a client session when one batch exceeds gophertunnel's default 812-packet safety limit, while preserving a finite packet-count bound.

## Scope

Add a configurable decoder packet-count limit and expose it through `minecraft.ListenConfig`. Keep the default at 812 for existing consumers. Set the BedrockDebugProxy client-facing listener to 16,384 packets per batch. Add focused decoder boundary tests and update the vendored patch record, changelog, fix history, and research note. Do not change the upstream dialer, forwarding bytes, capture schema, or packet ordering.

## Evidence and provenance

The local closed CubeCraft capture [captures/session-20260923T104518Z](https://github.com/BedrockDebugProxy-Captures/session-20260923T104518Z/releases/tag/capture-archive) at revision `bfc19102a312cbe5b0a18c9432c37afcbb446b24` reached spawn before its downstream listener rejected a client-originating batch with `decode batch: number of packets exceeds max=812`. The local rejection is confirmed. The cause of the large client batch is unknown; it is not evidence of a server kick or ban. The decoder and listener limit path were compared with pinned gophertunnel `v1.61.0` commit `283a5a97dfe65da94bcc0b401807f6aefa9e72ee`. See `docs/research/client-batch-packet-limit.md`.

## Validation

- Automated checks: `go test -count=1 ./minecraft/protocol/packet ./minecraft` passed in the nested gophertunnel module. `pwsh -NoProfile -File tools/quality.ps1 -LowMemory` passed, including root tests, build, static analysis with 0 issues, vulnerability scan with 0 reachable vulnerabilities, module and license checks, and `git diff --check`. The scan also reported 4 findings in imported packages and 2 in required modules that the code does not appear to call.
- Live validation: `pending`
- Capture or report: local [captures/session-20260923T104518Z](https://github.com/BedrockDebugProxy-Captures/session-20260923T104518Z/releases/tag/capture-archive); the matching pre-fix CubeCraft report is `validation/local/bfc19102a312cbe5b0a18c9432c37afcbb446b24/cubecraft-final-104518.json`.

## Limitations and follow-up

The exact rejected packet count is not available because the client batch is encrypted at the captured transport boundary. Build a stamped binary from the final revision, complete the required The Hive baseline, and repeat the CubeCraft flow. Require fresh evidence before treating the fix as live-validated or continuing release approval.
