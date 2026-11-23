# Bedrock 1.26.66 protocol metadata update

## Decision

Update the selected client game-version metadata and Mojang schema baseline to `1.26.66`. Keep network protocol `2225`, the reviewed 1.26.60 packet codec, and capture schema `bedrockdebugproxy.capture.v1`. The official `v1.26.66` and `v1.26.62` repository trees are identical, so this update adds no packet definitions or wire changes.

## Evidence table

| Claim | Protocol version | Source and revision | Independent check | Decision | Confidence |
| --- | --- | --- | --- | --- | --- |
| Mojang published stable schema release `v1.26.66` with network protocol `2225` | `2225` | Official [release page](https://github.com/Mojang/bedrock-protocol-docs/releases/tag/v1.26.66), release commit [`d011926b11583430de5d5c49f9c4e53eb7d1fe58`](https://github.com/Mojang/bedrock-protocol-docs/commit/d011926b11583430de5d5c49f9c4e53eb7d1fe58), published 2026-11-18 | The release body identifies the same version and protocol. The project watcher resolved the tag to this commit. | Set the reviewed schema and client metadata baseline to `1.26.66` | Confirmed |
| The `v1.26.66` schemas match the previously reviewed `v1.26.62` schemas | `2225` | Base commit [`f8d1496b5efc95105bd9ae1f2f2f076a43f71d82`](https://github.com/Mojang/bedrock-protocol-docs/commit/f8d1496b5efc95105bd9ae1f2f2f076a43f71d82); target commit [`d011926`](https://github.com/Mojang/bedrock-protocol-docs/commit/d011926b11583430de5d5c49f9c4e53eb7d1fe58). Both commits resolve to Git tree `29ec7b25f8443f38af8f198fceb501d496fdfa83`; the complete recursive tree comparison found zero changed entries. | The earlier [1.26.62 review](protocol-1.26.62.md) compared its final stable packet shapes with the selected local codec. | Keep the reviewed codec and exclude preview-only packet additions | Confirmed for the complete tagged repository tree |
| Protocol `2225` remains implemented by the independent gophertunnel source | `2225` | Sandertv gophertunnel master at [`d7725bfcbdedfcc2b7c2c13f8575585ff878ab04`](https://github.com/Sandertv/gophertunnel/commit/d7725bfcbdedfcc2b7c2c13f8575585ff878ab04), `minecraft/protocol/info.go` | Its metadata still names protocol `2225` and game version `1.26.60`, so it corroborates the protocol ID but does not independently confirm 1.26.66 metadata. | Do not import an upstream dependency or packet-code change | Confirmed for the protocol ID only |
| The schema license terms remain unchanged, and no Mojang source is copied | n/a | The `LICENSE` at the immutable `v1.26.66` tag states that Mojang owns the files and they are subject to the Minecraft EULA | This change copies no schema or developer-note content | Update project metadata and evidence only | Confirmed |
| Protocol and capture contracts do not change | `2225` | Local `minecraft/protocol/info.go`, `internal/capture/schema.go`, and the previous 1.26.62 packet review | The complete Mojang tree comparison has no changed packet schema or documentation files | Keep protocol `2225`, the 1.26.60 codec, and capture schema `bedrockdebugproxy.capture.v1` | Confirmed |

## Scope and limitations

This is a metadata and evidence update. It changes no packet IDs, fields, serialization, version gates, dependencies, or capture data. The shared protocol ID and identical official schema tree do not establish live compatibility with every Bedrock client or server. No live client or public server was used for this review.

The Mojang repository's license states that its files are covered by the Minecraft EULA. No source files from that repository were copied into BedrockDebugProxy.
