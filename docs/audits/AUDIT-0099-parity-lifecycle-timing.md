# Audit 0099 - Compare parity lifecycle timing

## Intent

Make the owner-controlled VanillaParity comparison useful for the lifecycle timing already retained in each local session report.

## Scope

- Add a versioned `lifecycle_timing` section to ProxyFingerprint comparisons.
- Report duration, ordered milestone prefix alignment, occurrence counts, and elapsed microseconds without changing captured packets or session reports.
- Document that timing is an observation aid and not evidence of a third-party enforcement rule.

## Evidence and provenance

ProxyFingerprint `SessionRecord` already stores an ordered lifecycle list with monotonic elapsed microseconds and a total duration. Its `ComparisonEngine` previously ignored those fields. The focused implementation is in owner-controlled PocketMine-MP commit [`323aa03de`](https://github.com/NhanAZ/PocketMine-MP/commit/323aa03de).

## Automated validation

- ProxyFingerprint self-test passed.
- PHP syntax checks passed for the changed PHP files.
- `git diff --check` passed in both repositories.
- BedrockDebugProxy runtime code and capture schema were not changed.

## Live validation

Not applicable. This change affects only local comparison output and does not contact any server.

## Known limitations and follow-up

Elapsed timing is host- and workload-dependent. It must be compared across repeated labeled local sessions and must not be treated as proof of a proxy implementation or a server policy.
