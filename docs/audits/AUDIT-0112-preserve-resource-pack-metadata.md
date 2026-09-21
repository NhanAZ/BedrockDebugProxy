# AUDIT-0112 - Preserve per-pack resource-pack metadata

- Commit subject: `Preserve per-pack resource-pack metadata`
- Change type: `runtime`
- Runtime impact: `yes, downstream resource-pack offer metadata`

## Intent

Keep the downstream `ResourcePacksInfo` semantically aligned with the upstream offer when the proxy rebuilds entries
from archive-backed packs.

## Scope

Match served archives to retained upstream entries by UUID, version, and compressed size. Copy `ContentIdentity`,
`SubPackName`, `HasScripts`, `AddonPack`, and `RTXEnabled`, and use the advertised content key only when the current
pack does not carry one. Existing archive delivery, URL removal, pack order, and encrypted current-connection key
handling remain unchanged.

## Evidence and provenance

`Conn.handleClientToServerHandshake` previously rebuilt entries from `resource.Pack`, which cannot represent all
wire-level offer metadata. The existing mixed-offer loopback path now supplies non-default values for URL and RakNet
entries and verifies them across both connection boundaries. No external source code was copied.

## Automated validation

- `go test ./internal/proxy -run 'TestRunnerPreservesMixedResourcePackOfferOrder' -count=3`
- Full `tools/quality.ps1 -LowMemory`, formatting, build, vulnerability scan, and `git diff --check` are required
  before commit.

## Live validation status

Pending. The change affects downstream resource-pack metadata only. A stamped exact-revision The Hive baseline and
owner-controlled encrypted and mixed-offer real-client flows remain required by project policy.

## Limitations and follow-up

Metadata is matched only when UUID, version, and compressed size identify a served archive. Upstream entries that
cannot be served are not synthesized, and URL entries remain intentionally replaced by proxy-served chunk delivery.
