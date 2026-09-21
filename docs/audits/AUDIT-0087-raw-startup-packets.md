# Audit AUDIT-0087 - Preserve raw startup packet bytes

- Commit subject: `Preserve raw startup packet bytes`
- Change type: `runtime`
- Runtime impact: `yes, bounded pre-spawn startup forwarding`

## Intent

Preserve the wire representation of upstream `DimensionData` and `ItemRegistry` packets across the terminating bridge when both connections use the same negotiated protocol. The local evidence showed equal decoded item registries with different bytes after typed re-serialization.

## Scope

- Retain copied raw startup packet bytes and their negotiated protocol ID in `minecraft.GameData`.
- Carry the raw packet through the login state machine, including deferred packet handling.
- Write the retained bytes from the listener-side startup sequence after validating packet ID and protocol.
- Keep typed packet reconstruction as the fallback for absent or incompatible raw data.
- Add unit and loopback integration coverage. `StartGame`, Jigsaw, VoxelShapes, resource-pack negotiation, and post-spawn forwarding remain unchanged.

## Evidence and provenance

- Owner-controlled capture `captures/session-20260921T003441Z` contained upstream and downstream `ItemRegistry` payloads with equal size and different hashes. Decoding both yielded 2,076 equal entries. The first difference was inside an NBT compound, consistent with map-order changes during typed serialization.
- The affected packet types are defined by the vendored gophertunnel protocol implementation. Raw ownership is copied before the login buffer can be released.
- No external service or account data is required for the regression tests.

## Validation

- Automated checks: nested `go test ./...`, focused and representative root proxy tests, full `tools/quality.ps1`, and `git diff --check` passed.
- Live validation: `not_observed`.
- Capture or report: local capture path above. The loopback integration test verifies exact downstream `ItemRegistry` bytes.

## Limitations and follow-up

Raw retention is enabled only when the downstream protocol ID matches the retained packet's protocol. Other startup packets may still be typed-reencoded and require separate evidence before changing. A stamped real-client validation remains a separate compatibility checkpoint.
