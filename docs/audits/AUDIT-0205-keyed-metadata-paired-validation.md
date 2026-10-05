# AUDIT-0205 - Verify paired plaintext metadata compatibility

- Commit subject: `docs: record paired plaintext metadata validation`
- Change type: `documentation`
- Runtime impact: `none`

## Intent and revisions

Record the completed synthetic producer-consumer check after both focused commits exist. This commit adds only this audit.

| Component | Revision |
| --- | --- |
| BedrockDebugProxy plaintext metadata fix | `accd4d65c3d28f04643e1d2c9c2f0adbbae373f9` |
| BedrockPackArchive matched fixture and documentation | `e61a863acf78dc32f0bf649fb8189858cd877e2f` |

Both commits are signed. The consumer uses its existing private origin. The [runtime audit](AUDIT-0204-plaintext-pack-metadata.md) and [research note](../research/plaintext-resource-pack-metadata.md) record the independent source review, narrow root metadata rule, and synthetic reproduction. Neither source change contains a real capture, pack, account, ContentKey, or live per-file key.

## Completed validation

- Canonical formatting and `tools/quality.ps1 -LowMemory` passed on the reviewed runtime source with Go 1.26.6. The resource-pack package passed under the race detector. Static analysis found zero issues. Govulncheck found no project-reachable vulnerability and six uncalled dependency findings.
- BedrockPackArchive passed its full tests, vet, build, and whitespace checks. The importer contract is unchanged, so no importer production code change was needed.
- `tools/build.ps1 -Version 0.3.0-dev.metadata` stamped the clean producer revision above. The binary SHA-256 was `c120ff7b9bd5f100edd2341e80ad61f0730e9289c6cee11b8564a26125610761`.
- The committed consumer's TestImportCapturePreservesRootManifestAndPackMetadata passed with that stamped binary selected through BPA_TEST_BEDROCK_DEBUG_PROXY. The producer verified the synthetic capture before the importer checked metadata bytes, archive digest, contents.json, three declared keys, and source sequence.
- Whitespace checks passed. The final push workflow verifies both remote main refs and the producer's exact-head CI separately.

## Scope and follow-up

This audit cites only synthetic fixture evidence. It makes no retail-client, public-server, gameplay, or release-matrix compatibility claim. The maintainer-requested asset intake is a separate private archive operation and cannot substitute for the documented release gate.

For human validation, use an authorized encrypted-pack Experience. Confirm that reconstructed root manifests parse, root manifest and icon bytes match the original archive, and pack exchange completes with a verifiable closed capture. Retain the original capture and first derivation error locally if the flow fails. Root metadata copying does not authenticate the rest of an AES-CFB8 payload.
