# Audit 0153 - Preserve omitted authenticated PlayFab ID

## Intent

Remove a proxy-created `ClientData.PlayFabId` from authenticated upstream logins when the downstream vanilla client omitted the field.

## Scope

- Updated the in-tree gophertunnel dialer to generate a PlayFab fallback only for offline logins.
- Preserved the downstream authenticated ClientData value, including an empty `PlayFabId`, for authenticated dials.
- Added deterministic coverage for the authenticated empty-field case while retaining the offline fallback test.
- Updated the gophertunnel patch provenance, local identity research, fix history, and changelog.
- No token claims, signatures, device metadata, forwarded packet bytes, capture schema, or public-server behavior changed.

## Evidence and provenance

- The closed local comparison from 2026-09-22 at the preceding revision recorded a direct vanilla login with no `ClientData.PlayFabId` and a proxied authenticated login with a generated PlayFab value whose HMAC differed from the token `mid` claim.
- The in-tree `defaultClientData` call occurred before the authenticated `login.Encode` branch and generated that value whenever the field was empty.
- The change keeps the upstream authenticated Android metadata boundary intact. It does not copy or forge the signed identity.
- No credentials, raw captures, or external endpoints are included in this audit.

## Automated validation

- `go test ./...` passes from the repository root.
- `go test ./...` passes in `third_party/gophertunnel`.
- `git diff --check` and the project quality gate remain required before commit.

## Live validation status

No post-change real-client session has been run yet. A fresh owner-controlled authenticated loopback capture is required to verify the emitted login metadata after this change.

## Known limitations and follow-up

Authenticated upstream dials still use the dependency's Android device metadata and a separate signed upstream session. This change addresses only the synthetic PlayFab field and does not establish public-server compatibility or enforcement behavior.
