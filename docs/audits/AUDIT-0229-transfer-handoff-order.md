# AUDIT-0229 - Prepare transfer identity before reconnect

## Problem and root cause

Transfer following already preserves the preceding upstream source address and client GUID, but a fast client could reach FetchResourcePacks before that identity was assigned. The proxy published its pending hop after flushing Transfer and populated the transport fields only after runHop returned. The next listener callback could therefore dial with defaults, or race the bounded route probe for the same source socket.

The enhanced loopback test reproduced different nonzero upstream GUIDs with the old implementation. This establishes the local race independently of any public server refusal. Previous transfer socket release and physical flow provenance fixes remain in use. No dependency source or external implementation is copied.

## Focused correction and paired consumer

Publish the complete pending hop before writing the rewritten Transfer. Its existing upstream dial waits on a readiness channel until the old hop is closed and the existing bounded route probe completes. Cancellation uses the linked session and actual downstream connection context. All packet bytes, raw capture evidence, protocol definitions, archive formats, and schema remain unchanged. No server-specific exception or timed reconnect delay is introduced.

The loopback regression covers ordinary client close and overlapping reconnect. It verifies equal nonzero upstream GUIDs, equal local UDP endpoints, a completed probe before the child connection, two world startups, and the late old-client raw packet remaining on downstream-1. The paired private BedrockPackArchive fixture adds the probe and child connection before the offer. It preserves archive digests, manifests, contents.json, three synthetic key declarations, source sequence, hop, and navigation path without importer production changes.

## Validation and limitations

Focused transfer checks passed three repetitions after the correction, and two repetitions with go test -race passed. Canonical tools/format.ps1 and tools/quality.ps1 -LowMemory passed with zero linter issues and no called vulnerability. The paired consumer full tests, vet, build, whitespace checks, and synthetic matrix using the preceding stamped producer verifier passed at 13409fcf6208270ab6661c712e95a0fcec9f465b. The importer contract is unchanged. The actual new stamped producer-verifier matrix and exact-revision live evidence follow this clean commit and are recorded separately before publication. Existing live captures cannot validate this new runtime.

For human validation, join Mineville Zeqa, choose its current BEDWARS NPC, and allow the normal transfer. Expect the same upstream source endpoint and GUID, complete child pack exchange, and child world startup. Repeat the ordinary The Hive baseline. If the server still refuses startup, retain the closed capture and sanitized report with the error. A preserved identity does not itself prove that the server will accept a connection. Retail gameplay and the six-server official release gate remain separate and unobserved by these loopback tests.
