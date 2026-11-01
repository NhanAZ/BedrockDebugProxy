# AUDIT-0243 - Update the Bedrock protocol baseline to 2225

- Commit subject: `Update Bedrock protocol to 2225`
- Change type: `protocol`
- Runtime impact: `yes`

## Intent

Update the selected Bedrock codec for the released 1.26.60 and 1.26.61 protocol baseline so the proxy can decode and encode the current protocol 2225 packet set.

## Scope

The change updates the vendored gophertunnel protocol metadata, packet registrations, packet layouts, NetherNet login metadata validation, protocol fixtures, research notes, provenance notices, README status, and the Unreleased changelog. It adds coverage for the new packet registrations and the corrected packet layouts. The local `PlayFabID` login compatibility patch and the pinned go-nethernet version remain unchanged.

## Evidence and provenance

The selected baseline is supported by the official Mojang `bedrock-protocol-docs` releases 1.26.60 and 1.26.61, both reporting protocol 2225. The codec incorporates the reviewed gophertunnel revisions recorded in `docs/research/protocol-1.26.60-1.26.61.md`. Mojang reports no packet changes between 1.26.60 and 1.26.61, so 1.26.60 packet definitions remain the selected codec for the 1.26.61 metadata.

## Validation

- Automated checks: `tools/quality.ps1 -LowMemory`, root and nested gophertunnel tests, `go test -race ./internal/bedrock -count=1`, stamped version check, and `git diff --check` passed before the release-policy follow-up.
- Live validation: `pending`
- Capture or report: `none`

## Limitations and follow-up

The protocol review and automated coverage do not establish compatibility with public servers or a retail client. A clean stamped runtime revision still requires the documented agent-controlled six-server validation matrix before an official release. Do not reuse historical reports from protocol 2193 or another runtime revision.

## Security, licensing, and compatibility

The change preserves the existing third-party provenance records and local compatibility patch. It does not add credentials, captures, resource-pack keys, or decrypted assets to the repository.
