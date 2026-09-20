# Audit 0073 - Assert local ClientData parity

## Intent

Extend the existing loopback proxy regression fixture so it detects loss of stable downstream `ClientData` fields, not only skin and cape bytes.

## Scope

- Updated `internal/proxy/proxy_test.go`.
- The fixture now supplies and checks deterministic `DeviceOS`, `GameVersion`, `LanguageCode`, `DeviceID`, `SelfSignedID`, `PlayFabID`, input modes, skin, and cape values at the unauthenticated local upstream boundary.
- Runtime forwarding code, packet bytes, capture schema, and public-server behavior are unchanged.

## Evidence and provenance

The fields are part of the pinned in-tree gophertunnel `login.ClientData` contract. The test exercises the same `FetchResourcePacks` to upstream `minecraft.Dialer` path used by the local identity-parity fix. Values are synthetic and contain no account or device data.

## Automated validation

- `gofmt -w internal/proxy/proxy_test.go`
- `go test ./internal/proxy -run TestRunnerForwardsBidirectionalPacketsAndCapturesSession -count=1 -v`
- `git diff --check`

## Live validation

No live or public server was contacted. Existing local PocketMine loopback evidence remains the runtime validation boundary.

## Known limitations and follow-up

The assertion covers the unauthenticated local path. Authenticated login still obtains identity and signing material from the upstream Xbox authentication chain, so a real vanilla-client validation remains separate and is not inferred from this test.
