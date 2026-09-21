# Audit AUDIT-0079 - Assert offline token identity claims

## Intent

Prevent regressions where authentication-bound identity fields reappear in the self-signed multiplayer token used by an unauthenticated upstream dial.

## Scope

- Added `TestOfflineEncodingOmitsAuthenticationBoundClaims` to the nested gophertunnel metadata tests.
- The test encodes and parses an offline login after `clearXBLIdentityData`, checking that XUID, title, and PlayFab claims remain empty while the display name and UUID survive.

## Evidence and provenance

The test covers the exact `clearXBLIdentityData` to `login.EncodeOffline` to `login.Parse` path implicated by the local ProxyFingerprint `auth_client_playfab_mismatch` report.

## Automated validation

- `gofmt` passed for `third_party/gophertunnel/minecraft/dial_metadata_test.go`.
- Nested gophertunnel tests for both identity-clearing cases passed.
- The project quality gate is required before commit.

## Live validation

No runtime or live-server behavior changed. The test uses synthetic identity data only.

## Known limitations and follow-up

This regression covers unauthenticated upstream token construction. It does not make an authenticated two-session proxy equivalent to a single official client or establish public-server policy behavior.
