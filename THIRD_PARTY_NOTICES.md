# Third-party notices and provenance

## Project code

Except where this file states otherwise, the source in this repository was written independently for BedrockDebugProxy and is licensed `GPL-3.0-or-later` under the root `LICENSE` file.

Reviewing another implementation, protocol definition, or packet capture does not by itself mean its source was copied. Research influences, source revisions, licenses, observations, and unresolved differences are recorded in `docs/research/`.

The legacy signaling portions of `internal/experience/signaling.go` are substantially adapted from [`bedrock-tool/bedrocktool`](https://github.com/bedrock-tool/bedrocktool) at commit `d7788b57acbdd3eb93ac1efdd4f1107b78aea9b0`. The upstream work is copyright its contributors and distributed under GNU GPL version 3. The adapted file is distributed under those terms. BedrockDebugProxy as a combined work may be distributed under GNU GPL version 3, which is permitted by the project's `GPL-3.0-or-later` choice. The JSON-RPC signaling path and integration hardening are project-owner code rather than copied bedrocktool code. Exact public source paths and behavioral evidence are recorded in [`docs/research/experience-routing.md`](docs/research/experience-routing.md).

The resource-pack decryption behavior was cross-checked against MIT, Apache-2.0, LGPL-3.0, and AGPL-3.0 implementations plus official NIST AES validation material. No source from those projects was copied or adapted. Exact revisions, file links, licenses, the resolved header-field discrepancy, and validation boundaries are recorded in [`docs/research/resource-pack-encryption.md`](docs/research/resource-pack-encryption.md).

## Embedded viewer fonts

The local capture viewer embeds the following unmodified font files. They remain under the SIL Open Font License 1.1 and are not relicensed under the project GPL. Their copyright notices and full license text are preserved in [`internal/viewer/assets/fonts/OFL-1.1.txt`](internal/viewer/assets/fonts/OFL-1.1.txt) and in the generated binary release license bundle.

| Font file | Source revision | SHA-256 | Viewer role |
| --- | --- | --- | --- |
| `Monocraft.ttf` | [`IdreesInc/Monocraft` `e498bf70aeb25b4bdcff1e44d878fb2cb4f7c2a9`](https://github.com/IdreesInc/Monocraft/blob/e498bf70aeb25b4bdcff1e44d878fb2cb4f7c2a9/dist/Monocraft-ttf/Monocraft.ttf) | `262e31822e0a75be567b02320bfa092143d871e8a273e529d1c5ab0c04cd215f` | Content, packet data, decoded JSON, and hex evidence |
| `Minecraft-Bold.otf` | [`IdreesInc/Minecraft-Font` `261ac77fbf28796ca09c22eb83ecdfe386c4b838`](https://github.com/IdreesInc/Minecraft-Font/blob/261ac77fbf28796ca09c22eb83ecdfe386c4b838/Minecraft-Bold.otf) | `2a7216d047fd2991aa8a2fcbc002e0782a903d68de3cc248092ffa88aa941b4f` | Embedded compatibility asset, not selected by the current stylesheet |

Both upstream projects describe the fonts as independent fan-made emulations and state that they do not contain assets or font files from Minecraft. BedrockDebugProxy is not affiliated with or endorsed by Mojang or Microsoft. The font names identify the upstream files and do not imply official status.

The viewer stylesheet also checks for locally installed `Mojangles` and `Minecraft Ten` fonts before these embedded fallbacks. Those proprietary files are not present in this repository, executable, or license bundle. A local installation remains subject to the rights holder's terms. The viewer does not read or serve arbitrary font paths.

The current stylesheet uses regular Monocraft as the fallback for every visible text role. The previously embedded Minecraft-Bold asset remains available from its existing protected asset route and license bundle, but this typography configuration does not select it.

## Direct runtime dependencies

The Go module uses the following direct dependencies. Their source remains under their respective licenses and is not relicensed by BedrockDebugProxy.

| Module | Version | License | Role |
| --- | --- | --- | --- |
| [`github.com/coder/websocket`](https://github.com/coder/websocket) | `v1.8.14` | ISC | Minecraft signaling WebSocket transport |
| [`github.com/df-mc/go-nethernet`](https://github.com/df-mc/go-nethernet) | `v1.0.20-0.20260818142457-6fc1eb6f907c` | MIT | NetherNet and WebRTC upstream transport |
| [`github.com/df-mc/go-playfab/v2`](https://github.com/df-mc/go-playfab) | `v2.0.2` | MIT | Minecraft services authentication |
| [`github.com/df-mc/go-xsapi/v2`](https://github.com/df-mc/go-xsapi) | `v2.0.3` | MIT | Xbox Live session used for service authentication |
| [`github.com/go-gl/mathgl`](https://github.com/go-gl/mathgl) | `v1.1.0` | BSD-3-Clause | Vector and matrix types used by the Bedrock protocol model |
| [`github.com/google/uuid`](https://github.com/google/uuid) | `v1.6.0` | BSD-3-Clause | UUID parsing and values |
| [`github.com/sandertv/go-raknet`](https://github.com/Sandertv/go-raknet) | `v1.15.2-0.20260705184311-0d1fd09e2cf6` plus the local GUID, repeated-acceptance recovery, and cancellation-stage diagnostic patches | MIT | RakNet transport |
| [`github.com/sandertv/gophertunnel`](https://github.com/Sandertv/gophertunnel) | Declared `v1.61.0`, with reviewed merged `v1.62.0` 1.26.50 corrections and local patches | MIT | Bedrock sessions, protocol, authentication, and resource packs |
| [`golang.org/x/oauth2`](https://github.com/golang/oauth2) | `v0.36.0` | BSD-3-Clause | Authentication token source API |

The complete direct and transitive module graph and exact checksums are recorded in `go.mod` and `go.sum`. `tools/collect-third-party-licenses.ps1` verifies that every module compiled into the product has root license material and creates the deterministic license bundle used by binary releases. Distributions that include dependency source or compiled dependency code must preserve the license and notice material required by those dependencies.

The in-tree copy under [`third_party/gophertunnel`](third_party/gophertunnel) uses Sandertv gophertunnel `v1.61.0` at commit `283a5a97dfe65da94bcc0b401807f6aefa9e72ee` as its stable base. Reviewed Bedrock 1.26.50 source is merged from the upstream feature branch through `481f3bd138766304a73d7a0412a47a87acec15ed`, together with the required `ClientboundUpdateSoundData` correction from master through `b8bd7357c24fcba3e32f940d1afb1d3b4e43b96d`. The final upstream merge is `v1.62.0` at `7a556a07335b663744b50d38062636ad8283f314`; its `709768d` height-map correction is now represented in the in-tree protocol source, while the `d564b7d` packed `Hand` correction was already present locally. The selected result preserves later stable fixes that are absent from the feature branch and retains the project-specific compatibility and observation changes described in [`third_party/gophertunnel/BEDROCKDEBUGPROXY_PATCH.md`](third_party/gophertunnel/BEDROCKDEBUGPROXY_PATCH.md). The packet codec remains the 1.26.50 composite, while the selected game-version metadata is 1.26.51 because Mojang's 1.26.51 release reports no packet changes. The upstream source and MIT license are retained.
The upstream resource-pack request-window behavior was researched against [gophertunnel PR #509](https://github.com/Sandertv/gophertunnel/pull/509), which is an open MIT-licensed upstream discussion. The project implementation is an independent narrow adaptation in `minecraft/conn.go` and `minecraft/resource_pack_queue.go`; the PR's source was not copied into the vendor tree. The research note records the source state, scope, and unresolved server-specific behavior.

The in-tree copy under [`third_party/go-raknet`](third_party/go-raknet) is based on the selected go-raknet revision above and retains its MIT license. It contains the documented `Dialer.ClientGUID`, `Dialer.clientGUID`, `Conn.ClientGUID`, and connection-state propagation needed to preserve RakNet identity across featured-experience transfers, the documented recovery for repeated `CONNECTION_REQUEST_ACCEPTED` packets during dial setup, and a cancellation-stage diagnostic for the acceptance wait. See [`third_party/go-raknet/BEDROCKDEBUGPROXY_PATCH.md`](third_party/go-raknet/BEDROCKDEBUGPROXY_PATCH.md).

## Development tools

`tools/quality.ps1` invokes pinned releases of golangci-lint and govulncheck through `go run`. They are development tools, not runtime dependencies and are not added to the BedrockDebugProxy module graph.
