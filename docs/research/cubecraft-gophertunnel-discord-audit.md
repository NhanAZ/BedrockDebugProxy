# CubeCraft and gophertunnel Discord audit

## Scope and method

This audit records the evidence visible in the Bedrock Gophers Discord server on 2026-09-13. The search query was `CubeCraft` in the open `gophertunnel` channel view. Discord reported 159 results, not 195, across seven pages with 25, 25, 25, 25, 25, 25, and 9 results. All seven result pages were inspected. The 195 count may come from an earlier search index, a different query, or a different time window. An attempted history query displayed as `in:gophertunnel CubeCraft` returned no results, so it was not treated as authoritative. The report uses the global `CubeCraft` result set and filters the cards grouped under `gophertunnel` or containing directly relevant technical evidence.

The search spans the Bedrock Gophers server, with result cards grouped by channel. Most matches are unrelated conversation. The entries below are the technically relevant matches. The Discord message links use the gophertunnel channel ID and the message IDs exposed by the search result DOM.

## Executive summary

The Discord history shows a long-running interoperability gap between gophertunnel-based clients or proxies and CubeCraft. It is not a new issue caused only by the current BedrockDebugProxy run.

The strongest evidence is:

- In August 2022, a user reported that they could not join Galaxite and CubeCraft through gophertunnel and explicitly suspected that gophertunnel differed from the normal client.
- In February 2022, gophertunnel issue #122 documented CubeCraft featured-server resource packs not loading through the default proxy. The maintainer response described the behavior as intentional listener/dialer behavior and gave a workaround.
- In July 2023, a CubeCraft connection produced malformed-varint and compressed-batch errors in gophertunnel. The same thread said the path worked without texture packs.
- In April 2024, a community member explicitly suspected that CubeCraft might detect proxies, while also stating that this was not confirmed.
- In December 2024, issue #273 documented invisible blocks on Hive and CubeCraft when using gophertunnel. The issue remains open and describes a decoder/re-encoder workaround.
- In September 2026, the current user reported a seven-day CubeCraft Sentinel ban after using a read-only packet-debugging proxy. Jack replied that nobody in the channel had managed to bypass Sentinel. This is a community observation, not a maintainer explanation of Sentinel's implementation.

These records support the conclusion that CubeCraft is a special and historically difficult target for gophertunnel. They do not prove that CubeCraft's Sentinel detected one specific field, nor that gophertunnel alone caused the ban.

## Directly relevant Discord evidence

