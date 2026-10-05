# AUDIT-0203 - Record paired protocol maintenance validation

- Commit subject: `docs: record paired protocol maintenance validation`
- Change type: `documentation`
- Runtime impact: `no`

## Intent and paired revisions

Preserve the source identities and completed producer-consumer checks for the 2026-10-05 maintenance review.

| Repository or change | Verified commit |
| --- | --- |
| BedrockDebugProxy tested maintenance revision | `d687009d8a56392c4f78943ca00d4c77bd5c5c11` |
| BedrockDebugProxy empty-pack producer change | `0ff5a3f0d1560f012626cc3f7f0ee3e259745f6b` |
| BedrockPackArchive matched consumer fixture and module graph | `1b1b8b4a156453d444d403d27989a7f398e8b5cf` |

The producer's five focused maintenance commits and the consumer commit are signed. The BedrockPackArchive origin remains the existing private `NhanAZ-Tools/BedrockPackArchive` repository. No capture, authentication state, real pack, ContentKey, or decrypted third-party content is committed.

## Automated evidence

- The clean baseline and reviewed maintenance source passed `tools/quality.ps1 -LowMemory` with Go 1.26.6. The exact tested maintenance revision passed the same gate after the focused commits. Formatting used `tools/format.ps1`. Static analysis reported zero issues, and govulncheck reported no project-reachable vulnerability plus six uncalled dependency findings. The gate also passed viewer asset and HTTP tests, JavaScript syntax, root tests and loopback sessions, build, module integrity, license inventory, and whitespace checks.
- `third_party/go-raknet` passed `go test -race -count=1 ./...` with CGO enabled. The previous source reproduced both a data race and fatal concurrent map access.
- The replacement gophertunnel minecraft, NBT, and packet suites passed. Empty download requests retain both initial and active exchange state. Fixed-width NBT overflow, truncation, byte order, streams, offsets, and empty arrays pass on amd64 and 386.
- go-nethernet v1.0.24 passed its full module suite. The watcher passed seven mocked cases and the read-only real release check. No real issue was posted.
- BedrockPackArchive passed `go test -count=1 ./...`, `go vet ./...`, `go build -o bin/bedrock-pack-archive.exe .`, and `git diff --check`.
- `tools/build.ps1 -Version dev` stamped the clean producer revision above. `version` reported Bedrock 1.26.51 and protocol 2193. That binary's SHA-256 was `420eaa172a4e28f29f0dad274d50294c22c72c549beb3d1a16179885a35e985b`.
- Setting `BPA_TEST_BEDROCK_DEBUG_PROXY` to the stamped binary and running `go test -count=1 -v -run TestImportCapturePreservesRootManifestAndPackMetadata .` passed. The actual capture verifier checked the synthetic event schemas, sequence, manifest counters, canonical blob paths, sizes, and hashes before the importer checked archive and index data.

The fixture retains the raw empty-request bytes, archive digest, root and nested manifests, plaintext contents.json, synthetic per-file key count, and index source sequence. The consumer contract is unchanged, so no importer production code modification was needed. This audit and the research-note result section change documentation only.

## Live status and next validation

Public-server, retail-client, live NetherNet Experience, and official six-server release validation are `not_observed`. Synthetic and loopback results must not be described as proof of player login, spawn, gameplay, or public-server compatibility.

Use an owner-controlled RakNet server or authorized Featured Experience for ordinary login, pack exchange, and graceful disconnect. Use an authorized NetherNet target for negotiation, packet transport, identity propagation, and cancellation. Verify closed captures and retain the first failure log if either flow fails. Retail pack-cache behavior, cross-network ICE reachability, and release validation remain pending.

Compared with v0.3.0, the repository contains networking, resource-pack, and capture-related changes beyond this task. A future release checkpoint would establish a tested recovery boundary, but requires the documented fresh six-server retail-client gate. This task authorizes source updates and pushes, not a release tag or publication.
