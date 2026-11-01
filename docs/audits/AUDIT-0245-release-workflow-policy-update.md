# AUDIT-0245 - Update release and protocol-update stewardship

- Commit subject: `Update release workflow stewardship`
- Change type: `process`
- Runtime impact: `no`

## Intent

Update the release workflow requested by the maintainer: agents own the
Golem-assisted local checks and protocol-client release matrix, release
sessions must decrypt every encrypted offered resource pack, the agent fixes
in-scope release failures before asking for help, CubeCraft does not require a
separate repeated rejection/kick/ban acknowledgement, and post-release work
includes the private maintainer handoff, recovery synchronization, and local
artifact cleanup. Require protocol updates to be paired with a private
BedrockPackArchive protocol-baseline and compatibility update.

Correct the interpretation of a leading `--force`. It is the one-prompt
project-workflow override already defined by `AGENTS.md` and ADR 0012. It is
not a release command flag, does not create a `-Force` readiness-checker mode,
and does not bypass technical validation or server policy.

## Scope

Update the active agent guide, contribution guidance, README, changelog,
release and validation procedures, server-target policy, CLI release decision,
protocol-update procedure, report-generation and readiness tooling, and the
focused readiness tests. Preserve the existing six-server matrix, agent-
controlled protocol-client evidence, Golem's supplemental local role, the
Enchanted transfer gate, capture provenance, and sensitive-data boundaries.

## Evidence and provenance

The leading-prefix semantics are established by `AGENTS.md`,
`docs/decisions/0012-force-prompt-project-workflow-override.md`, and
`docs/audits/AUDIT-0179-force-prefix-project-workflow-override.md`. The
resource-pack process uses the existing live-connection `ContentKey` boundary
and records both encrypted and derived pack counts in the sanitized report.
The private maintainer README is a post-publication step and was not treated as
read or completed because this change does not publish a release.

## Validation

- Automated checks: `git diff --check`; focused Go validation tests covering
  CLI and retail session modes, CubeCraft without a risk acknowledgement,
  encrypted-pack success, missing decryption, missing authentication or
  stability, missing transfer, failed automatic checks, failed results, wrong
  revisions, packs-only, and ambiguous sessions.
- Live validation: `not_observed` for this documentation and workflow change.
- Capture or report: no new capture or public release report.

## Limitations and follow-up

The release candidate remains subject to the six-server live gate. This policy
change does not claim that CubeCraft will accept the connection and does not
replace the normal read-only capture or technical checks. The paired
BedrockPackArchive update is required when a future protocol update occurs; it
is not performed by this release-workflow documentation change.

## Security and privacy

The workflow retains encrypted archives until post-release verification and
then removes local captures and derived decrypted artifacts using explicit
paths. It does not publish private handoff content, credentials, content keys,
raw captures, or decrypted third-party packs.
