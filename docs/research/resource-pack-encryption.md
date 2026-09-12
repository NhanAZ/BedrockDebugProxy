# Resource-pack encryption research

## Scope and conclusion

This note covers the encrypted `contents.json` resource-pack format implemented by BedrockDebugProxy. It does not claim that every Bedrock server, Marketplace pack, experiment, or future protocol version uses this format.

The independently implemented decoder accepts the following evidence-backed variant.

- The server provides a 32-byte master content key.
- `contents.json` starts with a 256-byte header. Bytes 0 through 3 are little-endian version zero, bytes 4 through 7 are `fc b9 cf 9b`, bytes 8 through 15 are zero, byte 16 is the content-ID byte length, and the remaining header bytes contain that UTF-8 ID followed by zero padding.
- Bytes after the header are AES-256-CFB8 ciphertext. The IV is the first 16 bytes of the same master key.
- The decrypted JSON has a `content` array. Each non-empty 32-byte entry key decrypts its named file with AES-256-CFB8 and its first 16 bytes as IV. An empty key means that the archive entry is copied unchanged.
- A root `contents.json` and one `contents.json` in each immediate `subpacks/<name>/` directory use the same master key and content ID. Paths in a subpack manifest are relative to that subpack.

No fallback cipher, inferred key conversion, alternate header, or missing-file workaround is attempted. An unsupported variant produces a structured derived error while the original archive remains available.

## Sources reviewed

| Source | Revision and license | Relevant evidence |
| --- | --- | --- |
| [iteplenky/bedrock-pack-tools](https://github.com/iteplenky/bedrock-pack-tools/tree/b6869edcb3d66ab400638f48663a00d77b6334fd) | `b6869edcb3d66ab400638f48663a00d77b6334fd`, MIT | [`decrypt.go`](https://github.com/iteplenky/bedrock-pack-tools/blob/b6869edcb3d66ab400638f48663a00d77b6334fd/decrypt.go), [`encrypt.go`](https://github.com/iteplenky/bedrock-pack-tools/blob/b6869edcb3d66ab400638f48663a00d77b6334fd/encrypt.go), and [`internal/cfb8/cfb8.go`](https://github.com/iteplenky/bedrock-pack-tools/blob/b6869edcb3d66ab400638f48663a00d77b6334fd/internal/cfb8/cfb8.go) corroborate the 256-byte header, 32-byte key, first-16-byte IV, CFB8 mode, per-file keys, and trailing padding handling. |
| [AkmalFairuz/bedrockpack](https://github.com/AkmalFairuz/bedrockpack/tree/5621fdcecf4a554cba1322dad30e76f9d6cae70d) | `5621fdcecf4a554cba1322dad30e76f9d6cae70d`, Apache-2.0 | [`pack/resource_pack.go`](https://github.com/AkmalFairuz/bedrockpack/blob/5621fdcecf4a554cba1322dad30e76f9d6cae70d/pack/resource_pack.go) and [`pack/encryption.go`](https://github.com/AkmalFairuz/bedrockpack/blob/5621fdcecf4a554cba1322dad30e76f9d6cae70d/pack/encryption.go) independently corroborate header removal, per-file keys, and AES-CFB8 mechanics. |
| [AllayMC/EncryptMyPack](https://github.com/AllayMC/EncryptMyPack/tree/54d0a0b25ffdc5dc6a34e215732608f111d326ec) | `54d0a0b25ffdc5dc6a34e215732608f111d326ec`, LGPL-3.0 | [`PackEncryptor.java`](https://github.com/AllayMC/EncryptMyPack/blob/54d0a0b25ffdc5dc6a34e215732608f111d326ec/src/main/java/org/allaymc/encryptmypack/PackEncryptor.java) explicitly writes the content-ID length, builds root and subpack manifests, and uses `AES/CFB8/NoPadding`. |
| [Kas-tle/ProxyPass](https://github.com/Kas-tle/ProxyPass/tree/baa6d9a565c58f5f2306c5646fc833704253109f) | `baa6d9a565c58f5f2306c5646fc833704253109f`, AGPL-3.0 | [`PackDownloader.java`](https://github.com/Kas-tle/ProxyPass/blob/baa6d9a565c58f5f2306c5646fc833704253109f/src/main/java/org/cloudburstmc/proxypass/network/bedrock/session/PackDownloader.java) corroborates `AES/CFB8/NoPadding`, the first 16 key bytes as IV, the 256-byte prefix, and per-file keys. It was used only as behavioral research. This file and revision do not exist in `CloudburstMC/ProxyPass`. |
| [NIST CAVP block-cipher validation](https://csrc.nist.gov/projects/cryptographic-algorithm-validation-program/block-ciphers) | Official validation material | The [`CFB8GFSbox256.rsp` file in NIST's AES KAT archive](https://csrc.nist.gov/CSRC/media/Projects/Cryptographic-Algorithm-Validation-Program/documents/aes/KAT_AES.zip) contains the exact key, IV, plaintext, and ciphertext used by the one-byte local known-answer test. |

One comment in bedrock-pack-tools calls header byte `0x24` a separator. EncryptMyPack writes that position as the content-ID length. The example content ID is a UUID whose length is 36 decimal, equal to hexadecimal `0x24`. BedrockDebugProxy therefore treats byte 16 as a length field and validates the resulting bounds instead of treating it as a fixed separator.

No third-party source was copied or adapted into `internal/resourcepack`. The implementation and tests were written independently from the behavior cross-checked above. In particular, the AGPL and LGPL projects remain research references rather than incorporated source.

The external references and their supporting claims were rechecked on 2026-08-30. The stale CloudburstMC path above was corrected to its actual Kas-tle repository without changing implementation behavior. See [`reference-audit.md`](reference-audit.md).

## Safety and fidelity boundaries

AES-CFB8 provides confidentiality but no authentication. A correct-looking output does not prove that a per-file key was correct. BedrockDebugProxy reports `authenticated: false`, retains decrypted `contents.json` blobs for later inspection, and keeps the encrypted archive as the authoritative observation.

The decoder rejects non-canonical or escaping paths, duplicate ZIP entries, symbolic links, other special file types, unknown header versions, invalid header padding, mismatched content IDs, invalid key lengths, missing declared files, more than 100,000 archive or declared entries, more than 1,024 contents manifests, a single `contents.json` larger than 16 MiB, more than 64 MiB of retained contents plaintext, or more than 1 GiB of derived file payload. The output ZIP is deterministic and uses conservative file modes. These constraints limit unsafe extraction behavior and resource exhaustion. They are implementation limits, not protocol claims.

## Verification status

Automated coverage includes the official one-byte NIST AES-256-CFB8 known-answer vector, an independently generated multi-byte .NET AES vector, wrong-key rejection, path-escape rejection, root and subpack round trips, capture event ancestry, preservation of the exact encrypted archive after a failure, and final capture verification.

Real Minecraft validation is still required. Test an owner-controlled server with no pack, an unencrypted pack, an encrypted pack, encrypted subpacks, multiple sequential packs, RakNet delivery, and an advertised HTTP URL. For the URL case, verify that the adapter retains the raw `ResourcePacksInfo`, records the reconstructed archive and content key, and reaches resource-pack completion without an unnecessary `ResourcePackChunkData` request. A successful synthetic test is not evidence that every live server uses the supported variant.
