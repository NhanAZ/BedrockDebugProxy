# Bedrock 1.26.62 protocol update

## Decision

Update the selected client game-version metadata from `1.26.61` to `1.26.62` while retaining network protocol `2225` and the reviewed 1.26.60 packet codec. Do not import preview-only 1.26.70 packet definitions.

This is a metadata and evidence update, not a packet-ID or wire-layout update. The decision is based on the official stable release schema, the selected local codec, and the current upstream gophertunnel metadata.

## Evidence table

| Claim | Protocol version | Source and revision | Independent check | Decision | Confidence |
| --- | --- | --- | --- | --- | --- |
| Bedrock 1.26.62 is the latest stable Mojang schema release | `2225` | Mojang [`v1.26.62` release](https://github.com/Mojang/bedrock-protocol-docs/releases/tag/v1.26.62), published 2026-11-05; release commit [`f8d1496b5efc95105bd9ae1f2f2f076a43f71d82`](https://github.com/Mojang/bedrock-protocol-docs/commit/f8d1496b5efc95105bd9ae1f2f2f076a43f71d82) | The repository watcher resolved `v1.26.62` as the latest non-prerelease release and reported protocol `2225` | Set the schema and client metadata baseline to `1.26.62` while retaining protocol `2225` | Confirmed |
| The final 1.26.62 schema uses the stable packet shape represented by the selected codec | `2225` | Official release commit [`f8d1496`](https://github.com/Mojang/bedrock-protocol-docs/commit/f8d1496b5efc95105bd9ae1f2f2f076a43f71d82); the final schema removes preview-only ChatPacket and music packet definitions, removes `ConcurrentMusic`, `Raw Message`, and `Last Output Is Raw Message`, and restores `AddVolumeEntity` | The local codec already has `AddVolumeEntity`, only `Music`/`Sound`, no preview Chat/music packet types, and no removed raw-message fields | Keep the reviewed 1.26.60 packet definitions; do not import preview packet code | Confirmed for the inspected changed schemas |
| An upstream gophertunnel dependency bump is not available as the protocol implementation source for this update | `2225` / `1.26.60` | Current upstream gophertunnel commit [`0e179fc4bb082c13e946c7d81ab1cf9f623d942c`](https://github.com/Sandertv/gophertunnel/commit/0e179fc4bb082c13e946c7d81ab1cf9f623d942c) and its [`minecraft/protocol/info.go`](https://raw.githubusercontent.com/Sandertv/gophertunnel/0e179fc4bb082c13e946c7d81ab1cf9f623d942c/minecraft/protocol/info.go) still advertise protocol `2225` with game version `1.26.60` | The local tree is a reviewed composite with project-specific observation and compatibility patches that must be preserved | Change only the local client metadata and documentation; retain the local composite | Confirmed |
| The protocol ID and capture schema remain unchanged | `2225` | Local `minecraft/protocol/info.go`, `internal/capture/schema.go`, and the prior 1.26.60/1.26.61 review | Focused protocol tests cover the existing packet registrations and binary fixtures; no new wire claim is introduced | Keep protocol `2225` and capture schema `bedrockdebugproxy.capture.v1` | Confirmed |

## Local codec comparison

The official 1.26.62 release commit is based on a preview tree and therefore contains many metadata and developer-note changes. The packet-level changes relevant to the final stable schema were checked against the in-tree codec:

- `AddVolumeEntity` is already implemented and registered in the local packet pool.
- `AudioContentPlaybackType` already contains only `Music` and `Sound`.
- `CommandBlockUpdate` already omits the preview `Last Output Is Raw Message` field.
- `CommandOutputMessage` already omits the preview `Raw Message` field.
- Preview-only `ChatPacket` and music playback packet types are not present in the local codec.

No packet source, packet ID, field order, version gate, or compatibility workaround was changed for this update.

## Validation scope and remaining uncertainty

The local automated suite validates the selected codec and metadata, but it does not establish complete compatibility with every Bedrock server or client category. Live validation remains required for the server categories listed in [`docs/protocol-updates.md`](../protocol-updates.md), including login, resource-pack negotiation, spawn, forwarding, disconnect, capture verification, and `analyze` output.

The paired private BedrockPackArchive checkout was not present in the workspace during this update. Its producer baseline and compatibility check remain pending before this protocol update can be considered complete under the repository workflow.
