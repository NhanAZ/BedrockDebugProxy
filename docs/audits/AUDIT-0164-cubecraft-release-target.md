# Audit AUDIT-0164 - Restore CubeCraft as a maintainer-risk-gated release target

- Commit subject: `Restore CubeCraft as a maintainer-risk-gated release target`
- Intent: include CubeCraft in the private release-validation process without claiming that CubeCraft permits proxy use or exposing private server discussions.
- Runtime impact: none. The product, packet forwarding path, capture schema, and authentication behavior are unchanged.

## Scope

- Add CubeCraft as the sixth release-matrix label and direct endpoint reference.
- Require a maintainer `risk_acknowledged=pass` check before a CubeCraft report can pass.
- Preserve the public policy warning and prohibit policy-evasion behavior, private discussion, and authorization claims in reports.
- Synchronize the agent guide, validation docs, release workflow, changelog, fix history, and readiness checker.

## Evidence and provenance

- [CubeCraft Rules](https://help.cubecraft.net/en/article/cubecraft-rules-1403lij/) reviewed 2026-09-22. The current page states that proxies in Minecraft Bedrock are not allowed.
- [Allowed Mods and Clients](https://www.cubecraft.net/threads/allowed-mods-and-clients.228596/) reviewed 2026-09-22. The published policy also states that Bedrock proxies are not allowed.
- No private server discussion, permission claim, account credential, or capture is stored in the repository.

## Automated validation

- `git diff --check`
- PowerShell syntax and project quality gate to be run before commit.
- Documentation review of the six-server references and the readiness-checker label list.

## Live validation status

Not performed. No CubeCraft connection was initiated for this documentation and release-workflow change.

## Known limitations and follow-up

- The readiness checker can verify only the sanitized report check. The maintainer must make the risk decision before any live session.
- A passing risk acknowledgement does not imply a policy exception or guarantee that CubeCraft will accept the connection.
