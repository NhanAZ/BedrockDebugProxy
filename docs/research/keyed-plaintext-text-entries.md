# Keyed plaintext text entries

## Symptom and classification

Some Bedrock resource packs contain a `contents.json` entry with a non-empty per-file key even though the corresponding archive entry is already plaintext. The previous decoder treated every non-empty key as authoritative and transformed those bytes with AES-CFB8. A GALAXITE `Galaxite UI` reproduction showed this behavior for `README.txt`: the source archive contained readable text, while the derived archive contained non-printable bytes. BedrockPackArchive then preserved the derived bytes correctly, so the corruption was introduced by the producer before import.

This is a new local resource-pack decryption root cause. It is separate from the existing root `manifest.json` and icon exception in [plaintext-resource-pack-metadata.md](plaintext-resource-pack-metadata.md). The importer was not skipping `.txt` files, and the producer already parsed the manifest's `content` array and counted the README declaration.

## Evidence and decision

The producer's `contents.json` parser maps every `content` entry to its archive path and key. The decryption format uses AES-CFB8 without authentication, so a keyed entry does not contain enough integrity information to prove that the server's declaration matches the bytes. A general fallback that copies every text file would hide genuine encrypted text, so the fix is deliberately bounded:

- probe only small files with known text-like extensions;
- preserve the original bytes only when they are valid UTF-8 with no NUL or disallowed control characters and the decrypted candidate is clearly not plaintext;
- continue normal streaming decryption for binary files, large text-like files, and text entries where both candidates look like plaintext.

The heuristic is an integrity safeguard for the observed unauthenticated format, not a new encryption format or an offline decryption path. The encrypted `textures/root.txt` regression remains covered to ensure a real keyed text entry is still decrypted.

BedrockPackArchive does not decrypt resource packs. Its matched synthetic fixture now includes a plaintext `README.txt` with a declared key and verifies that the importer copies those bytes unchanged, retains the contents manifest, and reports the declaration in the key count without exposing the key value.
