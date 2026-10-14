# AUDIT-0211 - Select stable Bedrock 1.26.52 client metadata

- Commit subject: `fix: select stable Bedrock 1.26.52 login metadata`
- Change type: protocol metadata and tooling
- Runtime impact: yes

## Intent and scope

The latest stable installed game is 26.52, while the selected login metadata still advertised 1.26.51. Change the current game-version constant and its regression. Retain protocol 2193, the reviewed packet definitions, raw forwarding, compression, authentication, and capture schema. Separate the watcher schema baseline from client metadata so an older schema does not appear as a new release and a future matching schema remains detectable.

## Evidence and provenance

[The maintenance review](../research/maintenance-2026-10-15.md) records the official game release, installed package, immutable latest stable Mojang schema, and reviewed upstream tag. There is no official 1.26.52 schema and no claim of exhaustive packet compatibility. The existing gophertunnel composite and original source-base declaration remain accurate.

## Validation

The protocol metadata regression and seven mocked watcher cases pass. The watcher cases failed before separating the schema baseline, including unchanged-release false positives and a missing future-schema notification. A read-only real watcher run also reports the unchanged 1.26.51 schema and protocol 2193. No real issue was posted. `tools/format.ps1` and `tools/quality.ps1 -LowMemory` pass with zero static-analysis issues and zero reachable vulnerabilities. Exact-revision live checks follow the reproduced skin-default correction.

BedrockPackArchive `8afee20944f45fad3faa545e7cdb13d923a26108` passes full tests, vet, build, and diff checks. Its fixture covers both prior and current producer metadata without changing archive or importer behavior. The same fixture passed with the stamped security-maintenance producer `a11ef5324b3ab37ff0f1b5679f75e766fcf30b95` and the paired security consumer `816d23aba51f95c5799a256389cf22c5526d9087`. A new stamped metadata producer check follows this commit. No retail-client session or release is claimed.

## Human validation

Join an owner-controlled server using the retail 26.52 client through the proxy. Confirm login, offered packs, spawn, and both traffic directions. Keep the closed capture and terminal diagnostics if any stage fails. Automated packs-only public-server checks remain separate from this gameplay observation.
