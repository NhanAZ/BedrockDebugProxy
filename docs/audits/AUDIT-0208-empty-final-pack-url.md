# AUDIT-0208 - Accept an empty final resource-pack URL

- Commit subject: `fix: retain chunk-only pack offers without duplicate archives`
- Change type: resource-pack capture compatibility
- Runtime impact: valid offers ending in an empty URL retain offer order without appending the downloaded cache list again.

## Intent and scope

Fix the demonstrated auxiliary decoder boundary by matching the canonical decoder's bytes.Buffer input. This is a local integration regression of FIX-0075. Packet definitions, merge selection, raw bytes, encryption, decryption, schema, CLI, and viewer interfaces remain unchanged. No third-party source or new dependency is introduced.

Affected files are resource_pack_url.go, its regression tests, the packs-only loopback test, resource-pack guide, changelog, fix history, and linked research. The paired consumer fixture includes an empty-URL offer at both hops. Its importer contract is unchanged, so no importer production change is needed.

## Validation

The decoder and chunk-only merge regressions fail before the fix with EOF and four retained packs for a two-pack offer. Focused proxy cases pass under the race detector after the fix, including malformed offers, mixed ordering, and packs-only loopback capture. The loopback now requires one archive for one offered pack.

Canonical formatting and tools/quality.ps1 -LowMemory passed with Go 1.26.6, including tests, static analysis, build, viewer syntax and HTTP checks, and whitespace validation. Static analysis reported zero issues. Govulncheck reported no project-reachable vulnerability and six uncalled dependency findings. Consumer tests, vet, build, and diff checks passed with Go 1.26.6.

The matched private BedrockPackArchive fixture is b1171da8238419bbdba410c97a13e71c516eeb86. A subsequent paired audit records this producer revision, the stamped verifier check, public-server evidence, and both remote revisions before the matched pair is published.

## Live boundary and human validation

This runtime commit cites synthetic evidence only. A stamped The Hive packs-only session will compare the raw offer and stack with archive events, hashes, contents.json, reconstructed folders, and index. Expect one archive per offered entry and no tool-generated second copy. The report remains incomplete because the flow stops before spawn.

For optional human validation, connect an authorized retail client to The Hive through this revision, download the packs, and observe world spawn. If it fails, retain the closed capture and terminal diagnostics. Automated and packs-only checks do not establish retail gameplay or the six-server release gate.
