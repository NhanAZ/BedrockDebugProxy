# Protocol maintenance review on 2026-10-05

The stable baseline remains Bedrock 1.26.51 and protocol 2193. This review selects focused maintenance fixes rather than replacing either patched transport tree or adopting preview packet definitions.

## Release and source evidence

| Claim | Source and revision | Decision |
| --- | --- | --- |
| The latest non-prerelease Mojang release is still 1.26.51, protocol 2193, with no packet changes from 1.26.50. | [Mojang release](https://github.com/Mojang/bedrock-protocol-docs/releases/tag/v1.26.51), tag commit `8667debe0adad27de6c692c88942c489a1193206`, and [official changelog](https://mojang.github.io/bedrock-protocol-docs/changelog/), checked on 2026-10-05. | Retain the current protocol and game-version metadata. The 1.26.60 releases are prereleases. |
| The release body formats the network protocol number as `*2193*`. | The same verified Mojang release body, returned by the GitHub API. The current watcher reports `protocol unknown` for that body. | Accept balanced Markdown emphasis and inline code around the numeric value. Test issue contents as well as detection. |
| go-nethernet v1.0.24 fixes SDP candidate selection, fingerprint casing, and cancellable SCTP startup. | [Upstream comparison](https://github.com/df-mc/go-nethernet/compare/v1.0.23...v1.0.24), from `6b1260c879e94129e06a47e716966cfdce9049cb` to `90bfcd1be9a4762a2d0c06243f880d195408561f`. Source files are `conn.go`, `dial.go`, and `listener.go`. MIT. | Update the dependency and its required Pion graph. Preserve the project observation wrapper and existing authenticated dial configuration. |
| v1.0.24 accepts an empty identity-provider domain without removing JWT or signed-fingerprint verification. | [`identity.go`](https://github.com/df-mc/go-nethernet/blob/90bfcd1be9a4762a2d0c06243f880d195408561f/identity.go) and upstream change [`e22519d`](https://github.com/df-mc/go-nethernet/commit/e22519df17b3ace63541d84b50aaceea775d000e). The change describes a 26.50 HTTP join compatibility failure. | Treat this as envelope validation, not a trusted issuer check. The project dialer continues to supply its existing nonempty domain. Server-token temporal behavior remains the previously recorded policy in [the JWT note](nethernet-server-identity-jwt.md). Retail HTTP join behavior is not independently verified here. |
| RakNet's first connection-request cache is shared by concurrent dials and was accessed without a lock. | [`0667bace52851f53ca3f1c916689620db017e7a3`](https://github.com/Sandertv/go-raknet/commit/0667bace52851f53ca3f1c916689620db017e7a3), `internal/message/open_connection_request_1.go`. MIT. | Backport only the cache mutex. Retain the selected transport base, client GUID, transfer, repeated-acceptance, and cancellation patches. |
| A request to download zero resource packs must not advance the download queue. | [`e86a90de9b1087c17952d1d123b175a89c214d5d`](https://github.com/Sandertv/gophertunnel/commit/e86a90de9b1087c17952d1d123b175a89c214d5d), `minecraft/conn.go`. MIT. The local path called `nextResourcePackDownload` and returned `no resource packs to download`. | Backport the empty-list guard and verify that it preserves an existing queue and expected packet set. Coordinate the unchanged capture contract with BedrockPackArchive. |
| Fixed-width NBT array counts were multiplied as int32 before allocating their byte buffers. | [`398154f44e601be70335de0d739c78890e4b6a1c`](https://github.com/Sandertv/gophertunnel/commit/398154f44e601be70335de0d739c78890e4b6a1c), `minecraft/nbt/reader.go` and both endian implementations. MIT. | Backport the length and available-byte checks. Cover integer overflow, truncation, empty arrays, streams, and both byte orders. Network-varint NBT encoding is unchanged. |

The upstream repositories and exact changed files were read through the GitHub API or immutable raw-source URLs. The selected changes retain their upstream MIT licenses. Mojang schemas are reference material and are not copied into project source.

## Review boundaries

The gophertunnel master comparison contains unrelated features, performance changes, Realms work, and a deferred-login change followed by a revert. These are not a reason to replace the reviewed composite or its broader local ordering coverage. The local in-tree version declaration continues to identify its original base.

The proposed upstream `ClientBoundDebugRenderer` correction remains unresolved in this review. Mojang's tagged `json/DebugRendererData.json` describes a string action with an underlying uint8 enum and a conditional marker object, while gophertunnel uses action strings plus an explicit optional marker and packed color. Cloudburst's v671 serializer uses a little-endian int32 action. These sources do not establish one matching wire layout for protocol 2193. Preserve raw observations and the current codec until a matching implementation or authorized capture resolves the difference. The compared immutable sources are [Mojang DebugRendererData](https://github.com/Mojang/bedrock-protocol-docs/blob/8667debe0adad27de6c692c88942c489a1193206/json/DebugRendererData.json), [gophertunnel correction](https://github.com/Sandertv/gophertunnel/commit/98c3d1f3438820b478eb5495c6ff9fb99fa758c5), and [Cloudburst serializer](https://github.com/CloudburstMC/Protocol/blob/4128f33429d842698993865bfcab226da94756bb/bedrock-codec/src/main/java/org/cloudburstmc/protocol/bedrock/codec/v671/serializer/ClientboundDebugRendererSerializer_v671.java).

## Validation scope

The clean starting revision `7927dd4f545016e4c91d5e449aa164bae2122bf3` passed `tools/quality.ps1 -LowMemory` with Go 1.26.6. Static analysis reported zero issues. Govulncheck reported no project-reachable vulnerability and six findings in dependency packages or modules not called by the project.

Each selected fix requires a focused deterministic regression check and the canonical quality gate. Nested replacement-module tests must be run explicitly because the root module's `go test ./...` does not traverse them. The BedrockPackArchive compatibility fixture must verify raw empty download-request evidence, archive bytes, both manifests, plaintext contents.json, per-file key count, and index provenance using synthetic data.

No public server, Microsoft account, retail client, or third-party pack is used by the deterministic checks. Automated results do not establish live NetherNet, retail-client, or six-server release compatibility. Those remain separate validation work.

## Completed checks

The canonical formatter and quality gate passed on the reviewed maintenance source. The replacement RakNet module passed its race-enabled suite. The replacement gophertunnel minecraft, NBT, and packet suites passed, and fixed-width NBT tests also passed on 32-bit Windows. The go-nethernet v1.0.24 module suite passed, including local transport, candidate, identity, endpoint, and discovery cases. The watcher passed seven mocked cases and the read-only current-release check.

BedrockPackArchive passed its tests, vet, build, and the synthetic importer fixture with the stamped producer binary. [AUDIT-0203](../audits/AUDIT-0203-paired-protocol-maintenance-validation.md) records both repository revisions and the exact tested binary. The importer and archive contract required no production code change.
