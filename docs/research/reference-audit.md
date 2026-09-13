# External reference audit

## Method and result

The project-wide audit was performed on 2026-08-30. It checked every external URL in documentation, research notes, notices, comments, and license text. Important research references were checked at both levels below.

1. The repository, revision, and file path still resolve.
2. The file content at that revision supports the claim made by BedrockDebugProxy.

HTTP success alone was not treated as evidence. GitHub repository metadata, commit objects, recursive trees, exact file contents, and stated repository licenses were compared with the surrounding claim. Legal and policy claims were checked against current official text where available.

One stale reference was found. `docs/research/resource-pack-encryption.md` attributed commit `baa6d9a565c58f5f2306c5646fc833704253109f` and `PackDownloader.java` to `CloudburstMC/ProxyPass`. GitHub rejects that ref and path in the CloudburstMC repository. The exact commit and file exist in [`Kas-tle/ProxyPass`](https://github.com/Kas-tle/ProxyPass/blob/baa6d9a565c58f5f2306c5646fc833704253109f/src/main/java/org/cloudburstmc/proxypass/network/bedrock/session/PackDownloader.java), and the file supports the cited AES-CFB8 behavior. The citation was corrected without changing code.

A follow-up audit standardized every bedrocktool citation on the original public `bedrock-tool/bedrocktool` repository at commit `d7788b57acbdd3eb93ac1efdd4f1107b78aea9b0`. The selector, Gatherings, legacy signaling, capture, and session claims were rechecked against their exact files. JSON-RPC signaling is not attributed to that repository because its public revision does not contain the implementation. Independent public evidence for that path is recorded from `lactyy2/pia` instead.

On 2026-09-13, provenance was reviewed again. All remaining bedrocktool URLs in this repository point to the public upstream project. Code that exists in that upstream revision is attributed there. Owner-specific changes are treated as project-owner code or owner-controlled observations and are not cited as a separate external source.

## Protocol and implementation references

| Source | Revision | Evidence rechecked | Result |
| --- | --- | --- | --- |
| Sandertv/gophertunnel | `283a5a97dfe65da94bcc0b401807f6aefa9e72ee`, tag `v1.61.0` | `minecraft/dial.go`, `listener.go`, `packet.go`, `conn.go`, `network.go` | Packet hooks, pre-play visibility, decoded connection state, and network-wrapper boundaries support the architecture and observability claims. |
| Sandertv/go-raknet | `ea813dc668b5a2a2767cc5577bf77869c965f27a` | `conn.go` | Reliable connection and latency behavior support the transport-boundary claim. |
| bedrock-tool/bedrocktool | `d7788b57acbdd3eb93ac1efdd4f1107b78aea9b0` | `utils/connectinfo/connect_info.go`, Gatherings, legacy signaling, capture, and session files | Supports Experience selection, RakNet join discovery, legacy signaling, and broad capture use-case comparisons. It is not cited as JSON-RPC signaling evidence. |
| lactyy2/pia | `91a85ff9f353c01eb1571aae0238c233a2f8a015` | `signaling/messaging/conn.go`, `dial.go`, and `realms/network_protocol.go` | Independent MIT-licensed evidence for JSON-RPC signaling methods, envelopes, endpoint, and protocol constants. No source is reused. |
| thejfkvis/BedrockX | `1f4b30d990116612cdc0a4c4ded0a26c0bc4aa08` | `example/src/client/Instance.js` | Independent MIT-licensed evidence that `DEFAULT` uses the ordinary host and port path while NetherNet variants use a network ID. No source is reused. |
| PrismarineJS/bedrock-protocol | `6011e261c1b028f92d732348dc91339fb12275dd` | `src/rak.js`, `src/relay.js`, and transform files | The files support the independent layered-codec and relay comparison. |
| Kas-tle/ProxyPass | `baa6d9a565c58f5f2306c5646fc833704253109f` | README and `PackDownloader.java` | Feature-surface and resource-pack decryption claims are supported. |
| CloudburstMC/ProxyPass | `b92c4d4f88ee18643df2e5df26a15f8ef4a7db00` | Current tree and README | Supports the basic proxy lineage only. It is not evidence for the Kas-tle pack downloader. |
| EndstoneMC/spyglass | `b4f5f7f879ec0089d353093ee15e6770caf7c1b0` | Network, capture, store, error, and packet-detail files | Supports the capture-hook, decoded-detail, disk-backed index, and visible-error comparison. |
| EndstoneMC/protocol-dumper | `ea87a290a8154d850f41fef8aaffa0bbe7ebfbd4` | `src/main.cpp`, `visitor.cpp`, and `models.h` | Supports runtime schema-extraction claims. No repository-level license was found, so no source is reused. |
| MrSterdy/bedrock-packet-interceptor | `fd933b85c3672b368df6fd674f5842d7507d338a` | Relay, emitter, and event endpoint files | Supports the relay plus interactive event-view comparison. No license was found, so no source is reused. |
| Endermanbugzjfc/PacketLoggerGophertunnel | `e718c1432397bb66fa95c5188ca239617a2c7298` | Repository source and metadata | Supports only the small historical packet-logging comparison. |
| df-mc/dragonfly | `a36ed0edb548298ab482939e1c653e39f9683719`, tag `v0.11.4` | Server and player implementation files | Supports use as a possible complete-server fixture, not as a proxy architecture. |

The resource-pack-specific review additionally confirmed exact immutable files for iteplenky/bedrock-pack-tools, AkmalFairuz/bedrockpack, AllayMC/EncryptMyPack, and Kas-tle/ProxyPass. The official [NIST AES KAT archive](https://csrc.nist.gov/CSRC/media/Projects/Cryptographic-Algorithm-Validation-Program/documents/aes/KAT_AES.zip) was opened and `CFB8GFSbox256.rsp` was checked against the exact key, IV, plaintext `00`, and ciphertext `5c` used by the local test.

## Legal and policy references

The current official [U.S. Copyright Act chapters 1 and 12](https://www.copyright.gov/title17/), [37 C.F.R. 201.40](https://www.copyright.gov/title37/201/37cfr201-40.html), [Minecraft EULA](https://www.minecraft.net/en-us/eula), [Minecraft Usage Guidelines](https://www.minecraft.net/en-us/usage-guidelines), [Microsoft Services Agreement](https://www.microsoft.com/en-us/servicesagreement), [GitHub DMCA Takedown Policy](https://docs.github.com/en/site-policy/content-removal-policies/dmca-takedown-policy), and the Ninth Circuit's official [MDY Industries opinion](https://cdn.ca9.uscourts.gov/datastore/opinions/2011/02/17/09-15932.pdf) remain accessible and support the narrow claims made in `docs/legal-and-responsible-use.md`.

The linked Chamberlain opinion is hosted by Justia because a stable official Federal Circuit copy of that 2004 opinion was not located during this audit. Its text and citation match `381 F.3d 1178`, but the host is secondary. The legal note presents Chamberlain and MDY as differing appellate approaches rather than treating either as universal law.

## Notices and standard license links

The dependency repository links in `THIRD_PARTY_NOTICES.md` resolve to the projects named in `go.mod`, and the listed license identifiers match their current module provenance. The URLs embedded in the unmodified GNU GPL text point to GNU and FSF license material and are not protocol evidence.

Future important references must be rechecked against the claim before addition. A missing historical source should be marked unverified until equivalent evidence is found. It must not be replaced with a nearby URL merely to remove a 404.
