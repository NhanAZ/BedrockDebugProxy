# AUDIT-0273 - Bound connection timeouts across transfer hops

- Commit subject: `Apply connection timeout to each transfer hop`
- Change type: `runtime`
- Runtime impact: `yes`

## Intent

Stop a followed server transfer from escaping the caller's configured connection deadline. Let pack acquisition use a shorter default wait while retaining an explicit longer bound for slow offers.

## Scope

Apply Go Golem's `ConnectTimeout` independently to each transfer dial, spawn, and resource-pack observer. Add `run --connect-timeout` for each proxy upstream login and resource-pack exchange. Correct the timeout disconnect wording so it reports the configured proxy deadline rather than asserting that the destination server did not respond. The paired BedrockPackArchive `fetch --connect-timeout` option and transferred-hop synthetic fixture are maintained in the consumer repository.

## Evidence and provenance

Before this change, `followTransfer` used the long-lived session context after the initial connection's shorter context had ended. The BedrockPackArchive Enchanted reconnect budget was fixed at 16 minutes to outlast the proxy's fixed 15-minute upstream timeout. Existing Galaxite research records active resource-pack chunk delivery beyond five minutes, so the proxy's default remains 15 minutes while acquisition has a 3-minute default and an explicit 15-minute maximum.

## Validation

- Automated checks: `tools/quality.ps1` passed after the final BedrockDebugProxy diff, including formatting, the full test suite, build, static analysis, vulnerability scanning, and whitespace validation. BedrockPackArchive `go test -count=1 ./...`, `go vet ./...`, `go build ./...`, and `git diff --check` passed. Its exact stamped-producer fixture remains pending until the producer commit is built.
- Live validation: not observed on the exact committed revision. An earlier working-tree Enchanted observation is not retained as permanent evidence because its local capture included derived decrypted pack content.
- Capture or report: none.

## Limitations and follow-up

The deadline is a total per-hop bound, not a progress-resetting idle timer. A slow active pack transfer may need a larger configured value. A timeout alone does not prove that the upstream server failed. Run the full authenticated six-server validation matrix before publishing a new official release.
