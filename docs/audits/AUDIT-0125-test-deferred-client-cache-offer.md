# Audit AUDIT-0125 - Test deferred client-cache offer ordering

- Commit subject: `Test deferred client-cache offer ordering`
- Change type: `test | documentation`
- Runtime impact: `no`

## Intent

Make the client-cache parity fix fail closed in tests if a future change sends
`ResourcePacksInfo` before the downstream `ClientCacheStatus` has been
observed, or if the observed boolean is not retained.

## Scope

- Add a focused gophertunnel state-machine test for both cache capability
  values.
- Verify the listener emits `PlayStatus` first, defers the resource-pack offer,
  then emits `ResourcePacksInfo` only after `ClientCacheStatus`.
- Record the test coverage in the fix-history index.

## Evidence and provenance

The test exercises the project-specific listener mode documented in
`docs/decisions/0011-forward-client-cache-capability.md` and the existing
packet handlers in `minecraft/conn.go`. It does not alter packet serialization
or production behavior.

## Validation

- Focused nested test: `go test ./minecraft -run '^TestDeferredResourcePackOfferWaitsForClientCacheStatus$' -count=1`.
- Full quality gate: `tools/quality.ps1 -LowMemory` passed before commit.
- Live validation: not applicable. This commit changes tests and documentation
  only.

## Limitations and follow-up

The test covers login ordering and both boolean values, not cache blob or miss
response payloads. Those still require a controlled local server fixture.
