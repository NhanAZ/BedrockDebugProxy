# Resource-pack requirement negotiation

## Scope

This note records a narrow live observation from The Hive on 2026-08-30 using Minecraft Bedrock 1.26.45, protocol 2169. It exists to prevent a future change from assuming that `ResourcePacksInfo.TexturePackRequired` and `ResourcePackStack.TexturePackRequired` must always have the same value.

The source capture is sensitive, but this note does not record its session identifier or an archive link. The evidence below records only packet identities, sequence numbers, blob digests, and the first serialized boolean. It contains no archive bytes, content keys, player identity, or authentication material.

## Observed evidence

The closed capture passed the project verifier and later reached `session.spawned`. Its upstream server-to-client packet evidence contains the following values.

| Packet | Sequence | Raw blob SHA-256 | First payload byte |
| --- | ---: | --- | ---: |
| `ResourcePacksInfo`, ID 6 | 35 | `87e101e2c3e53bf8a8c1479e5e7e8c33cdf73bb61f12067333d52508a04442ec` | `0x01` |
| `ResourcePackStack`, ID 7 | 967 | `8ea32f62e54350db8d038795336d7432a6ff7da790859b7b748cd1905e17684c` | `0x00` |

At gophertunnel `v1.61.0`, commit `283a5a97dfe65da94bcc0b401807f6aefa9e72ee`, [`minecraft/packet.go`](https://github.com/Sandertv/gophertunnel/blob/283a5a97dfe65da94bcc0b401807f6aefa9e72ee/minecraft/packet.go) passes `PacketFunc` the payload after reading the packet header. Both packet definitions serialize `TexturePackRequired` as their first field. See the immutable [`ResourcePacksInfo`](https://github.com/Sandertv/gophertunnel/blob/283a5a97dfe65da94bcc0b401807f6aefa9e72ee/minecraft/protocol/packet/resource_packs_info.go) and [`ResourcePackStack`](https://github.com/Sandertv/gophertunnel/blob/283a5a97dfe65da94bcc0b401807f6aefa9e72ee/minecraft/protocol/packet/resource_pack_stack.go) definitions.

The observation therefore decodes to `true` in `ResourcePacksInfo` and `false` in `ResourcePackStack`. The session completing spawn shows that this difference is accepted in at least this real server flow. It is not evidence that every server uses the same values or ordering.

## Project impact

BedrockDebugProxy marks the initial downstream `ResourcePacksInfo` offer as required so Minecraft can present the normal download-and-join or leave choice. The pinned gophertunnel listener creates the later downstream `ResourcePackStack` with its default false value. The observed The Hive flow demonstrates that forcing the later stack flag to match the initial offer is not justified by this evidence.

The patched connection API now retains the decoded upstream `ResourcePacksInfo` on dialled `GameData`. BedrockDebugProxy uses that value to mirror offer metadata downstream while rebuilding the pack entries from the archive-backed resources it can serve. The upstream URL list is therefore not forwarded verbatim, and a listener without an upstream offer still uses its configured policy. A real Minecraft client still has to validate the visible prompt on each release candidate.

The rebuilt downstream entries now preserve the per-pack `ContentIdentity`, `SubPackName`, `HasScripts`, `AddonPack`, and
`RTXEnabled` fields by matching each served archive to its advertised UUID, version, and compressed size. The
downstream `DownloadURL` remains controlled by the archive-backed delivery path and is not copied from the upstream
offer. This closes a confirmed metadata-loss boundary without changing archive bytes or pack order.

An owner-controlled PocketMine capture from [session-20260921T003441Z](https://github.com/BedrockDebugProxy-Captures/session-20260921T003441Z/releases/tag/capture-archive) provided a local regression fixture for this boundary. Its upstream `ResourcePacksInfo` payload began with `0x00`, while the pre-fix downstream payload began with `0x01`. The local server therefore exposed a concrete optional-offer mismatch. This is protocol metadata evidence from a local test server, not evidence about any public server's enforcement behavior.

## Mixed URL and RakNet offers

The source review found a separate confirmed boundary in the URL prefetch adapter. `internal/proxy/proxy.go` previously replaced `conn.ResourcePacks()` with `urlResourcePackCache.Packs()` whenever at least one URL pack was prefetched. `Packs()` reports cache completion order rather than the complete `ResourcePacksInfo` offer order. If the offer mixed URL and non-URL entries, or if one entry was not yet present in the cache, replacing the connection list could omit an entry and would discard the original offer order.

The adapter now records the complete advertised cache-key order and merges prefetched URL archives with the normal gophertunnel list. A validated URL archive wins for its key, non-URL packs remain present, and the original offer order is retained. The unit test covers URL, RakNet, URL ordering, and the root proxy loopback test drives the same mixed offer through the listener and verifies the downstream metadata and stack order. The loopback client declines pack downloads so the test isolates negotiation metadata and does not claim archive application or public-server behavior. This is a confirmed source-level defect and regression fix. No live mixed-offer capture has been run yet, so the frequency of this server behavior is unknown and the real-client result remains pending.

## Login cancellation boundary

The listener's dynamic resource-pack callback runs during the login sequence,
before `Listener.Accept` returns. The observing network wrapper therefore cannot
be the only source of the downstream context: its `ConnectionContextFunc` runs
when `Network.wrap` surrounds the accepted connection, after the callback has
already started upstream preparation. The listener connection itself already
owns the transport context, so the proxy uses that context as an early-login
fallback and prefers the observed wrapper context once it is available.

This is confirmed by the gophertunnel listener and network call order. It keeps
URL prefetch, upstream dialing, and resource-pack recording cancellable when a
client disconnects during login, without changing packet order or archive
selection. A local unit test covers both the pre-`Accept` fallback and the
post-`Accept` tracked context. No public-server behavior is inferred.
