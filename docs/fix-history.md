# Fix history

This is the agent-facing index of investigated bugs and compatibility fixes. It is a navigation aid, not a replacement for Git history, architecture decisions, research notes, captures, or validation reports. Git remains the authoritative source for the complete commit history.

## How to use this index before changing code

1. Search this file for the reported symptom, server, packet family, and error text.
2. Open the linked decision or research note and inspect the cited capture or validation report when it exists.
3. Classify the report as a known fix, a regression, a known limitation, a duplicate, or a new root cause. Do not create a second implementation for a known fix.
4. If the evidence shows a different root cause, add a new `FIX-NNNN` entry in the same coherent change as the fix and link the old entry as related history.
5. Keep server-specific observations separate. A passing capture on one server is not evidence that another server uses the same packet order, timing, optional packets, transport, or resource-pack policy.

Every new entry should state the symptom, root cause, implementation commits, evidence, validation status, limitations, and follow-up work. The per-commit record in [`docs/audits/`](audits/) explains the exact change that introduced or updated an entry.

## Compatibility and bug index

### FIX-0001 - Experience destination resolution

- Symptom: default or named experience destinations could resolve to the wrong address or fail to select the intended featured experience.
- Root cause: destination resolution and the default experience path did not share one explicit selector boundary.
- Implementation: `42ae317`, `68b9795`.
- Evidence and validation: [`docs/research/experience-routing.md`](research/experience-routing.md); local experience routing tests.
- Status: fixed. Recheck the resolver when adding or renaming an Experience.

### FIX-0002 - Device authentication reuse and logout

- Symptom: repeated runs required unnecessary device-code authentication, and switching accounts had no supported CLI action.
- Root cause: the device credential cache had no operator-facing lifecycle command or one-click login URL.
- Implementation: `eb8eab0`, `483614f`, `88ab196`.
- Evidence and validation: auth cache tests and CLI checks. The cached token remains sensitive and is never committed.
- Status: fixed. Account switching requires the explicit logout command before the next device login.

### FIX-0003 - URL resource-pack offers

- Symptom: servers offering HTTP(S) resource packs could fall back to incomplete delivery or disconnect before the client received the advertised pack.
- Root cause: URL offers were not reconstructed and delivered through the same session-bound pack path as protocol chunk downloads.
- Implementation: `112ceb4`, `d46a10f`, `dd8845e`.
- Evidence and validation: [`docs/resource-packs.md`](resource-packs.md), [`docs/research/galaxite-resource-pack-flow.md`](research/galaxite-resource-pack-flow.md), and captured pack archives.
- Status: fixed for the supported current-connection flow. Offline retrieval, key guessing, and arbitrary remote pack download remain out of scope.

### FIX-0004 - Coordinated shutdown and upstream cancellation

- Symptom: after the downstream client left, an upstream read or recorder worker could continue and print a misleading error.
- Root cause: coordinated shutdown did not cancel all upstream work at the same connection boundary, and recorder-close races were logged at error level.
- Implementation: `3f32962`, `813e19f`, `5f553fa`.
- Evidence and validation: loopback shutdown tests and closed-capture verification.
- Status: fixed. The final recorder warning is expected when a read finishes after the recorder closes; it is not a capture-integrity failure by itself.

### FIX-0005 - Featured-experience transfer following

- Symptom: a server `Transfer` ended the first hop, but the client remained on a loading screen because the next backend was not reached or its RakNet route was not ready.
- Root cause: the proxy did not follow transfers, or the next hop lost the route and handshake state required by the featured backend.
- Implementation: `9878bb5`, `ab23673`, `d1929d4`.
- Evidence and validation: [`docs/decisions/0003-opt-in-transfer-following.md`](decisions/0003-opt-in-transfer-following.md), [`docs/research/enchanted-transfer-handshake.md`](research/enchanted-transfer-handshake.md), and the successful Enchanted transfer capture cited there.
- Status: fixed for the opt-in `--follow-transfers` RakNet path. The option is intentionally opt-in and each hop has its own connection and pack exchange.

### FIX-0006 - Featured-experience spawn without `ChunkRadiusUpdated`

