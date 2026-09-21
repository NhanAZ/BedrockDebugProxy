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

The versioned Mojang protocol reference for game version 1.26.50 and protocol
2193 describes packet 129 as a packet sent once by the client at login to
communicate whether it supports the cache:
<https://mojang.github.io/bedrock-protocol-docs/1.26.50/>
(accessed 2026-09-21). The reference documents a client capability, not a
platform-derived default. No supported source reviewed here establishes that
`DeviceOS` is a safe substitute for the packet value.

## Handshake boundary

The difference is caused by the current two-connection startup order. The
upstream Dialer sends its cache status after the upstream login success and
before it receives the upstream `ResourcePacksInfo`. The downstream listener
learns the real cache status only after it has already sent downstream
`ResourcePacksInfo`. BedrockDebugProxy currently needs the upstream resource
pack result before it can complete that downstream offer. Therefore the exact
downstream boolean is not available when the upstream Dialer must choose its
status.

The owner-controlled PocketMine source shows the same dependency in the local
fixture. `NetworkSession::onServerLoginSuccess()` sends
`PlayStatus(LoginSuccess)` and installs `ResourcePacksPacketHandler`, whose
`setUp()` then sends `ResourcePacksInfo`; the handler can only observe the
client response afterward. This local source review matches the captured
downstream order and does not establish a universal server order.

Changing the value unconditionally would make clients that report false use a
server path they do not support. Guessing from device metadata would also turn
an observed capability into an inference. The bridge now uses a narrow startup
ordering change instead. Its listener sends `PlayStatus(LoginSuccess)`, waits
for the client's `ClientCacheStatus`, and only then runs the existing upstream
and resource-pack callbacks. The upstream Dialer receives that exact boolean
through `EnableClientCache`. Listeners that do not opt into this deferred offer
mode retain their existing ordering.

## Evidence classification

- Confirmed: Windows client captures in this repository send downstream
  `ClientCacheStatus=true` while the current upstream Dialer sends false.
- Confirmed: the current proxy preserves both raw packets and records both
  connection values in the closed capture.
- Confirmed: the Mojang 1.26.50 protocol reference defines packet 129 as a
  client-reported capability and does not provide a platform default.
- Confirmed: the owner-controlled PocketMine login path sends its resource-pack
  offer before it can observe the client's cache status.
- Confirmed: the owner-controlled loopback integration now sends a cache-enabled
  client through the deferred offer path and observes `true` on both the
  downstream and upstream connection metadata.
- Inferred: when an upstream server honours the capability, the mismatch can
  change `LevelChunk.CacheEnabled`, blob-cache packets, and later transport
  payload boundaries.
- Unknown: whether a particular server uses the cache capability for any policy
  decision, and whether matching it improves compatibility for all supported
  client platforms.

## Current decision

For the BedrockDebugProxy listener, defer the dynamic resource-pack offer until
the downstream capability is observed, then pass that boolean to the upstream
Dialer. This preserves the client's declared capability without choosing a
platform default, and it keeps the existing offer and archive-delivery path
unchanged after the observation point. The raw status packets and both
connection values remain visible in the capture.

## Follow-up

Use an owner-controlled local server that deliberately sends cache-enabled
`LevelChunk` and `ClientCacheBlobStatus` traffic to validate the current
forwarding path. Compare direct and proxied captures for packet order, cache
hashes, miss responses, and spawn completion. Do not infer a public-server
policy from this local capability difference.
