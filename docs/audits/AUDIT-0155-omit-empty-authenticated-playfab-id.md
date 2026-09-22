# Audit 0155 - Omit empty authenticated PlayFab ID

## Intent

Match the authenticated upstream ClientData JSON shape when the downstream vanilla client omits `PlayFabId`.

## Scope

- Mark the gophertunnel `ClientData.PlayFabId` JSON field with `omitempty`.
- Add deterministic coverage for omission of an empty field and retention of a non-empty field.
- Update the gophertunnel fork note, local parity research, fix history, and changelog.

## Evidence and provenance

- The preceding local authenticated loopback at revision `970531e7d7132a4e0f889e1ab9ad735a8d4002c1` recorded no selected `PlayFabId` value but PocketMine still warned that the empty JSON property was unexpected.
- The source path is the in-tree gophertunnel login `ClientData` JSON model and the authenticated dial metadata tests.
- This change is limited to JSON representation. It does not alter token claims, signatures, device metadata, packet forwarding, or capture storage.
- No public server, credential, raw capture, or identifier is added to the repository.

## Automated validation

- Nested gophertunnel tests and the root test suite are required before commit.
- The low-memory project quality gate and `git diff --check` are required before commit.

## Live validation status

Pending a fresh owner-controlled real-client loopback against local PocketMine after the stamped binary is rebuilt. The previous run is retained as the motivating evidence and is not claimed as validation of this follow-up.

## Known limitations and follow-up

The authenticated Android metadata boundary remains documented. This local wire-shape correction does not establish compatibility with or enforcement behavior of any public server.
