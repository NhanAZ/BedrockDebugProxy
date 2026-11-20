# AUDIT-0265 - Preserve transfer probe failure evidence

- Commit subject: `Record HTTPS transfer probe failures before fallback`
- Change type: `capture diagnostics`
- Runtime impact: `no transport-selection change`

## Intent

Preserve the reason an optional HTTPS NetherNet route probe did not select that transport before the existing RakNet fallback proceeds.

## Root cause and scope

`prepareTransferHop` treated the HTTPS probe as best effort and discarded its error. The proxy then followed its existing RakNet route without recording why HTTPS was unavailable. The cause can be transient or endpoint-specific. The available evidence does not justify changing the fallback, adding retries, or attributing the failure to a server.

The change records a `transfer.route_probe` warning containing the target, attempted transport, fallback transport, configured timeout, best-effort status, error message, and error type. A capture write failure is propagated. The capture schema version and transport selection are unchanged.

## Validation

- `go test -run '^TestPrepareTransferHopRecordsHTTPSProbeFailureBeforeRakNetFallback$' -count=1 ./internal/proxy` passed against a synthetic TLS endpoint returning HTTP 503.
- The test verifies that the error and timeout are captured, the capture remains closable, and the proxy retains its RakNet fallback.
- `tools/format.ps1` passed and formatted 479 Go source files. The focused diff remains limited to the intended source and documentation files.
- `tools/quality.ps1` passed, including all Go tests, build, static analysis with 0 issues, vulnerability scan with no reachable findings, tracked-capture checks, and whitespace checks.
- `git diff --check` passed in both repositories. BedrockPackArchive full tests, vet, and build passed.
- The BedrockPackArchive synthetic importer fixture was updated for the optional warning event. Its exact stamped-producer verification and commit cross-reference are pending and will be added after the paired consumer commit.

## Live validation and limitations

An earlier authenticated Enchanted Golem session reached spawn through `nethernet-https`, collected the Hub and child offers, and produced a closed, verified capture. It predates this diagnostic change, so it does not verify the new warning event. A fresh live session is needed to observe the changed binary. No real capture or endpoint is cited here.

This change does not establish why a particular HTTPS probe fails. It does not change fallback behavior, retry policy, address probing, or accepted server routes. The synthetic test verifies the warning path, not public-server compatibility.
