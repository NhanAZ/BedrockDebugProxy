# Audit 0156 - Validate empty authenticated PlayFab ID omission

## Intent

Record the live local validation of the empty `ClientData.PlayFabId` JSON omission from audit 0155.

## Scope

- Add the stamped real-client loopback evidence to [`docs/research/local-identity-parity.md`](../research/local-identity-parity.md).
- Update FIX-0119 in [`docs/fix-history.md`](../fix-history.md) from pending wire-shape validation to confirmed local validation.
- No runtime, protocol, capture schema, or viewer code changed.

## Evidence and provenance

- The tested binary was revision `9994e4e11f595f7980fc6795f5cc8b05271d4de7`.
- Capture [captures/session-20260922T043146Z](https://github.com/BedrockDebugProxy-Captures/session-20260922T043146Z/releases/tag/capture-archive) closed with 8,250 events, 3,837 blobs, zero write errors, zero dropped events, zero truncation, and zero decode errors.
- ProxyFingerprint report `20260922T043303.076778Z_47906eadaa099058_unlabeled.json` completed after `player_quit`, recorded `ClientData.PlayFabId` as absent, and emitted only `non_public_peer`.
- PocketMine recorded `XUID match` and no `Unexpected JSON property for ClientData: PlayFabId` warning for this session.
- The test used only the owner-controlled local PocketMine server. No public server, credential, raw capture, or identifier was added to the repository.

## Automated validation

- The runtime change passed the nested gophertunnel tests, root tests, low-memory quality gate, and `git diff --check` before the live session.
- This commit is documentation-only and has no runtime impact.

## Live validation status

Owner-controlled real-client authenticated loopback passed. The empty-key wire-shape correction is confirmed locally.

## Known limitations and follow-up

This evidence does not establish compatibility with or enforcement behavior of any public server. The authenticated Android metadata boundary remains documented, and a release still requires the documented authorized validation matrix.
