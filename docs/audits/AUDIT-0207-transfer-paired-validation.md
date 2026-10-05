# AUDIT-0207 - Verify paired transfer compatibility

- Commit subject: `docs: record paired transfer validation`
- Change type: `documentation`
- Runtime impact: `none`

## Intent and revisions

Record the completed synthetic producer-consumer check after both focused commits exist. This commit adds only this audit.

| Component | Revision |
| --- | --- |
| BedrockDebugProxy transfer completion fix | `c5b1bf21dbb5748c940d495ebc6e14b20bec412a` |
| BedrockPackArchive transferred-hop fixture | `e031b79c892c0c9ee63033133bab825c8a10a95f` |

The [runtime audit](AUDIT-0206-transfer-close-order.md) and [research note](../research/transfer-close-order.md) record the local root cause, deterministic reproduction, narrow completion rule, and unchanged capture contract. No dependency code was copied or adapted.

## Completed validation

- Canonical formatting and tools/quality.ps1 -LowMemory passed with Go 1.26.6. The completion-order regression and existing two-hop transfer loopback test passed under the race detector. Static analysis found zero issues. Govulncheck found no project-reachable vulnerability and six uncalled dependency findings.
- BedrockPackArchive passed full tests, vet, build, and whitespace checks with Go 1.26.6. No importer production change was needed because the event and archive contract is unchanged.
- tools/build.ps1 stamped the clean producer revision above as `0.3.0-dev.transfer`. Its binary SHA-256 was `c84e9089be68f5671830c1a6c4ed4819fb149d738a00c7e4065131f931a64923`.
- The committed consumer fixture passed with that binary selected through BPA_TEST_BEDROCK_DEBUG_PROXY. Both hop cases were verified by the actual producer before the importer checked archive digest, metadata bytes, contents.json, declared key count, source sequence, hop, and menu path.
- The publication workflow verifies both existing remote main refs, consumer privacy, and exact-head producer CI. Sanitized development reports for the separate maintainer-requested asset intake remain local. Their incomplete result is not a retail-client or release pass.

## Limits and human validation

This audit cites synthetic fixture evidence only. The private pack intake observes server-offered archives and loopback client transfer flows. Its child client can intentionally stop before the downstream spawn handshake completes, which preserves spawn cancellation diagnostics and cannot prove a player can enter or remain in a world. No capture, server asset, ContentKey, account, or live per-file key is committed here.

For human validation, select an Enchanted child mode through its server-sent menu. Confirm Transfer, hop 2 login and pack exchange, child spawn, and continued connection using an authorized retail client. Retain the capture and terminal diagnostics locally if the flow fails. The six-server release gate remains separate.
