# Audit AUDIT-0077 - Keep authenticated identity fields out of offline upstream logins

## Intent

Correct the unauthenticated upstream login path so it does not emit PlayFab identity claims copied from a downstream authenticated client while retaining the local player's display name and UUID.

## Scope

- Updated `third_party/gophertunnel/minecraft/dial.go` to clear `PlayFabID` and `PlayFabTitleID` alongside XUID and legacy title ID before `EncodeOffline`.
- Added a focused nested-module regression test for clearing authentication-bound fields without changing offline identity fields.
- Updated the fix history, local identity research, changelog, and third-party patch provenance.

## Evidence and provenance

- The local 1.26.51 vanilla-client loopback report showed a self-signed upstream token with a `mid` claim derived from forwarded identity data while `ClientData.PlayFabId` remained the client value.
- The direct vanilla login shape did not include an authenticated `mid` claim in the corresponding offline path.
- `EncodeOffline` serializes `IdentityData.PlayFabID` and `IdentityData.PlayFabTitleID` into the multiplayer token, which made the omission in `clearXBLIdentityData` the verified local cause.

## Automated validation

- `gofmt` passed for changed Go files.
- Nested gophertunnel tests passed for the new identity-clearing regression and the existing 1.26.51 metadata regression.
- The focused proxy forwarding integration test passed.
- `tools/quality.ps1` passed, including Go tests, build, static analysis, vulnerability scan, viewer checks, tracked-artifact checks, and whitespace validation.
- `git diff --check` passed.

## Live validation

The change is limited to unauthenticated operator-controlled local upstream dials. No public server was contacted and no anti-cheat or authentication policy was bypassed.

## Known limitations and follow-up

This fixes a local identity-shape mismatch. It does not prove that public-server enforcement classifies a proxy in any particular way, and authenticated upstream parity still requires an authorized test server.
