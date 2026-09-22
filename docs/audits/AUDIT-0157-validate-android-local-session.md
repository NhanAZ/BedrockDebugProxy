# Audit 0157 - Validate Android local session

## Intent

Record the owner-controlled Android client validation of the stamped authenticated proxy build.

## Scope

- Add Android downstream evidence to [`docs/research/local-identity-parity.md`](../research/local-identity-parity.md).
- Add the observed real-client NetworkSettings mismatch to [`docs/research/network-settings-parity.md`](../research/network-settings-parity.md).
- No runtime, protocol, capture schema, or viewer code changed.

## Evidence and provenance

- The tested binary was revision `9994e4e11f595f7980fc6795f5cc8b05271d4de7`.
- Capture `captures/session-20260922T044324Z` closed with 11,325 events, 6,235 blobs, zero write errors, zero dropped events, zero truncation, and zero decode errors.
- The downstream client was Android account `NhanAZ`. The authenticated upstream local session used the cached `SrDuck84` token and PocketMine observed `SrDuck84` with `XUID match`.
- ProxyFingerprint report `20260922T045726.014738Z_f5c3fe0b06ede822_unlabeled.json` completed after `player_quit`, recorded Android metadata, `ClientData.PlayFabId` as absent, and only `non_public_peer`.
- The capture recorded downstream compression threshold `256` and upstream threshold `1`. This is documented as an existing local NetworkSettings boundary and was not changed by this test.
- No public server, raw capture, credential, or identifier was added to the repository.

## Automated validation

- The stamped runtime had already passed the nested gophertunnel tests, root tests, low-memory quality gate, and build before the live session.
- This commit is documentation-only and has no runtime impact.

## Live validation status

Owner-controlled Android local loopback passed. Login, resource-pack exchange, spawn, bidirectional gameplay, and clean quit were observed.

## Known limitations and follow-up

This evidence does not establish compatibility with or enforcement behavior of any public server. The separate authenticated upstream session and the NetworkSettings threshold mismatch remain documented compatibility boundaries.
