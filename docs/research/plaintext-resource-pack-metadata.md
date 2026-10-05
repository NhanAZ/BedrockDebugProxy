# Plaintext resource-pack metadata

Reviewed on 2026-10-05.

## Evidence and scope

The synthetic regression reproduces an encrypted contents.json that declares keys for plaintext root manifest.json and icons. Applying those declarations to every entry transforms the metadata into unrelated bytes. This is a new decoder root cause, separate from the implicit-directory and manifest-selection fixes.

The in-tree gophertunnel resource reader parses the original pack manifest before the capture decryption step. The [official manifest reference](https://learn.microsoft.com/en-us/minecraft/creator/reference/content/addonsreference/packmanifest?view=minecraft-bedrock-stable) identifies the manifest as the pack identity metadata. It does not specify this encryption exception.

The independent [MCBEPackCrypt cryptoService.ts](https://github.com/EnderRealmMC/MCBEPackCrypt/blob/0902b5c0a194bad96bdc52c7e1632c1ff769ad37/src/backend/services/cryptoService.ts) at revision `0902b5c0a194bad96bdc52c7e1632c1ff769ad37` lists manifest.json, pack_icon.png, and bug_pack_icon.png as excluded metadata and copies them into its decrypted archive. Its encryptor declares null keys for those entries. The reviewed source is GPL-3.0. This is research influence only. No source was copied or adapted.

## Decision and verification

Copy only the selected root manifest and the two icons at that pack root. Keep the original archive, contents.json plaintext, key declarations, and declared-key counts unchanged. Other paths with the same basename retain their normal keyed decryption. The actual report counts metadata as copied files, rather than decrypted files.

The deterministic test failed on the previous decoder and passes with the focused change. It covers both flat and nested pack roots, all three metadata names, and keyed non-metadata files with matching basenames. BedrockPackArchive's matched fixture checks preserved root manifest and icon bytes, the selected archive digest, contents.json, declared-key count, and source sequence. No capture schema change or importer production change is required.

The AES-CFB8 result remains unauthenticated. This fix does not establish general plaintext integrity, decode unknown encryption formats, recover keys, or add an offline decryption interface.
