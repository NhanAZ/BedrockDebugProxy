# Audit AUDIT-0123 - Forward client cache capability

- Commit subject: `Forward the observed client cache capability`
- Change type: `runtime | protocol | capture`
- Runtime impact: `yes`

## Intent

Remove the confirmed local parity difference where a vanilla-shaped downstream
client reported client blob-cache support but the upstream Dialer sent the
false zero value.

## Scope

- Add an opt-in listener ordering mode for dynamic resource-pack callbacks.
- Consume downstream `ClientCacheStatus` before BedrockDebugProxy starts the
  upstream login and resource-pack exchange.
- Pass that observed boolean to `minecraft.Dialer.EnableClientCache`.
- Preserve the existing offer reconstruction, raw packet capture, stack
  retention, and gameplay forwarding paths.
- Update the client-cache research note, ADR, fix history, and changelog.

## Evidence and provenance

The owner-controlled capture `captures/session-20260921T122228Z` contains
downstream packet ID 129 payload `01` and upstream payload `00`. The vendored
gophertunnel listener and Dialer source show the ordering boundary and the
existing `EnableClientCache` capability. The change does not infer a platform
default or use public-server enforcement behavior as evidence.

## Validation

- Automated checks: focused proxy integration test, all `internal/proxy` tests,
  nested gophertunnel resource-pack tests, full quality gate, and
  `git diff --check`.
- Live validation: `not_observed`.
- Capture or report: the focused integration capture verifies both
  `session.spawned` cache values are `true`.

## Limitations and follow-up

The focused test verifies capability propagation and login ordering, but it
does not exercise `LevelChunk` blob-cache hashes or miss responses. Add an
owner-controlled local server fixture for those packets before claiming full
cache-path parity. No public-server connection is required for this change.