- Symptom: Enchanted and Lifeboat could send `PlayStatusPlayerSpawn` without the radius acknowledgement that the upstream library expected, leaving the dialer blocked.
- Root cause: the connection finalizer required two signals even though the observed featured flow completed with one.
- Implementation: `0df6b63`.
- Evidence and validation: [`docs/research/bedrock-ecosystem.md`](research/bedrock-ecosystem.md) and `minecraft/conn_featured_experience_test.go`.
- Status: fixed for the observed optional-packet flow. The patch does not rewrite or suppress the packet.

### FIX-0007 - Slow resource-pack and transfer bounds

- Symptom: a slow but active featured resource-pack exchange could time out, while an unreachable transfer backend could leave the client waiting indefinitely.
- Root cause: the upstream login and transfer route had no useful bounded wait aligned with observed pack pacing.
- Implementation: `c94d105`, `0fc7ef8`.
- Evidence and validation: [`docs/resource-packs.md`](resource-packs.md), [`docs/validation.md`](validation.md), and timeout regression tests.
- Status: fixed with bounded waits. A timeout is recorded as an observable failure and does not silently discard the capture.

### FIX-0008 - Same-protocol raw forwarding

- Symptom: decoding and re-encoding an otherwise unchanged packet could alter optional-field encoding, batching, or other wire details.
- Root cause: the bridge used the typed read/write path even when both sides negotiated the same protocol.
- Implementation: `7691681`, `96d8a01`.
- Evidence and validation: [`docs/decisions/0004-same-protocol-raw-forwarding.md`](decisions/0004-same-protocol-raw-forwarding.md), packet-preservation tests, and raw forwarding timing evidence.
- Status: fixed. Typed decoding remains a derived observation; raw packet bytes remain authoritative.

### FIX-0009 - Console and shutdown observability

- Symptom: high-volume resource-pack traffic spammed the console, and orderly shutdown looked like an unexplained error or appeared to stop before capture verification finished.
- Root cause: operator summaries were emitted for every small batch and shutdown stages were not named consistently.
- Implementation: `a1ff5fe`, `5f553fa`, `2818a67`.
- Evidence and validation: logger tests, PocketMine color comparison, and closed-capture terminal runs.
- Status: fixed. Detailed evidence remains in the capture, while the console reports progress through forwarding stop, connection close, recorder close, artifact completion, and verification.

### FIX-0010 - Deferred login packets across valid server orderings

- Symptom: Enchanted hop 2 sent `ResourcePacksInfo` before `PlayStatus`; the offer was deferred and the client stayed on the resource-pack loading screen until the upstream wait expired.
- Root cause: changing the login state did not re-check packets already received for a later state. A later-phase packet could also sit ahead of the packet that advances the current state.
- Implementation: `a600bbf`, `7d94229`, `7a64a0d`.
- Evidence and validation: failed capture `captures/session-20260914T105050Z`, successful older Enchanted capture `captures/session-20260914T124716Z`, [`docs/decisions/0005-deferred-login-packet-drain.md`](decisions/0005-deferred-login-packet-drain.md), and login-order permutation tests.
- Status: fixed in tests and validated on Enchanted with the earlier state-aware implementation. The latest queue-scan implementation still needs an exact-revision The Hive baseline and a fresh five-server release matrix.

## Ordering safety contract

The deferred-packet change is deliberately narrow. `expectedIDs` remains the only authority for what the current login state accepts. The queue scan selects the earliest packet that is currently expected, leaves packets for later states queued, and never mutates or reorders bytes on the wire. The regression tests cover canonical resource-pack order, featured early `ResourcePacksInfo`, a later-phase packet queued ahead of the current phase, and all six permutations of the three login packets.

This is a compatibility guard, not permission to assume a universal Bedrock sequence. Before changing an expected set, inspect the packet definitions, the relevant capture, and the server-specific validation flow. A new server observation must add a focused test and update this index or a linked research note.

## Open and pending items

- Generate a stamped binary from the current commit and run The Hive minimum baseline.
- Re-run the direct Enchanted transfer and resource-pack flow with the same binary.
- Complete the exact-revision matrix for Galaxite, Lifeboat, Mineville Zeqa, and Enchanted after the The Hive baseline before release.
- If a server fails, record the capture and classify the failure before changing the generic state machine.
