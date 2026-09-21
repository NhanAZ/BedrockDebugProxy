# Resource-pack requirement negotiation

## Scope

This note records a narrow live observation from The Hive on 2026-08-30 using Minecraft Bedrock 1.26.45, protocol 2169. It exists to prevent a future change from assuming that `ResourcePacksInfo.TexturePackRequired` and `ResourcePackStack.TexturePackRequired` must always have the same value.

The source capture remains local and sensitive. The evidence below records only packet identities, sequence numbers, blob digests, and the first serialized boolean. It contains no archive bytes, content keys, player identity, or authentication material.

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

An owner-controlled PocketMine capture from `session-20260921T003441Z` provided a local regression fixture for this boundary. Its upstream `ResourcePacksInfo` payload began with `0x00`, while the pre-fix downstream payload began with `0x01`. The local server therefore exposed a concrete optional-offer mismatch. This is protocol metadata evidence from a local test server, not evidence about any public server's enforcement behavior.
