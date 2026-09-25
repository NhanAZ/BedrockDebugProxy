# Bedrock 1.26.51 metadata update

This note records the protocol review for updating the selected Bedrock game-version metadata from `1.26.50` to `1.26.51` while retaining protocol `2193` and the reviewed in-tree packet definitions.

## Evidence

| Claim | Evidence | Decision |
| --- | --- | --- |
| Bedrock 1.26.51 uses network protocol 2193 | Mojang [`bedrock-protocol-docs` release `v1.26.51`](https://github.com/Mojang/bedrock-protocol-docs/releases/tag/v1.26.51), resolved to commit [`8667debe0adad27de6c692c88942c489a1193206`](https://github.com/Mojang/bedrock-protocol-docs/tree/8667debe0adad27de6c692c88942c489a1193206). | Keep `CurrentProtocol = 2193`. |
| The 1.26.51 release has no packet schema changes | Mojang's [protocol changelog](https://mojang.github.io/bedrock-protocol-docs/changelog/) lists stable 1.26.51 as protocol 2193 with `0` packet changes. | Keep the reviewed 1.26.50 packet codec and update only the advertised game-version metadata. |
| The old authenticated proxy path emitted a different upstream ClientData game version | The closed local capture [captures/session-20260920T135412Z](https://github.com/BedrockDebugProxy-Captures/session-20260920T135412Z/releases/tag/capture-archive) records downstream `ClientData.GameVersion = 1.26.51` and authenticated upstream `ClientData.GameVersion = 1.26.50`. Both sides recorded protocol 2193 and Android device OS metadata. | Correct the selected metadata version so gophertunnel's authenticated Android login data matches the downstream 1.26.51 client. |
| The upstream code used the selected protocol constant for authenticated Android metadata | The in-tree [`minecraft/dial.go`](../../third_party/gophertunnel/minecraft/dial.go) `setAndroidData` path assigns `protocol.CurrentVersion` to `ClientData.GameVersion`. | Change `minecraft/protocol/info.go` to `CurrentVersion = "1.26.51"`; do not remove authentication or rewrite identity fields. |

## Scope and disagreement

This is a version-metadata correction, not a wholesale gophertunnel 1.26.51 import. The selected source remains the documented gophertunnel 1.26.50 composite because the official 1.26.51 release reports no packet changes. The project therefore does not claim that every upstream source file has been reviewed for a new packet release. The packet ID, field order, serialization, compression, encryption, and forwarding paths are unchanged.

The metadata change is not an anti-cheat bypass and does not make any claim about the cause of the Hive punishment. The affected public-server capture remains sensitive evidence, and no public-server retest is authorized by this project record.

## Validation and limits

- Focused protocol and proxy tests must assert 1.26.51 metadata and continue covering packet IDs and exact packet bodies.
- The project quality gate must pass, including build, static analysis, vulnerability scan, and viewer checks.
- The existing local PocketMine `ProxyFingerprint` direct-versus-proxy comparison remains the runtime evidence for the unauthenticated loopback path. It recorded zero field and packet-count differences before this metadata-only update.
- No authenticated vanilla-client or public-server validation has been performed for this change. A real client and an authorized test server are still required to establish authenticated interoperability.
