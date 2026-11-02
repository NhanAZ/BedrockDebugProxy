# AUDIT-0252 - Preserve keyed plaintext text entries

- Commit subject: `Preserve keyed plaintext text entries`
- Change type: `runtime`
- Runtime impact: `yes`

## Intent

Prevent the resource-pack producer from corrupting text files that are already
plaintext when `contents.json` declares a non-empty per-file key. The
regression was found in GALAXITE `README.txt` files. BedrockPackArchive must
continue to preserve the producer's derived bytes without decrypting them.

## Scope

- Resource-pack decryption now probes small text-like entries before applying
  AES-CFB8 and preserves bytes that are clearly plaintext while the decrypted
  candidate is not text.
- Added a deterministic keyed plaintext README regression test and a research
  note describing the data path and limitation.
- Updated the changelog and the coordinated BedrockPackArchive fixture and
  compatibility documentation in commit `9a56129`.
- Root manifest and icon handling, genuinely encrypted text entries, binary
  entries, capture schema, and the importer production path remain unchanged.

## Evidence and provenance

- A fresh GALAXITE pack audit reproduced the issue: the raw README bytes were
  valid plaintext, while the previous producer output was non-UTF-8 and
  contained control bytes.
- The fixed producer output kept the 21 GALAXITE README files valid UTF-8,
  including `Galaxite UI/README.txt`, and retained the `README.txt` entry in
  `contents.json` with its declared key.
- The coordinated BedrockPackArchive synthetic fixture proves that a keyed
  plaintext README is imported byte-for-byte and that its key declaration is
  retained without exposing the key value.

## Validation

- Automated checks: focused resource-pack tests, `go test ./...`, `go vet ./...`,
  the stamped BedrockPackArchive producer-verifier test, and the full
  `tools/quality.ps1 -LowMemory` gate passed. BedrockPackArchive tests, vet,
  build, and diff checks passed.
- Live validation: `pass` for the GALAXITE resource-pack exchange and fresh
  reconstruction. This does not claim that the incomplete Enchanted OP SMP
  route passed.
- Capture or report: `none`; temporary diagnostic data was not committed.

## Limitations and follow-up

The heuristic is intentionally limited to small text-like paths and preserves
bytes only when the source is clearly text and the AES-CFB8 candidate is not.
It does not recover bytes already corrupted by an older producer. GALAXITE
asset repositories require a private backfill from the fixed producer. The
Enchanted batch remains excluded until its OP SMP capture completes.
