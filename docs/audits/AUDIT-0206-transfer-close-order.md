# AUDIT-0206 - Preserve transfer after a racing client close

- Commit subject: `fix: retain transfer after a racing client close`
- Change type: `networking compatibility`
- Runtime impact: opt-in transfer following recognizes a completed second-loop transfer after a context-canceled client close.

## Intent and scope

Correct the existing runHop completion path without changing protocol definitions, packet bytes, routing policy, CLI, capture schema, or importer interfaces. Both loop errors and the original transfer source address and GUID remain observable. Unrelated first-loop failures are not treated as successful transfers.

Affected files are internal/proxy/proxy.go, its focused transfer tests, the changelog, fix history, and the linked research note. No dependency source was copied or adapted. The unchanged viewer event contract is covered by the canonical asset and HTTP tests.

## Evidence and validation

The deterministic client-close-first and wrapped-transfer cases failed with the original completion selection. The focused tests and two-hop loopback regression passed under the race detector. Canonical formatting and tools/quality.ps1 -LowMemory passed with Go 1.26.6, including static analysis, build, viewer, and whitespace checks. Static analysis found zero issues. Govulncheck found no project-reachable vulnerability and six uncalled dependency findings.

BedrockPackArchive's matched fixture at `e031b79c892c0c9ee63033133bab825c8a10a95f` extends the existing synthetic pack case to both initial and transferred hops. It checks a prior session.close event, child navigation path, source sequence, root and nested manifests, pack icon, archive hash, contents.json, three synthetic key declarations, and index data. The consumer contract is unchanged, so no importer production change is needed. Full consumer tests, vet, build, and whitespace checks passed with Go 1.26.6. The stamped verifier result and both revision identities are recorded before the paired push in a subsequent audit.

## Live boundary and follow-up

This commit cites only synthetic evidence. The requested private asset intake is separate from release validation. A stamped Enchanted traversal and The Hive packs-only baseline are planned after the tested revision is committed. Retail-client gameplay and the six-server release matrix are unobserved.

For human validation, select an Enchanted child mode in an authorized client session. Observe the server Transfer, hop 2 pack exchange, and child spawn. The client should reconnect through the proxy and remain in the child world. If it fails, retain the closed capture and terminal diagnostics locally. Automated loopback spawn is not evidence of retail-client gameplay.
