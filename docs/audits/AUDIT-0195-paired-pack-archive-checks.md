# Audit AUDIT-0195 - Record paired resource-pack commit checks

- Commit subject: `Record paired resource-pack commit checks`
- Change type: `documentation`
- Runtime impact: `no`

## Intent

Record the exact BedrockDebugProxy and BedrockPackArchive commit pair and verify the results after both repositories were pushed.

## Scope

- Document the two commit hashes, verified remote refs, and CI status for the resource-pack manifest-selection change.
- Update AUDIT-0194 with the completed GitHub Actions result.
- No runtime, capture, or public interface changes.

## Evidence and provenance

- BedrockDebugProxy runtime commit: `bb0875e97fa60b7d249d53e8b1cffc61c372cc28`.
- BedrockPackArchive compatibility-test commit: `95c919474fb2dfa9549527a2c5f6389fdc8f165f`.
- Both commits were pushed to their existing `main` branches. The BedrockPackArchive repository remains private.
- BedrockDebugProxy GitHub Actions run [36143436642](https://github.com/NhanAZ/BedrockDebugProxy/actions/runs/36143436642) completed successfully. The Linux Quality gate, including race detection, and Windows compatibility tests and build both passed for the runtime commit.
- BedrockPackArchive has no configured GitHub Actions workflows. Its synthetic integration test, `go vet`, and build passed locally before its push.

## Validation

- Automated checks: `git diff --check` passed for this documentation-only change. The runtime commit's GitHub Actions checks passed as listed above. This follow-up commit's GitHub Actions result will be checked after push.
- Live validation: `not_observed` for the exact runtime revision. See AUDIT-0194 for the required The Hive and OneBlockOnline follow-up.
- Capture or report: `none`. This commit contains no capture data or pack contents.

## Limitations and follow-up

This record verifies source-control and automated checks only. It does not replace the pending exact-revision live validation documented in AUDIT-0194.
