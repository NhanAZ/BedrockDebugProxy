# Audit 0158 - Record direct Android baseline

## Intent

Record the direct Android-to-PocketMine baseline used to interpret the preceding proxied session.

## Scope

- Add the direct local report and account-scoped claim comparison to [`docs/research/local-identity-parity.md`](../research/local-identity-parity.md).
- No runtime, protocol, capture schema, or viewer code changed.

## Evidence and provenance

- Direct report `20260922T051215.790689Z_a81b5174db46135f_unlabeled.json` completed after `player_quit` with 1,672 packets, Android metadata, `ClientData.PlayFabId` absent, and only `non_public_peer`.
- PocketMine's direct log identified the expected client username, authenticated Xbox Live, reached spawn, and closed cleanly.
- The comparison used the proxied report `20260922T045726.014738Z_f5c3fe0b06ede822_unlabeled.json`. Shared selected Android metadata matched. The proxy report contained `IsPartyLeader` and `PartyId` claim keys that the direct report omitted, but the authenticated account identities were different, so the difference is account-scoped and not attributable to the proxy.
- No public server, raw capture, credential, or identifier was added to the repository.

## Automated validation

- This commit is documentation-only and has no runtime impact.
- The preceding runtime revision passed the project quality gate and the Android local loopback.

## Live validation status

Direct Android local baseline passed. It is suitable for transport and protocol comparison, but not for same-account identity parity.

## Known limitations and follow-up

A same-account authenticated upstream is required before treating account-scoped ClientData claim differences as proxy-induced. Public-server behavior remains untested and out of scope.
