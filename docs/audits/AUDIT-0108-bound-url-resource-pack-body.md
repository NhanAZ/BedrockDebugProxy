# AUDIT-0108 - Bound URL resource-pack response bodies

- Commit subject: `Bound URL resource-pack response bodies`
- Change type: `runtime`
- Runtime impact: `yes, URL resource-pack prefetch failure boundary`

## Intent

Prevent a server-advertised HTTP resource pack from causing unbounded temporary-archive and parse work when its
response body is larger than the advertised compressed size.

## Scope

Add the project-specific vendored `resource.ReadURLLimited` helper and use it for URL-advertised packs with the
advertised size as the limit. The helper rejects an oversized `Content-Length` and also bounds the bytes actually
read when the header is absent or incorrect. Existing URL validation, UUID/version/size matching, cache fallback,
raw packet retention, and downstream URL removal are unchanged.

## Evidence and provenance

The previous `resource.ReadURL` streamed the response directly into the vendored temporary archive and had no
caller-provided limit. `Content-Length` is optional, so metadata validation happened only after the response had
already been copied and parsed. The change is a narrow local patch to the in-tree Sandertv gophertunnel copy. No
external source code was copied.

## Automated validation

- `go test ./internal/proxy -run 'TestURLResourcePackCache' -count=3`
- The new loopback fixture sends a valid archive followed by extra bytes using a flushed, chunked response and
  confirms that the cache returns no pack and an exceeds-limit error.
- Full `tools/quality.ps1 -LowMemory`, formatting, build, vulnerability scan, and `git diff --check` are required
  before commit.

## Live validation status

Pending. The change only affects URL-prefetched resource-pack failure handling. A stamped exact-revision The Hive
baseline and owner-controlled HTTP resource-pack flow remain required by the project validation policy.

## Limitations and follow-up

The bound applies to the compressed HTTP response body. It does not cap decompressed ZIP entry expansion during later
analysis, and it does not change the normal RakNet chunk-delivery path. Real-client URL and mixed-offer validation
remain pending.
