# AUDIT-0241 - Prepare the 0.4.2 release

The maintainer requested publication after the maintenance review and release gates. This candidate updates only CHANGELOG.md and this audit. The validated runtime is af36246a91af3ea8f81e0524e99ff94628e9b85b. Reports retain that tested_revision. Bedrock 1.26.52, protocol 2193, Go 1.27.3, and klauspost/compress 1.18.7 identify the supported build.

The maintenance implementation is 34472e23c379bc68d289a9f2de08580c99d8c3e5. AUDIT-0237 records the capture-layout, viewer, cache-policy, rename, and signaling fixes and shorter README. AUDIT-0238 records the test deadline correction. AUDIT-0239 and AUDIT-0240 record the security updates discovered during the resumed release gate. Historical reports were preserved and the entire matrix was repeated after those updates.

The canonical race-enabled quality gate passed with zero static-analysis findings and zero called-function vulnerabilities. Both local replacement-module suites passed under a workspace selecting the actual compression version. Linux and Windows passed in [Quality run 39823126222](https://github.com/NhanAZ/BedrockDebugProxy/actions/runs/39823126222). The clean stamped runtime passed a browser smoke check of the synthetic viewer fixture, empty-blob details, and console. Earlier checks cover live closure and filtering. BedrockPackArchive commit 2e5c9e2fcbf21812e477198760ffcf994144dbb6 passed race tests, vet, build, and the producer compatibility fixture against this exact binary. Its private origin remains at that commit.

All six authenticated CLI sessions completed offered packs, spawned, retained bidirectional decoded traffic for at least 30 seconds, and closed with verified capture integrity and matching completed artifacts. No events were dropped or truncated and no decode or blocking errors occurred. Expected read-side termination remains visible after client completion.

| Server | Verified private evidence | Pack archives by hop | Final-spawn CLI stability |
| --- | --- | --- | --- |
| The Hive | [session-20261022T173227Z](https://github.com/BedrockDebugProxy-Captures/session-20261022T173227Z/releases/tag/capture-archive) | 25 | 35 seconds |
| Galaxite | [session-20261022T173436Z](https://github.com/BedrockDebugProxy-Captures/session-20261022T173436Z/releases/tag/capture-archive) | 20 | 35 seconds |
| Lifeboat | [session-20261022T173744Z](https://github.com/BedrockDebugProxy-Captures/session-20261022T173744Z/releases/tag/capture-archive) | 11 | 35 seconds |
| Mineville Zeqa | [session-20261022T173918Z](https://github.com/BedrockDebugProxy-Captures/session-20261022T173918Z/releases/tag/capture-archive) | 9 | 35 seconds |
| Enchanted | [session-20261022T174318Z](https://github.com/BedrockDebugProxy-Captures/session-20261022T174318Z/releases/tag/capture-archive) | 1 + 7 | 35 seconds |
| CubeCraft | [session-20261022T174540Z](https://github.com/BedrockDebugProxy-Captures/session-20261022T174540Z/releases/tag/capture-archive) | 5 | 35 seconds |

Enchanted selected the current OP SMP form, recorded its response and Transfer rewrite, authenticated the child connection, completed its seven packs, and reached child spawn. CubeCraft used the normal observation route after explicit maintainer acceptance of rejection, kick, or ban risk for this fresh matrix. That acknowledgement is not server permission or a policy exception.

Each archive preserves the complete original capture and a separate exact driver source bundle. Private visibility, uploaded SHA-256 digests, asset sizes, and unchanged source were verified. Captures, credentials, keys, and third-party pack content are excluded from source history, official release assets, and the private source backup. Public assets contain sanitized reports, the stamped Windows binary and checksum, project license and notices, and the generated dependency license bundle.

The candidate requires the documentation-only equivalence check against the runtime above, its own successful CI, a signed tag, published-asset verification, and backup-ref verification before completion. Git and GitHub metadata identify the final candidate and publication. CLI evidence does not establish retail rendering or gameplay. Live resource-pack decryption was disabled.
