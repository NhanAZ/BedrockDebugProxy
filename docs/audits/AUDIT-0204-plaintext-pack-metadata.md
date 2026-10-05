# AUDIT-0204 - Preserve plaintext pack metadata

- Commit subject: `fix: preserve plaintext resource-pack metadata`
- Change type: `runtime | resource packs`
- Runtime impact: `yes`

## Intent, scope, and evidence

Fix the new decoder root cause reproduced by TestDecryptArchivePreservesKeyedRootMetadata. A keyed contents declaration transformed plaintext root metadata in the derived archive. The focused decoder change copies the selected root manifest and root icons without applying their declared keys. It retains all original evidence and key declarations, and continues decrypting unrelated paths with matching basenames.

The [research note](../research/plaintext-resource-pack-metadata.md) records the source revision, license, independent implementation review, and the failing synthetic reproduction. No third-party code was copied or adapted. The change affects internal/resourcepack, its regression test, resource-pack documentation, fix history, and the changelog. It adds no CLI, schema field, dependency, or offline key path.

BedrockPackArchive extends its synthetic importer fixture and compatibility documentation in commit `e61a863acf78dc32f0bf649fb8189858cd877e2f`. No importer production code change is needed because the capture contract and reconstructed archive format are unchanged. The paired final audit cross-references both commit identities after the producer commit exists.

## Validation

- The new test failed on the previous decoder for all three root metadata paths and passes under the race detector after the fix. Both flat and nested roots preserve metadata bytes, while keyed files with matching basenames outside the metadata root decrypt normally.
- BedrockPackArchive passed its full tests, vet, build, and whitespace checks. Its fixture preserves root manifest and icon bytes, the selected archive digest, contents.json, three declared-key entries, and source sequence.
- Canonical `tools/format.ps1`, `tools/quality.ps1 -LowMemory` with Go 1.26.6, and whitespace checks passed. Static analysis found zero issues. Govulncheck found no project-reachable vulnerability and six uncalled dependency findings. The matched importer fixture passed with the actual verifier built from the reviewed producer source. The paired final audit records the later stamped check and exact revisions.
- No real asset, capture, ContentKey, or account state is included in either repository change. This audit relies on synthetic fixtures and cites no capture.

## Human validation and limitations

Use an authorized Experience that offers encrypted packs with plaintext metadata, such as the maintainer-requested Galaxite archive session. Exercise login and the pack exchange, then inspect every reconstructed root manifest and compare root metadata bytes with the original archive. Capture verification must pass, and the contents declarations and key count must remain intact. Keep the closed capture and first derivation error locally if the flow fails.

The result remains AES-CFB8 without authentication. Automated fixture success does not prove plaintext integrity for arbitrary files, retail-client spawn, gameplay, or the six-server release matrix.
