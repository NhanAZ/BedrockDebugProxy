# Audit 0101 - Preserve mixed resource-pack offer order

- Commit subject: `Preserve mixed resource-pack offer order`
- Change type: `runtime`
- Runtime impact: `yes`

## Intent

Keep every resource pack from a mixed `ResourcePacksInfo` offer and preserve the order in which the upstream server advertised them.

## Scope

- Record the complete offer key order in `internal/proxy/resource_pack_url.go`.
- Merge the URL-prefetch cache with the normal gophertunnel resource-pack list in `internal/proxy/proxy.go`.
- Add a deterministic mixed URL and RakNet regression test.
- Leave URL validation, archive retention, raw packet capture, and failed-download fallback unchanged.

## Evidence and provenance

The previous `connectUpstream` path replaced the complete `Conn.ResourcePacks()` result with `urlResourcePackCache.Packs()`. The latter reports cache completion order and does not encode the complete advertised offer order, so source inspection confirms an omission and ordering risk when entries are mixed or still incomplete. The new test uses synthetic archives and an offer ordered URL, RakNet, URL. No public server or third-party source was contacted for this change.

## Validation

- Automated checks: focused URL-cache tests pass. Full root and vendored-module tests, `tools/quality.ps1 -LowMemory`, and `git diff --check` are required before commit.
- Live validation: `not_observed`
- Capture or report: `none`

## Limitations and follow-up

The fix is validated against the local merge path only. Run an owner-controlled real-client session with a mixed URL and RakNet offer and inspect the downstream `ResourcePacksInfo`, resource-pack artifacts, stack order, and spawn completion before treating live compatibility as verified.
