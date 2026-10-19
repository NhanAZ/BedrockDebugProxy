# AUDIT-0236 - Prepare the CLI-validated 0.4.1 release

The maintainer explicitly requested CLI validation without the retail Minecraft application and publication of version 0.4.1. AUDIT-0234 records the policy decision and AUDIT-0235 implements its report gates. The validated runtime is bc595ed64632a3eb48ac1758f94b32e35f5da7e9. This candidate adds only the root Markdown changelog and this audit. Preserve that runtime revision in all reports and use the explicit documentation-only equivalence before tagging.

## Live evidence

The six fresh owner-controlled sessions used the same clean stamped 0.4.1 binary, approved cached Microsoft account, concrete loopback listener, and ordinary downstream CLI protocol client. The proxy authenticated upstream through each Experience selector. Each report passes every automatic requirement and the cli_session, authenticated_upstream, post_spawn_stability, experience_resolution, and packs_exchange review checks.

| Server | Verified private capture archive | Retained pack archives | Decoded traffic after final spawn in seconds |
| --- | --- | --- | --- |
| The Hive | [session-20261019T052100Z](https://github.com/BedrockDebugProxy-Captures/session-20261019T052100Z/releases/tag/capture-archive) | 25 | 34.966 |
| Galaxite | [session-20261019T052304Z](https://github.com/BedrockDebugProxy-Captures/session-20261019T052304Z/releases/tag/capture-archive) | 20 | 34.999 |
| Lifeboat | [session-20261019T052603Z](https://github.com/BedrockDebugProxy-Captures/session-20261019T052603Z/releases/tag/capture-archive) | 11 | 34.983 |
| Mineville Zeqa | [session-20261019T052714Z](https://github.com/BedrockDebugProxy-Captures/session-20261019T052714Z/releases/tag/capture-archive) | 9 | 34.952 |
| Enchanted | [session-20261019T053051Z](https://github.com/BedrockDebugProxy-Captures/session-20261019T053051Z/releases/tag/capture-archive) | 8 | 35.963 |
| CubeCraft | [session-20261019T053253Z](https://github.com/BedrockDebugProxy-Captures/session-20261019T053253Z/releases/tag/capture-archive) | 5 | 34.982 |

Enchanted follows the current server-sent OP SMP form choice, Transfer, child pack exchange, and child spawn. The two upstream hops retain equal nonzero RakNet GUID and UDP source values. A best-effort route-probe DNS timeout remains in the capture while the normal child connection succeeds. Its report includes transfer_following=pass and resource_pack_transfer=pass. No other minigame is selected. The Hive is lobby-only and Mineville BEDWARS is excluded. CubeCraft uses the ordinary read-only route after the maintainer's accepted rejection, kick, or ban risk, recorded as risk_acknowledged=pass. This acknowledgement is not a policy exception.

Each complete capture was preserved unchanged in its own private archive repository. The private README records the ZIP SHA-256, and the uploaded asset's size and digest match the prepared ZIP. Original captures, credentials, keys, packet payloads, and third-party pack contents are excluded from source history, official release assets, and the recovery mirror.

## Automated checks and provenance

The runtime passed canonical tools/format.ps1 and tools/quality.ps1 -LowMemory, focused report and matrix regressions, and both jobs of [Quality run 39204069869](https://github.com/NhanAZ/BedrockDebugProxy/actions/runs/39204069869). Static analysis reported zero issues and the vulnerability scan found no called vulnerabilities. The exact-runtime six-server release-readiness gate passes. The candidate pre-commit canonical format and quality gates also passed, with zero static-analysis issues and no called vulnerabilities. Focused replacement-module deferred-completion and archive-stream regressions passed with race detection. Verify the clean candidate build, exact-head CI, and documentation-only equivalence before publication.

The private BedrockPackArchive consumer remains at 13456426b87e980f2c96ca294fae1d14978d9ce7. Its focused root-manifest and pack-metadata fixture passes with this exact runtime as the actual producer verifier. The report-tooling changes alter no producer or importer contract and need no consumer patch. Earlier paired resource-pack runtime checks remain documented in AUDIT-0231. Go dependencies and licenses are unchanged. The release includes the project license, notices, compiled-module and embedded-asset license bundle, and Go runtime license.

Review since v0.4.0 identifies the focused stream-read error, deferred pack completion, retained route-hint, and fast reconnect identity fixes. No packet schema or viewer code changed. Bedrock 1.26.52 and protocol 2193 remain selected. The release uses a signed annotated tag and a clean stamped Windows binary, checksum, and six sanitized reports.

The private recovery mirror still holds the source history from before its public privacy cleanup. After verifying the release, preserve its existing branch and tag objects under dated backup-history refs before aligning the current source refs with explicit leases. This reviewed, recoverable update is limited to the private mirror and must verify every source branch and tag.

## Limits

CLI sessions establish the observed authentication, packs, spawn, traffic, transfer, and capture flows. Retail rendering, UI, input handling, and visible gameplay remain unobserved and are not required by the accepted CLI release criterion. These sessions do not establish every minigame, pack variant, client build, or future server state. Historical incomplete reports and failed BEDWARS evidence remain unchanged.
