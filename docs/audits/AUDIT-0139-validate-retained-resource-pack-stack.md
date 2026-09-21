# Audit AUDIT-0139

## Intent

Prevent a terminating listener from sending stale retained `ResourcePackStack` fields when a per-connection callback prepares a different stack, while preserving unknown trailing wire bytes for observation.

## Scope

- Updated vendored gophertunnel listener stack selection in `third_party/gophertunnel/minecraft/conn.go`.
- Added focused nested regression coverage in `third_party/gophertunnel/minecraft/conn_featured_experience_test.go`.
- Updated the resource-pack research note, resource-pack documentation, fix history, and Unreleased changelog entry.

## Evidence and provenance

The listener receives a prepared typed stack through a callback while also retaining an upstream raw stack and its protocol ID. Protocol and source review showed that protocol equality and packet ID validation alone did not establish that the callback result still matched the retained fields. The fix decodes the fields known by the current protocol revision and compares them with the prepared stack. It deliberately permits trailing bytes so newer fields are not discarded. No external source code was copied.

## Automated validation

- `go test ./minecraft -run 'TestResourcePackResponseWritesRetainedStackBytes|TestResourcePackResponseFallsBackWhenRetainedStackDiffers' -count=1 -v` in `third_party/gophertunnel`.
- `git diff --check`.
- Full project quality and build validation is required before commit.

## Live validation

No live server, public service, or external account was contacted. Real-client resource-pack validation remains pending for the stamped binary.

## Known limitations and follow-up

The comparison covers fields understood by the current vendored protocol. A changed or malformed known field uses typed serialization, while unknown trailing bytes are preserved only when the known fields still match. Raw UDP, RakNet, and authenticated public-server behavior remain outside this local change.
