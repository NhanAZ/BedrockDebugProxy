# Resource-pack capture

## Current behavior

The gophertunnel `v1.61.0` dialer normally downloads every upstream resource pack accepted by `DownloadResourcePack` over its RakNet chunk path. BedrockDebugProxy adds a narrow per-connection adapter for the HTTP URL form of `ResourcePacksInfo`: when an upstream offer contains a URL, the adapter downloads and validates that archive before gophertunnel handles the offer, then exposes it through gophertunnel's public `ResourcePackCache` hook. This allows URL-delivered packs to complete negotiation while retaining the URL on the reconstructed pack for downstream metadata and capture.

Only URLs advertised by the current upstream `ResourcePacksInfo` are fetched. The URL must have an `http` or `https` scheme and a host, the compressed archive must match the advertised UUID, version, and size, and the advertised size must be no larger than the local 512 MiB safety bound. A failed URL download remains visible through the upstream cache warning and the original raw packet; gophertunnel then follows its normal chunk fallback. The adapter does not retrieve packs offline or from a separate user-supplied source.

The downstream listener marks offered resource packs as required. Minecraft therefore gives the player the normal choice to download and join or leave. This is a downstream proxy policy, not a claim that every upstream server requires its packs. Gophertunnel exposes `ListenConfig.TexturePacksRequired` for the listener but does not expose the negotiated upstream `ResourcePacksInfo.TexturePackRequired` value through the public connection API, so BedrockDebugProxy cannot mirror an optional upstream policy exactly. The limitation remains explicit in the capture manifest.

Each `resource_pack.archive` event records the following evidence.

- Parsed UUID, version, name, description, and manifest plus any retained download URL and content key
- Parsed manifest and broad module flags
- Download mechanism inferred from the reconstructed pack object
- Archive byte length and SHA-256 checksum
- Upstream connection, endpoints, hop, direction, and sequence
- A blob reference to the exact archive bytes before decryption or extraction

The recorder streams the archive to disk and hashes it in one pass. This avoids allocating a second full copy of a large pack. The resulting digest is checked against the checksum computed by gophertunnel from the downloaded archive.

Archives are saved during the session, not by a later `export` command. Each archive event's `blob.path` locates ZIP bytes under `blobs/sha256/`, with a content-addressed `.bin` filename. A background reader also creates named ZIP copies and an index under `artifacts/packs/`. It does not unpack archive entries. `export` only packages the canonical capture into a portable `.bdpcap` container. See [session artifacts](session-artifacts.md).

## Ordering evidence

The current gophertunnel listener handles requested resource packs sequentially. It sends one `ResourcePackDataInfo`, services the chunk requests for that pack, and then advances to the next pack. BedrockDebugProxy uses that public listener path when offering the downloaded upstream packs to the client. It does not add a second transfer scheduler or assume a fixed ordering between independent server implementations.

## Opt-in decryption

The server-supplied content key is preserved because it is required to analyze encrypted pack entries. The archive itself is never overwritten. Decryption is disabled by default and can be requested with `--decrypt-resource-packs`.

Select that flag on the initial `run` command. Supported plaintext archives are derived and stored during the same session, with independent ZIP copies under `artifacts/packs/decrypted/`. They are not unpacked into individual pack files. The flag does not change skin, chunk, entity, or inventory capture. It only enables the resource-pack derivation described below.

For the documented 32-byte-key AES-256-CFB8 format, an enabled run writes the following derived events with the original `resource_pack.archive` sequence as their parent.

- `resource_pack.contents_manifest` retains each decrypted root or subpack `contents.json` as a sensitive JSON blob. Its event metadata contains counts, paths, and identifiers but not per-file keys.
- `resource_pack.decrypted_archive` retains a deterministic ZIP containing copied plaintext files and decrypted declared files. It omits the encrypted `contents.json` files because their plaintext is already preserved separately.
- `resource_pack.decrypt_error` records an unsupported format or failed derivation as a warning. It also adds a capture limitation and does not abort the connection or discard the raw archive.

AES-CFB8 is not authenticated. A derived archive is useful for analysis but is not integrity proof and never replaces the original captured bytes. The supported header, key, subpack, path, size, provenance, and verification boundaries are documented in [`docs/research/resource-pack-encryption.md`](research/resource-pack-encryption.md).

Resource-pack content keys and archives may be private, licensed, or account-scoped. They remain inside the sensitive local capture and must not be committed or published without permission.

The master key comes from the upstream server's `ResourcePacksInfo` for the current connection. BedrockDebugProxy does not derive, search for, guess, brute-force, or recover it. The current CLI has no offline decrypt command. Decryption produces a mode-restricted ZIP in the operating system's temporary directory and persistent plaintext blobs inside the capture, so it must not be described as memory-only. The temporary file is removed on normal success or failure, but an abnormal process or machine termination may leave it behind for local cleanup.

The GPL license for BedrockDebugProxy source does not relicense captured packs or their textures, models, sounds, scripts, or other assets. Rights and redistribution permission remain with the applicable creator, server operator, Microsoft, Mojang, Marketplace partner, or other rights holder. The precise data flow and legal-risk distinctions are documented in [`docs/legal-and-responsible-use.md`](legal-and-responsible-use.md).

## Evidence and validation

The implementation follows the `ResourcePacksInfo`, pack download, `ResourcePackCache`, `FetchResourcePacks`, `Pack.ReadAt`, `Pack.Checksum`, and sequential listener delivery behavior reviewed in Sandertv gophertunnel `v1.61.0` at commit `283a5a97dfe65da94bcc0b401807f6aefa9e72ee`. Its default dial path requests `ResourcePackChunkData` over RakNet even when `ResourcePacksInfo` advertises an HTTP URL. BedrockDebugProxy's adapter prefetches that URL through `resource.ReadURL` and supplies a cache hit before the default path is selected. The URL therefore remains observable both in the raw packet event and on the reconstructed `resource.Pack`; the exact archive bytes remain the canonical evidence.

At the same pinned revision, `minecraft/listener.go` copies `ListenConfig.TexturePacksRequired` into the downstream connection before `FetchResourcePacks` dynamically returns the upstream archives. `minecraft/protocol/packet/resource_packs_info.go` documents and serializes that flag before the pack list. This supports the player prompt policy above. It does not provide evidence for the unavailable upstream value.

A verified The Hive session showed that the upstream initial offer used `TexturePackRequired = true` while its later stack used `TexturePackRequired = false`, and the session still reached spawn. The packet sequences, raw blob hashes, source definitions, interpretation, and limits are recorded in [`docs/research/resource-pack-negotiation.md`](research/resource-pack-negotiation.md).

Synthetic archive tests verify byte equality, content-key retention, metadata, endpoint context, stream sizing, deduplication, encryption vectors, root and subpack decryption, failure retention, derived event ancestry, and final capture integrity. URL-cache unit tests additionally verify URL prefetch, content-key retention, duplicate-offer suppression, malformed packet handling, and advertised metadata mismatch. Live validation with a real client and owner-controlled servers that use no pack, one pack, multiple packs, an advertised HTTP URL, RakNet delivery, and encrypted entries remains required. For an advertised HTTP URL, verify that the raw `ResourcePacksInfo` retains the URL, the capture records the reconstructed archive with `delivery: "http"`, and the upstream flow reaches resource-pack completion without a `ResourcePackChunkData` request for that URL pack.
