# AUDIT-0028 - Align 1.26.50 sub-chunk height maps

- Commit subject: `fix: align 1.26.50 subchunk height maps`
- Change type: `protocol | runtime | documentation`
- Runtime impact: `yes`

## Intent

Complete the Bedrock 1.26.50 protocol review after Sandertv gophertunnel merged PR #520. The local composite decoded sub-chunk height maps as a flat 272-byte slice, while the released schema represents two-dimensional 16 by 16 signed heights with a length prefix for each row.

## Scope

Change `third_party/gophertunnel/minecraft/protocol/sub_chunk.go` so both optional height maps use the typed `HeightMap` marshaler and reject a row whose varuint32 length is not 16. Add a deterministic encode and decode regression test in `internal/bedrock/protocol_update_test.go`. Update the protocol research, fix history, changelog, dependency provenance, and local patch note.

The raw packet capture model, same-protocol raw forwarding path, transfer hooks in `minecraft/dial.go`, local go-raknet GUID patch, and the existing packed item-use `Hand` correction are intentionally unchanged. The `Hand` correction from upstream `d564b7d` was reviewed and is already covered by the local implementation and test.

## Evidence and provenance

- Mojang [`bedrock-protocol-docs` release `v1.26.50`](https://github.com/Mojang/bedrock-protocol-docs/releases/tag/v1.26.50) resolves to `475bd72ed89036af4eb18426774ef3b953de7603` and declares network protocol 2193. Its [`SubChunkHeightmapData.json`](https://github.com/Mojang/bedrock-protocol-docs/blob/v1.26.50/json/SubChunkHeightmapData.json) defines the 16 by 16 `[z][x]` `int8` arrays.
- Sandertv gophertunnel [`v1.62.0`](https://github.com/Sandertv/gophertunnel/releases/tag/v1.62.0) resolves to `7a556a07335b663744b50d38062636ad8283f314` after PR #520. Commit [`709768d`](https://github.com/Sandertv/gophertunnel/commit/709768d) adds the row-aware `HeightMap` marshaler.
- The local source was compared with all files changed by PR #520. The remaining `minecraft/dial.go` differences are intentional BedrockDebugProxy transfer hooks, and module metadata retains the documented local replace directives.
- The change is a protocol-model correction. Valid wire bytes remain raw-forwardable through the existing same-protocol path; the fix improves typed decoding and row-length validation without mutating forwarded bytes.

## Validation

- Automated checks: `pass - gofmt, tools/quality.ps1, nested gophertunnel go test ./..., focused protocol tests including malformed-row rejection, and git diff --check`
- Live validation: `pending`
- Capture or report: `none`

## Limitations and follow-up

The focused test does not replace interoperability evidence. Build a stamped binary from this commit and run the required The Hive baseline before treating the runtime protocol change as complete. Then repeat the five-server release matrix if a release is planned. Do not infer that all servers emit sub-chunk metadata in the same order from this correction alone.
