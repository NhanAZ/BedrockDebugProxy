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
- BedrockPackArchive commit `731203abc1c49403493d4decccfd5b6be93bb1c8` adds the optional failed-probe event to its synthetic transferred-hop fixture. `TestImportCapturePreservesRootManifestAndPackMetadata` passed against the stamped BedrockDebugProxy binary at producer commit `71eb362293e01abe2710837a5e1c7a8995aa388b`. The binary reports Bedrock 1.26.62, protocol 2225, and SHA-256 `35F03997C1D6A77B9A147B076B65E7C8F025F28C9DE6742AFE58DD0DB6016E5D`. The importer contract and production code remain unchanged.

## Live validation and limitations

An authenticated Enchanted acquisition used the exact stamped producer binary from commit `71eb362293e01abe2710837a5e1c7a8995aa388b`. Hub login, pack delivery, spawn, and canonical capture verification succeeded. Existing completed child paths were skipped. `Trending Games` and `Other Games` produced no further form or `Transfer` before the menu wait expired, so the acquisition correctly returned incomplete. `Steal a Mob` returned the same structural menu and the consumer stopped that cycle without enqueuing duplicate child routes. This verifies the current successful login path and the paired menu-cycle fix. It did not exercise a child `Transfer` or the failed HTTPS probe warning. The synthetic HTTP 503 test remains the direct evidence for that warning path. No local capture path or private server message is cited.

This change does not establish why a particular HTTPS probe fails. It does not change fallback behavior, retry policy, address probing, or accepted server routes. The synthetic test verifies the warning path, not public-server compatibility.

For live validation, use a known Enchanted child route through BedrockPackArchive and the exact producer binary. Observe the server `Transfer` and the resulting `transfer.route_probe` event. A successful HTTPS probe should preserve the existing HTTPS route. A failed probe should record the target, timeout, error, and RakNet fallback before the unchanged fallback is attempted. Retain and verify the closed local capture if the flow fails. This diagnostic-only change does not satisfy the six-server release gate.