| Date | Channel and author | Evidence | Interpretation |
|---|---|---|---|
| 2026-09-13 | `gophertunnel`, NhanAZ | [Question and ban report](https://discord.com/channels/623638955262345216/637335508166377513/1548416562908889170). The proxy is described as read-only and the ban reason is `Unfair Advantage (Sentinel Automatic Cheat Detection)`. | Direct report of the current incident. It establishes intent and outcome, but not the server-side trigger. |
| 2026-09-13 | `gophertunnel`, jack | ["good luck bypassing sentinel"](https://discord.com/channels/623638955262345216/637335508166377513/1548419435860066486) and ["nobody here has managed to do it"](https://discord.com/channels/623638955262345216/637335508166377513/1548419458777882656). | Indicates that the channel has no known public Sentinel compatibility solution. It is not an official technical statement and uses bypass terminology. |
| 2026-09-13 | `gophertunnel`, Tim osman and jack | [Follow-up joking about a CubeCraft bypass](https://discord.com/channels/623638955262345216/637335508166377513/1548420561506533537), followed by Jack's ["good luck you'll need it"](https://discord.com/channels/623638955262345216/637335508166377513/1548420615751340175), ["the chipotle method doesn't count btw"](https://discord.com/channels/623638955262345216/637335508166377513/1548420808492187830), and ["if you do that you're automatically a skid"](https://discord.com/channels/623638955262345216/637335508166377513/1548420812342558764). | No technical method or reproducible fix is provided. The unexplained "chipotle method" should not be treated as evidence or investigated as a bypass recipe. |
| 2026-05-11 | `gophertunnel`, Waleed | [CubeCraft bot issue](https://discord.com/channels/623638955262345216/637335508166377513/1503196000041439374). A logged-in bot reaches `DoSpawn()` but appears about 0.5 blocks above the ground and movement packets are ignored. | Shows that successful login and spawn do not guarantee normal gameplay semantics on CubeCraft. |
| 2026-05-09 | `development`, Waleed | [CubeCraft dial timeout](https://discord.com/channels/623638955262345216/637335508166377513/1502612906359918612). | Another report of a CubeCraft-specific connection failure through the library. |
| 2026-03-17 | `gophertunnel`, GitHub App | [PR #406 test report](https://discord.com/channels/623638955262345216/637335508166377513/1483433120245223575). Full resource-pack download was tested against CubeCraft with five packs via CDN and reported `ok`. | Current gophertunnel can complete at least one CubeCraft pack flow. This does not test proxy identity, anti-cheat, or gameplay fidelity. |
| 2025-05-09 | `gophertunnel`, Bex | [PS4/PS5 and title ID discussion](https://discord.com/channels/623638955262345216/637335508166377513/1370096525220515890). CubeCraft worked for affected users while Galaxite did not; the post quotes gophertunnel's Xbox title-ID validation. | Confirms that server and platform authentication behavior differs. It does not identify a CubeCraft anti-cheat signal. |
| 2025-05-09 | `gophertunnel`, Seb | [16 MB decompressed-limit discussion](https://discord.com/channels/623638955262345216/637335508166377513/1370429478681051256). The post says CubeCraft and WDPE use a 16 MB decompressed limit. | A server-specific payload-size constraint that can affect a proxy or decoder. |
| 2025-03-02 | `dragonfly`, Doge | [Claim that CubeCraft uses TCP downstream](https://discord.com/channels/623638955262345216/637335508166377513/1345446860856496291). | Potential transport clue, but the surrounding context is unavailable and the claim was not independently verified in the Discord result. Treat as unconfirmed. |
| 2024-12-02 | `gophertunnel`, GitHub App | [Issue #273 announcement](https://discord.com/channels/623638955262345216/637335508166377513/1313123637800472697). Invisible blocks were reported on Hive and CubeCraft; re-decoding and re-encoding `LevelChunk` was shown to make the problem disappear. | Concrete evidence of server-specific chunk representation or codec behavior. See the linked upstream issue below. |
| 2024-04-05 | `gophertunnel`, Seb | [Proxy-detection suspicion](https://discord.com/channels/623638955262345216/637335508166377513/1225572100631298078). The exact wording is that CubeCraft "might have something that detects proxies" and that this is "not 100% sure". | The oldest direct proxy-detection hypothesis found in the search. It is explicitly speculation, not proof. |
| 2023-11-24 | `gophertunnel`, Tim osman | ["Cubecraft still kicks tho"](https://discord.com/channels/623638955262345216/637335508166377513/1177342674915840115). | Direct historical evidence that joining or staying connected could fail even after other protocol work. Context is incomplete. |
| 2023-07-28 | `gophertunnel`, sebn't | [CubeCraft malformed packet and compressed-batch errors](https://discord.com/channels/623638955262345216/637335508166377513/1134225962150662255). The post records an invalid NBT string and `number of packets 1210 in compressed batch exceeds 812`. | Concrete protocol or decoder incompatibility. It is unrelated to anti-cheat by itself but demonstrates that CubeCraft exposes unusual traffic to gophertunnel. |
| 2023-07-28 | `gophertunnel`, sebn't | ["if you play on cubecraft without any texture packs it works"](https://discord.com/channels/623638955262345216/637335508166377513/1134510986657808626). | Resource-pack negotiation or transfer was a known compatibility boundary. |
| 2023-06-18 | `gophertunnel`, QMoon | [Claim that CubeCraft uses different encryption](https://discord.com/channels/623638955262345216/637335508166377513/1119976654295539813). | Unverified community claim. It should not be treated as evidence of a custom cryptographic protocol without a capture or source. |
| 2022-11-27 | `gophertunnel`, alvin0319 | [Question about connecting to CubeCraft](https://discord.com/channels/623638955262345216/637335508166377513/1046459969773514782). | Historical connection problem, with no visible answer in the result card. |
| 2022-10-06 | `gophertunnel`, Bex | [Featured-server timeout report](https://discord.com/channels/623638955262345216/637335508166377513/1027290115061600347). CubeCraft and Hive timed out while NetherGames worked. | Confirms that featured servers were already a separate compatibility class. |
| 2022-08-09 | `off-topic`, arfush | [Explicit client-difference question](https://discord.com/channels/623638955262345216/637335508166377513/1006548727772745768). The user could not join Galaxite and CubeCraft through gophertunnel and asked whether gophertunnel differed from the client. | Directly matches the current investigation and shows that the client-versus-library difference was noticed years ago. |
| 2022-02-12 to 2022-02-13 | `gophertunnel`, GitHub App | [Issue #122 report](https://discord.com/channels/623638955262345216/637335508166377513/942064223394099211) and [maintainer response](https://discord.com/channels/623638955262345216/637335508166377513/942160805359665182). The default proxy connected to `play.cubecraft.net:19132` but custom models did not load. The response said this was intentional listener/dialer behavior and recommended dialing first, then passing `conn.ResourcePacks()` into the listener. | A confirmed and documented resource-pack difference, with a supported workaround. |
| 2020-10-08 | `go-raknet`, GitHub App | [CubeCraft ACK observation](https://discord.com/channels/623638955262345216/637335508166377513/763464987766423584). Some servers, including CubeCraft, reportedly never send a single ACK while the normal client still works. | A possible RakNet transport quirk. Treat the wording as a remembered observation, not a complete wire-level proof. |

## Upstream issue correlation

The Discord search also surfaced upstream gophertunnel issues that explain why a proxy can differ from a real client even when it does not intentionally edit gameplay packets.

### Issue #122 - resource-pack handling

The [upstream issue](https://github.com/Sandertv/gophertunnel/issues/122) was opened on February 12, 2022. The report used the default proxy against `play.cubecraft.net:19132`; the client entered the game without the custom models. The issue is closed. The maintainer response quoted in Discord says that listeners and dialers intentionally handle packs separately and suggests dialing upstream first, then assigning `conn.ResourcePacks()` to the listener.

This is not an anti-cheat issue, but it proves that a gophertunnel relay can alter observable resource-pack behavior unless the bridge deliberately mirrors the upstream pack set. BedrockDebugProxy already has an explicit upstream-first pack path, but the raw `ResourcePacksInfo` and CDN/protocol differences still need to remain visible in captures.

### Issue #214 - login order

The [upstream issue](https://github.com/Sandertv/gophertunnel/issues/214) was opened on November 24, 2023. It documents that Hive omitted `PlayStatus` and sent `StartGame` immediately, while CubeCraft's observed sequence included `ServerToClientHandshake`, `PlayStatus`, `ResourcePacksInfo`, `ResourcePackStack`, and `StartGame`. The issue body says the Hive behavior no longer occurred as of December 20, 2024, but it remains open as a guard against future ordering differences.

The relevance is architectural: featured servers do not necessarily share one login sequence. A proxy that terminates two sessions must tolerate and preserve those differences instead of assuming the vanilla order.

### Issue #273 - chunk decoding

The [upstream issue](https://github.com/Sandertv/gophertunnel/issues/273) was opened on December 2, 2024 and remains open. It reports invisible blocks on Hive and CubeCraft and shows that decoding and re-encoding `LevelChunk` data can make them visible. This is evidence of a real gophertunnel representation problem on featured servers, not evidence of a cheat signal.

### Issue #407 - authenticated MITM identity binding

The [upstream issue](https://github.com/Sandertv/gophertunnel/issues/407) is a separate but highly relevant class of problem. It documents BDS 1.26.10.4, protocol 944, where the OIDC multiplayer token's `cpk` is cryptographically bound to the client's key pair. The report concludes that a MITM proxy cannot simultaneously preserve the player's identity and use its own encryption key, and lists token rebinding or a different handshake architecture as possible directions. The issue is closed and does not document a general identity-preserving solution.

Issue #407 is not proof of the CubeCraft Sentinel trigger. Its environment is BDS 1.26.10.4 with `online-mode=false`, while CubeCraft is a featured server with its own server-side implementation and a different protocol generation. It does, however, validate the general concern that authenticated proxying can produce a different identity/key/auth-chain presentation even when packet forwarding is observational.

### PR #406 - deferred-packet ordering

The [upstream pull request](https://github.com/Sandertv/gophertunnel/pull/406) addresses a handshake deadlock when packets arrive before they are added to the expected set. The Discord report says the proposed change was tested with CubeCraft's five CDN packs and reported `ok`. This is a handshake-ordering fix, not an anti-cheat or identity fix.

## What the evidence says about CubeCraft detection

### Confirmed

1. CubeCraft has repeatedly exposed server-specific login, resource-pack, transport, packet-size, and chunk behavior to gophertunnel users.
2. Gophertunnel-based connections can therefore differ from an official client without a project intentionally modifying gameplay packets.
3. The current user was actually disconnected and temporarily banned by CubeCraft with the stated Sentinel reason.
4. No public Discord participant identified a working Sentinel-compatible proxy path. The only direct responses were the two remarks from Jack, not a maintainer diagnosis.

### Plausible but unproven

1. CubeCraft may use proxy detection or may classify a proxy's resulting behavior as an unfair advantage. The April 2024 message is explicitly uncertain.
2. Authenticated upstream identity, key, token, or chain differences may be visible to server-side validation. Issue #407 demonstrates this mechanism on a different server environment.
3. Two independent connections, packet decode/re-encode, batching, compression, timing, and transport behavior may produce an anti-cheat-visible profile even when semantic packet intent is unchanged.

### Not established by this audit

1. The exact Sentinel rule or packet that triggered the ban.
2. Whether CubeCraft detects the proxy by transport, authentication, packet timing, packet shape, identity metadata, or a combination.
3. Whether the ban would occur with every gophertunnel version or every proxy architecture.
4. Whether a code-only change can make a proxy acceptable under CubeCraft's rules.

## Implications for BedrockDebugProxy

The project is not a simple transparent relay. It uses gophertunnel as the Bedrock session and protocol adapter, creates a downstream listener and a separate upstream dial, and forwards decoded packets through `ReadPacket()` and `WritePacket()`. Its own capture layer records raw and decoded evidence, but the live wire path still has two session handshakes and gophertunnel-managed authentication, encryption, batching, compression, and packet serialization.

Therefore, "read-only" accurately describes the project's intended mutation policy, but it does not mean "indistinguishable from the official client." A server evaluates the resulting connection and authenticated session, not the proxy's source-level intent.

The Discord record supports keeping CubeCraft as an important research target. It does not support treating CubeCraft as a safe live validation target while the account is punished or while proxy use is not authorized by CubeCraft. Do not use another account to bypass the punishment.

## Recommended maintainer questions

Ask gophertunnel maintainers about compatibility and supported identity semantics, not about bypassing Sentinel:

1. Is there a supported identity-preserving authenticated MITM mode for current gophertunnel?
2. Does the upstream dial path intentionally generate a new key, auth chain, or client identity, and can those be retained from the downstream client without violating the protocol?
3. Were the CubeCraft proxy-kick reports from 2022-2024 ever resolved, or are they still known limitations?
4. Which current gophertunnel revision contains the resource-pack and deferred-packet fixes, and what CubeCraft flows have actually been tested?
5. Can maintainers recommend an owner-controlled test server or fixture that reproduces the identity and packet-ordering conditions without involving a public server's anti-cheat?

## Bottom line

The channel history confirms a persistent, multi-year CubeCraft/gophertunnel compatibility problem. It makes the current ban technically plausible as a consequence of the proxy's observable session differences, but it does not identify Sentinel's exact detector or prove a single gophertunnel bug. The next productive step is to resolve identity/authentication and fidelity questions with gophertunnel maintainers and reproduce the differences on authorized infrastructure, while preserving the CubeCraft capture as evidence rather than running another live test during the punishment.

The later Oomph source review adds a separate clue: Oomph's public configuration explicitly names proxy detections, while its public detector registration does not expose their implementation. See [`oomph-proxy-detection.md`](oomph-proxy-detection.md) for the revision, deployment model, and the comparison with the local CubeCraft capture.
