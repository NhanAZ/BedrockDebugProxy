# AUDIT-0224 - Verify paired stream-error correction

- Commit subject: `docs: record paired stream-error validation`
- Change type: documentation
- Runtime impact: none. This follow-up does not change the stamped runtime.

## Revisions and automated checks

| Component | Git revision |
| --- | --- |
| BedrockDebugProxy FIX-0160 runtime | `09a1548f76ef4029972e028a8f954bd6ed29be60` |
| BedrockPackArchive 0.4.x gate and producer fixture | `cbfa0879f24b141873bef4042e8de048e1f38cef` |

The [runtime audit](AUDIT-0223-resource-pack-stream-errors.md) and [failure review](../research/capture-failure-review-2026-10-18.md) record the narrow defect, original capture evidence, source path, and external failures. No upstream issue or pull request was published.

- Canonical tools/format.ps1 and tools/quality.ps1 -LowMemory pass with Go 1.27.2. Root tests, build, viewer syntax, static analysis, module/license integrity, vulnerability scanning, and whitespace checks pass. The vulnerability scan reports no project-reachable finding and retains the uncalled dependency findings.
- The three synthetic reader failure cases fail before the runtime correction and pass after it. Successful archive bytes are unchanged. The full vendored resource package tests and the URL cache's incomplete-HTTP-body regression pass.
- BedrockPackArchive full tests, vet, build, and whitespace pass. Its fixture passes with `BPA_TEST_BEDROCK_DEBUG_PROXY` selecting the actual clean stamped producer below. The index retains generator version 0.4.0, archive digest, root/nested manifests, plaintext metadata, contents.json, three synthetic declarations, source sequence, both hops, and child path. No importer data-path change was needed.
- The adapter version tests fail on the old 0.3.x-only gate for 0.4.0 and 0.4.1-dev, then pass after its correction. Bare dev and unreviewed versions remain rejected.
- tools/build.ps1 built the clean runtime revision as `0.4.1-dev`. Binary SHA-256 is `b1071a70c69e8bdf8350d80d359e61b1f34d58b6f2df425017d26bf1671f98d7`.

## Authorized live checks

The maintainer-selected alternate account was selected through the existing CLI. Its cached token refreshed successfully. The account selection remains local and does not change source files. The live checks all use the stamped runtime revision above. No new CubeCraft session was made.

| Target | Child or initial packs | Evidence and actual scope |
| --- | ---: | --- |
| The Hive | 25 | [session-20261018T165449Z](https://github.com/BedrockDebugProxy-Captures/session-20261018T165449Z/releases/tag/capture-archive). Packs-only completed. No listener, spawn, or gameplay was requested. Report incomplete. |
| Enchanted / OP SMP | 7 | [session-20261018T165644Z](https://github.com/BedrockDebugProxy-Captures/session-20261018T165644Z/releases/tag/capture-archive). Child upstream pack exchange completed, but the loopback child login did not complete before upstream Disconnect. Report fail. |
| Enchanted / Box PvP | 7 | [session-20261018T165802Z](https://github.com/BedrockDebugProxy-Captures/session-20261018T165802Z/releases/tag/capture-archive). Loopback child login completed after menu and Transfer. Client intentionally closed before downstream spawn. Report incomplete. |
| Enchanted / Mines | 10 | [session-20261018T165927Z](https://github.com/BedrockDebugProxy-Captures/session-20261018T165927Z/releases/tag/capture-archive). Loopback child login completed after menu and Transfer. Client intentionally closed before downstream spawn. Report incomplete. |

Enchanted child counts exclude the repeated Hub archive. The existing menu-session functions from the committed consumer were exercised by a temporary local harness, without adding a product command or changing production navigation. Its per-route results, not the harness process exit alone, determine the outcomes above. Sanitized machine-readable reports remain local and preserve their exact tested_revision. They do not mark these sessions as retail validation.

Box PvP and Mines manifests remain failed because the existing archive client closes immediately after child login and pack delivery, canceling the proxy's downstream StartGame. This is intentional scoped teardown evidence, not a normal player session. The original error and status are not suppressed or rewritten. OP SMP instead receives an upstream internal-error Disconnect before the loopback child login completes and remains an unresolved failed route. Each live capture passes the integrity verifier and is archived unchanged in a verified private repository with matching uploaded size/SHA-256. No capture, authentication state, or third-party pack content is committed to source.

## Remaining OP SMP timing evidence

The failed original OP SMP capture records upstream PlayStatusPlayerSpawn at about 126.408 seconds and upstream Disconnect at 151.920 seconds. Downstream reports its packs downloaded only at 171.556 seconds. The successful original OP SMP capture reports upstream PlayStatusPlayerSpawn at 120.445 seconds and downstream packs downloaded at 120.554 seconds. The new failed OP SMP test reports upstream PlayStatusPlayerSpawn at 20.764 seconds, upstream Disconnect at 46.609 seconds, and downstream packs downloaded at 60.366 seconds. These values come from monotonic capture elapsed timestamps.

Both failed observations therefore terminate upstream while downstream pack delivery is still pending. This is a demonstrated timing correlation with the proxy's documented upstream-first buffering, not proof of a particular backend timeout or a packet-definition defect. The narrow stream-reader correction does not resolve it. No pacing bound was removed, timeout extended, endpoint substituted, or protocol sequence changed to force a pass. A controlled retail comparison with the current pack cache state is still needed before selecting a broader startup change.

For human validation, use an authorized retail client through the new binary, enter Enchanted, select OP SMP, and retain the complete capture through child loading or failure. Observe Transfer, both pack exchanges, downstream child spawn, and continued connection. The full release matrix remains unrun. The local commits are not pushed in this task.
