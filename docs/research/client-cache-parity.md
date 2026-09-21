# Client cache parity research

## Question

Does the upstream connection receive the same client blob-cache capability that
the downstream Bedrock client advertises?

## Confirmed observation

The owner-controlled capture `captures/session-20260921T122228Z` records a
different `ClientCacheStatus` value on the two connections. The downstream raw
packet with ID 129 at sequence 197 has a one-byte payload of `01`, which means
the vanilla Windows client enabled the blob cache. The upstream raw packet with
ID 129 at sequence 30 has a one-byte payload of `00`, which means the proxy's
upstream connection disabled it. The same downstream `01` and upstream `00`
pattern appears throughout the completed Windows client captures. The capture
also records `downstream_client_cache_enabled: true` and
`upstream_client_cache_enabled: false` in `session.spawned`.

This is a confirmed wire and session-metadata difference. It is not inferred
from packet counts or from a decoder summary.

## Source path

- The vendored gophertunnel `protocol/packet/client_cache_status.go` defines
  `ClientCacheStatus.Enabled` and states that a false value tells the server not
  to use the blob cache.
- `third_party/gophertunnel/minecraft/conn.go` applies the received value in
  `handleClientCacheStatus` for listener connections.
- The same file sends `ClientCacheStatus{Enabled: conn.cacheEnabled}` after
  upstream `PlayStatusLoginSuccess`.
- `third_party/gophertunnel/minecraft/dial.go` initializes a dialled
  connection from `Dialer.EnableClientCache`, whose zero value is false.
- `internal/proxy/proxy.go` does not set `EnableClientCache` in
  `connectUpstream`, so the upstream value is currently the Dialer default.

The pinned dependency source used for this review is gophertunnel commit
`283a5a97dfe65da94bcc0b401807f6aefa9e72ee`. No external source code was copied.

## Handshake boundary

The difference is caused by the current two-connection startup order. The
upstream Dialer sends its cache status after the upstream login success and
before it receives the upstream `ResourcePacksInfo`. The downstream listener
learns the real cache status only after it has already sent downstream
`ResourcePacksInfo`. BedrockDebugProxy currently needs the upstream resource
pack result before it can complete that downstream offer. Therefore the exact
downstream boolean is not available when the upstream Dialer must choose its
status.

Changing the value unconditionally would make clients that report false use a
server path they do not support. Waiting for the downstream value before
starting the upstream exchange would require a new startup design and could
deadlock the resource-pack negotiation. Guessing from device metadata would
also turn an observed capability into an inference.

## Evidence classification

- Confirmed: Windows client captures in this repository send downstream
  `ClientCacheStatus=true` while the current upstream Dialer sends false.
- Confirmed: the current proxy preserves both raw packets and records both
  connection values in the closed capture.
- Inferred: when an upstream server honours the capability, the mismatch can
  change `LevelChunk.CacheEnabled`, blob-cache packets, and later transport
  payload boundaries.
- Unknown: whether a particular server uses the cache capability for any policy
  decision, and whether matching it improves compatibility for all supported
  client platforms.
- Unknown: a safe generic startup design that learns the downstream status
  before upstream resource-pack negotiation without changing valid packet order.

## Current decision

No runtime change is made from this evidence alone. Keeping the upstream value
false avoids claiming a capability that has not been observed for the upstream
connection and avoids sending cache-enabled chunks to a downstream client that
may not support them. The mismatch remains visible in raw evidence and session
metadata instead of being hidden or normalized.

## Follow-up

Use an owner-controlled local server that deliberately sends cache-enabled
`LevelChunk` and `ClientCacheBlobStatus` traffic to test any future startup
redesign. Compare direct and proxied captures for packet order, cache hashes,
miss responses, and spawn completion. Do not infer a public-server policy from
this local capability difference.
