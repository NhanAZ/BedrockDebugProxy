# AUDIT-0226 - Validate deferred pack completion with authenticated sessions

This is a documentation-only follow-up with no runtime impact. It records the tested runtime without changing the executable.

## Revisions and scope

| Component | Revision |
| --- | --- |
| BedrockDebugProxy FIX-0161 runtime | `bb9c9e9d0c323d19ec5bda4d5609e82a56e8e33e` |
| BedrockPackArchive completion-order fixture | `a8114b8d1e1b89bec697a1228d00efaff964a0c3` |

The [runtime audit](AUDIT-0225-deferred-resource-pack-completion.md) and [decision](../decisions/0014-defer-upstream-pack-completion.md) record the root cause, implementation, and alternatives. The clean runtime was built as 0.4.1-dev with Bedrock 1.26.52 and protocol 2193. Its executable SHA-256 is `e9160a5cd0bcc31ec7c7082084080169681e627adaf8afe6d0d70f2b8c857e0a`.

The maintainer authorized using the alternate account already cached locally. The normal Microsoft authentication path was used, and every retained upstream identity view was checked against the selected account. The loopback client selected the actual Enchanted menu entries, followed the returned transfers through the proxy, completed spawn, read ordinary traffic, and handled the server's normal latency requests. Each test session used a fresh client cache directory. The Hive used no client pack cache. The runtime proxy did not synthesize gameplay or latency replies while serving packs.

## Exact-runtime outcomes

All four captures closed normally, retained their offered source archives, completed spawn, and remained connected for the 60-second observation interval after the final hop spawned. The reported intervals differ from 60 by a few milliseconds because the proxy records session.spawned after the client receives its spawn signal. All capture-derived automatic requirements pass, with zero blocking errors, decode errors, unknown packets, and resource-pack decryption errors. Ordinary shutdown diagnostics remain recorded. No server Disconnect packet was observed.

| Target | Source archive counts by hop | Final post-spawn interval | Verified private capture |
| --- | --- | --- | --- |
| The Hive | 25 | 59.999 seconds | [session-20261018T173211Z](https://github.com/BedrockDebugProxy-Captures/session-20261018T173211Z/releases/tag/capture-archive) |
| Enchanted / OP SMP | 1 + 7 | 59.998 seconds | [session-20261018T173227Z](https://github.com/BedrockDebugProxy-Captures/session-20261018T173227Z/releases/tag/capture-archive) |
| Enchanted / Box PvP | 1 + 7 | 59.999 seconds | [session-20261018T173500Z](https://github.com/BedrockDebugProxy-Captures/session-20261018T173500Z/releases/tag/capture-archive) |
| Enchanted / Mines | 1 + 10 | 59.995 seconds | [session-20261018T173730Z](https://github.com/BedrockDebugProxy-Captures/session-20261018T173730Z/releases/tag/capture-archive) |

The counts are retained upstream archives. An Enchanted row includes the Hub pack before the child set. These are separate from a retail client's download prompt or cache-dependent count. The Box PvP set includes its dedicated mode pack and six shared archives, consistent with the prior recorded offer.

## Completion-order evidence

The retained packet bodies were decoded with the matching protocol implementation. Every hop contains exactly one downstream and one upstream PackResponseCompleted. The downstream completion precedes the upstream completion, and session.spawned follows both.

| Target and hop | Downstream completion sequence | Upstream completion sequence | Spawn sequence |
| --- | --- | --- | --- |
| The Hive, hop 1 | 1427 | 1428 | 1663 |
| Enchanted / OP SMP, hop 1 | 85 | 86 | 137 |
| Enchanted / OP SMP, hop 2 | 872 | 873 | 1040 |
| Enchanted / Box PvP, hop 1 | 85 | 86 | 138 |
| Enchanted / Box PvP, hop 2 | 873 | 874 | 999 |
| Enchanted / Mines, hop 1 | 85 | 86 | 137 |
| Enchanted / Mines, hop 2 | 1802 | 1803 | 2190 |

The OP SMP cold-cache result resolves the startup timing failure left open by [AUDIT-0224](AUDIT-0224-paired-stream-error-validation.md). Delaying the final resource-pack response prevents upstream world startup from occupying the downstream download interval. The separately observed authentication outages, HTTP failures, and Mines backend refusal remain historical observations rather than rewritten successes.

## Paired and automated validation

The canonical tools/format.ps1 and tools/quality.ps1 -LowMemory checks pass. Root focused checks cover ordinary forwarding, mixed URL and chunk offers, deferred completion order, and the unchanged packs-only boundary. Nested minecraft tests and focused root and nested race checks pass. The old early-completion configuration fails the deterministic ordering regression.

The consumer's full tests, vet, build, and whitespace checks pass. Its root and transferred-hop importer matrix also passes with the actual stamped producer executable, covering both 1.26.51 and 1.26.52 fixtures, short and large language metadata, archive digests, root and nested manifests, plaintext contents.json, three declared synthetic keys, generator version, source sequence, hop, and menu path. The importer data contract is unchanged. The source changes are recorded as separate local commits. BedrockPackArchive remains private, and no source commit was pushed during this task.

The sanitized machine-readable reports retain tested_revision `bb9c9e9d0c323d19ec5bda4d5609e82a56e8e33e`. Each report deliberately has result incomplete and normal_session not_observed because the client is headless. Its synthetic_spawn, packs_exchange, and post_spawn_60s checks pass, with transfer also passing for Enchanted. The automatic requirements pass. This is scoped live protocol evidence, not retail rendering, player movement, gameplay, or an official release gate.

## Archive verification

Each capture was kept unchanged as a complete ZIP release asset on the linked private capture repository. Repository privacy, README contents, asset byte size, and the GitHub SHA-256 asset digest were verified against the prepared ZIP. Source captures remain local. The Hive run disabled decryption. Enchanted source packs are unencrypted and none of these captures contains a derived decrypted archive event.

| Session | Verified ZIP SHA-256 | ZIP bytes |
| --- | --- | --- |
| [session-20261018T173211Z](https://github.com/BedrockDebugProxy-Captures/session-20261018T173211Z/releases/tag/capture-archive) | `3e225af60aa4a248397d64e66af040bbdf2eb51db48f28c3a56ca3f51120403b` | 133,219,461 |
| [session-20261018T173227Z](https://github.com/BedrockDebugProxy-Captures/session-20261018T173227Z/releases/tag/capture-archive) | `180237fe4c1fcf8e456b785e5a10ce471dddbe6f8df0500458885ca30bdcefb8` | 66,530,743 |
| [session-20261018T173500Z](https://github.com/BedrockDebugProxy-Captures/session-20261018T173500Z/releases/tag/capture-archive) | `00901f5baca5e747813aa40240e7a40062c8eb6e3f7a568ccd3becaaf1f35c9f` | 67,292,158 |
| [session-20261018T173730Z](https://github.com/BedrockDebugProxy-Captures/session-20261018T173730Z/releases/tag/capture-archive) | `e7d812d8c4f5ab5f0a11621b45eff646187624e8e4d1dd10671d5b29355a59a2` | 160,807,672 |

## Future retail release check

For a future release, enter Enchanted Hub with a retail client whose OP SMP packs are not cached, select OP SMP, and observe pack completion followed by world spawn and ordinary visible play. The capture should show downstream ResourcePackClientResponse completion before upstream completion, followed by StartGame and both-hop session.spawned records without an upstream Disconnect. If it fails, preserve the complete closed capture and inspect the completion responses, Disconnect, transport errors, and decoder evidence. Complete the existing six-server release gate separately before publishing a release.
