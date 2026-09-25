# Audit 0154 - Validate authenticated PlayFab omission

## Intent

Record the real-client local validation of the authenticated `ClientData.PlayFabId` parity fix from audit 0153.

## Scope

- Added the post-fix evidence to [`docs/research/local-identity-parity.md`](../research/local-identity-parity.md).
- Updated FIX-0119 in [`docs/fix-history.md`](../fix-history.md) from pending real-client validation to confirmed owner-controlled loopback validation.
- No runtime, protocol, capture schema, or viewer code changed.

## Evidence and provenance

- The stamped binary revision was `970531e7d7132a4e0f889e1ab9ad735a8d4002c1`.
- Capture [captures/session-20260922T041204Z](https://github.com/BedrockDebugProxy-Captures/session-20260922T041204Z/releases/tag/capture-archive) closed with 12,750 events, 6,118 blobs, zero write errors, zero dropped events, zero truncation, and zero decode errors.
- The matching ProxyFingerprint report completed after `player_quit`, recorded `ClientData.PlayFabId` as absent, and emitted no `auth_client_playfab_mismatch` signal.
- The test used only the owner-controlled local PocketMine server. No public server, credential, raw capture, or identifier was added to the repository.

## Automated validation

- The implementation and nested dependency tests passed before the live session.
- The low-memory project quality gate passed before the live session.
- `git diff --check` passes for this documentation-only change.

## Live validation status

Owner-controlled real-client authenticated loopback passed. The upstream Android metadata boundary remains observed and documented.

## Known limitations and follow-up

This evidence does not establish compatibility with or enforcement behavior of any public server. A release still requires the documented authorized validation matrix.
