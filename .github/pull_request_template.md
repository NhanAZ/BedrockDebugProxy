## Summary

Describe the problem, root cause, and focused change.

## Scope

- What changed
- What deliberately did not change
- Sources or provenance used for protocol, compatibility, or adapted code

## Validation impact

- [ ] Documentation, comments, or formatting only. Live validation is not required.
- [ ] Runtime behavior may change. The Hive validation is required.
- [ ] A specific flow changed and was exercised directly.

## Evidence

- [ ] `tools/quality.ps1` passed
- [ ] GitHub Actions passed
- Tested revision
- The Hive report path or reason for exemption
- Flow-specific report path when required
- Remaining uncertainty or anomalies

## Review checklist

- [ ] The change is focused and avoids duplicate functionality
- [ ] Capture fidelity and sensitive-data boundaries are preserved
- [ ] Documentation matches the behavior
- [ ] Required validation reports match the current pull request head
