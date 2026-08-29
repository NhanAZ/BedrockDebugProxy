# Resource-pack capture

## Current behavior

The gophertunnel `v1.61.0` dialer downloads every upstream resource pack accepted by `DownloadResourcePack`. BedrockDebugProxy records the raw packet payloads involved in negotiation and transfer. After gophertunnel completes each download, the proxy also stores the full reconstructed archive as a content-addressed blob.

Each `resource_pack.archive` event records the following evidence.

- Parsed UUID, version, name, description, and manifest plus any retained download URL and content key
- Parsed manifest and broad module flags
- Download mechanism reported as RakNet or HTTP
- Archive byte length and SHA-256 checksum
- Upstream connection, endpoints, hop, direction, and sequence
- A blob reference to the exact archive bytes before decryption or extraction

The recorder streams the archive to disk and hashes it in one pass. This avoids allocating a second full copy of a large pack. The resulting digest is checked against the checksum computed by gophertunnel from the downloaded archive.

## Ordering evidence

The current gophertunnel listener handles requested resource packs sequentially. It sends one `ResourcePackDataInfo`, services the chunk requests for that pack, and then advances to the next pack. BedrockDebugProxy uses that public listener path when offering the downloaded upstream packs to the client. It does not add a second transfer scheduler or assume a fixed ordering between independent server implementations.

## Encryption status

The server-supplied content key is preserved because it is required to analyze encrypted pack entries. The archive itself is never overwritten. This version does not claim to decrypt or extract a pack.

Resource-pack content keys and archives may be private, licensed, or account-scoped. They remain inside the sensitive local capture and must not be committed or published without permission.

## Evidence and validation

The implementation follows the `ResourcePacksInfo`, pack download, `FetchResourcePacks`, `Pack.ReadAt`, `Pack.Checksum`, and sequential listener delivery behavior reviewed in Sandertv gophertunnel `v1.61.0` at commit `283a5a97dfe65da94bcc0b401807f6aefa9e72ee`.

Synthetic archive tests verify byte equality, content-key retention, metadata, endpoint context, stream sizing, deduplication, and final capture integrity. Live validation with a real client and servers that use no pack, one pack, multiple packs, HTTP delivery, RakNet delivery, and encrypted entries remains required.
