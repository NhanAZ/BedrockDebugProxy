# Audit 0184 - Stabilize loopback release tests

- Commit subject: `Stabilize loopback release tests`
- Change type: `test maintenance`
- Runtime impact: `no`

## Intent

Remove nondeterminism from two loopback proxy tests that failed the first release-candidate CI run.

## Scope

- Keep the fake upstream connection alive after `StartGame` until the proxy closes it.
- Read the shared test log buffer under its mutex.
- Assert the identity of the first downstream request observed after `Transfer`, without replacing it with a later reconnect packet.
- No product code, capture behavior, protocol handling, or user documentation changes.

## Evidence and provenance

GitHub Actions run `35945153597` for candidate `1b8619945cd83cc641f8d066f4ffe630686d79dc` failed in `TestRunnerFollowsTransferAndAcceptsNextHop` with a test-only `strings.Builder` data race and an assertion overwritten by a later reconnect packet. The Windows job also timed out in `TestRunnerPreservesMixedResourcePackOfferOrder`; its fake upstream closed immediately after `StartGame` returned. Both focused tests passed on the local Windows environment before this change.

## Validation

- Automated checks: the repeated focused tests and focused race test passed. `tools/quality.ps1 -Race` passed, including the full race-detected Go suite, build, static analysis with 0 issues, vulnerability scan with 0 reachable vulnerabilities, viewer checks, module and license checks, and whitespace validation. The scan also reported 4 findings in imported packages and 2 in required modules that the code does not appear to call.
- CI: pending for the commit containing this record.
- Live validation: no runtime impact; existing release live-validation requirements remain separate.

## Limitations and follow-up

The test changes preserve the assertions for the first post-transfer packet and for mixed resource-pack offer order. The full release candidate still requires a green race-enabled CI run and the required live server matrix.
