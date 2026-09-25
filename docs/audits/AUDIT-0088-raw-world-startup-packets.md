# Audit AUDIT-0088 - Preserve raw world-start startup packets

- Commit subject: `Preserve raw world-start startup packets`
- Change type: `runtime`
- Runtime impact: `yes, bounded pre-spawn startup forwarding`

## Intent

Preserve the wire representation of the upstream `JigsawStructureData`, `StartGame`, and `VoxelShapes` packets across the terminating bridge when both connections use the same negotiated protocol. The prior raw retention fix covered `DimensionData` and `ItemRegistry`, but a capture probe showed that other startup packet serializers could also change bytes.

## Scope

- Retain copied raw world-start packet bytes and the negotiated protocol ID in `minecraft.GameData`.
- Carry the bytes through login and deferred-packet handling without changing packet order.
- Validate packet ID and protocol before writing retained bytes from the listener-side startup sequence.
- Keep decoded packet reconstruction as the fallback for absent or incompatible raw data.
- Extend unit and proxy loopback coverage to assert exact bytes for all retained world-start packets.

## Evidence and provenance

- Owner-controlled capture [captures/session-20260921T003441Z](https://github.com/BedrockDebugProxy-Captures/session-20260921T003441Z/releases/tag/capture-archive) was inspected locally. A focused probe found that typed re-encoding changed the bytes of `JigsawStructureData` and `StartGame`, while `VoxelShapes` happened to round-trip identically in that capture.
- The packet definitions and serialization paths are in the vendored gophertunnel protocol implementation. Raw ownership is copied before the login buffer can be released.
- No external service, public server, credential, or account data is required for the regression tests.

## Validation

- Automated checks: focused nested gophertunnel tests and `TestRunnerForwardsBidirectionalPacketsAndCapturesSession` passed after the change.
- Full project quality gate: `tools/quality.ps1` passed, including formatting, viewer checks, tests, build, static analysis, vulnerability scan, tracked-capture checks, and whitespace validation.
- Live validation: `not_observed`.
- Capture or report: the linked private capture archive, local probe, and loopback test provide the evidence; the raw archive remains outside this project's source history.

## Limitations and follow-up

Raw retention is enabled only when the downstream protocol ID matches the retained packet's protocol. It does not make authenticated login chains identical, does not alter post-spawn forwarding, and does not establish compatibility with any public server. Run the required stamped local vanilla-client validation before treating this runtime change as live-validated.
