# AUDIT-0200 - Validate fixed-width NBT array lengths

- Commit subject: `fix: validate fixed-width NBT array lengths`
- Change type: `runtime | decoding`
- Runtime impact: `yes`

## Intent and root cause

The fixed-width int32 and int64 NBT array readers multiplied the advertised count using int32 arithmetic. A synthetic count of `0x40000000` for int32 elements produced a zero-length allocation followed by an index panic. Large incomplete inputs could also request allocations before the reader established that the data existed.

## Scope and provenance

Backport MIT upstream revision `398154f44e601be70335de0d739c78890e4b6a1c` into `minecraft/nbt/reader.go` and the two endian implementations. Update the upstream patch description, notices, fix history, and changelog. This is copied upstream source with its existing license retained. The [maintenance review](../research/protocol-maintenance-2026-10-05.md) records the immutable source and selected scope.

The helper checks platform length limits and available buffered bytes, or reads a bounded stream before validating the received length. It returns the existing `BufferOverrunError` on malformed or truncated arrays. Valid values, byte order, offsets, and trailing data remain intact. Network-varint NBT, packet IDs, capture schema, and raw evidence storage are unchanged.

## Validation

- Automated checks: the previous source panicked on the overflow fixture. After the backport, fixed-width array tests passed for both byte orders, buffered and streaming readers, negative counts, int32 and int64 multiplication overflow, truncation, empty arrays, valid signed values, offsets, and trailing bytes. `go test -count=1 ./minecraft/nbt -timeout 30s` also passed with `GOARCH=386`. The focused minecraft, NBT, and packet replacement-module suite passed. `tools/format.ps1`, `tools/quality.ps1 -LowMemory`, and `git diff --check` passed on the reviewed maintenance source. Static analysis reported zero issues and no project-reachable vulnerability was found.
- Live validation: `not_observed`.
- Capture or report: none. Fixtures contain synthetic bytes only.

## Human validation and limitations

Use owner-controlled fixed-width NBT data and a normal authorized Bedrock session that exercises NBT decoding. Valid arrays should retain their values and malformed arrays should produce an explicit decode error while raw packet evidence remains available. Return a closed capture and the first decode error on an unexpected rejection. This change does not impose a new NBT size limit or claim to cover every malformed NBT type.
