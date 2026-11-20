# AUDIT-0262 - Reuse URL resource-pack cache in Go Golem

- Commit subject: `Reuse URL pack cache in Go Golem`
- Change type: `runtime, resource-pack compatibility`
- Runtime impact: `yes`

## Intent

Make the direct Go Golem client complete server-advertised HTTP resource-pack offers using the same bounded cache already used by proxy and packs-only sessions.

## Scope and root cause

The proxy dial path observed each raw upstream `ResourcePacksInfo` packet before gophertunnel decoded it, fetched advertised URLs with the per-connection cache, and supplied a validated pack through `minecraft.ResourcePackCache`. The direct Golem dial path had neither hook. It therefore fell back to the RakNet chunk path even when the offer supplied an HTTP URL.

The existing cache moved to `internal/urlpackcache` and is now shared by proxy and Golem dials. Packet observation still retains the raw offer, and cache validation, cancellation, download bounds, fallback behavior, and archive metadata are unchanged. No packet is rewritten or suppressed.

The operator documentation now identifies Go `golem network` as the authenticated public protocol client and the PHP Golem mode as the local plugin simulator. The validation and release examples match the direct-client capture workflow.

## BedrockPackArchive pairing

BedrockPackArchive commit [`1ea804d`](https://github.com/NhanAZ-Tools/BedrockPackArchive/commit/1ea804dc671dc54b286569a8cd918623c747056e) adds a synthetic Golem compatibility fixture with a URL offer, an `AllPacksDownloaded` response, and a canonical archive event. The archive importer contract is unchanged, so no production importer code change was needed. Its `go-jose` indirect pin is aligned with the current gophertunnel replacement module so the sibling consumer builds against this producer checkout.

## Validation

- Focused checks: `go test -count=1 ./internal/urlpackcache ./internal/proxy ./internal/golemnet` passes, including a loopback server that advertises an HTTPS pack URL without a chunk archive. Go Golem fetches one pack, responds `AllPacksDownloaded`, and reaches spawn.
- Full producer quality gate: pass with `tools/quality.ps1`, including the complete test suite, build, static analysis, vulnerability scan, tracked artifact checks, and whitespace check.
- BedrockPackArchive commit `1ea804dc671dc54b286569a8cd918623c747056e`: full tests, focused URL fixture, vet, build, and `git diff --check` pass. Its exact stamped producer verifier also passes against the clean producer build for commit `00845025d143752f032c675d46f898e2ebadadc3`.
- Exact stamped public-server validation: pending. No fresh acquisition result is inferred from earlier captures.

## Limitations and follow-up

The cache still depends on the server's advertised URL being reachable and the archive matching the offer's UUID, version, and byte size. A URL fetch failure remains visible and leaves gophertunnel's normal chunk fallback available. The full fresh 14-experience acquisition, child-route history, and snapshot validation remain separate work. CubeCraft is limited to pack completion, lobby spawn, then immediate disconnect.
