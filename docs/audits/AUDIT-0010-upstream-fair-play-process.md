# AUDIT-0010 - Document upstream fair-play contribution process

- Commit subject: `Document upstream fair-play contribution process`
- Change type: `documentation`
- Runtime impact: `no`

## Intent

Document a consent-first, evidence-driven process for helping third-party dependency maintainers when a reproducible upstream defect is discovered during BedrockDebugProxy work.

## Scope

Add the upstream fair-play and good-faith contribution policy to `AGENTS.md`. The policy covers investigation, maintainer approval, issue or pull request selection, privacy and security boundaries, license and contribution rules, and follow-up after submission. It does not change runtime behavior, dependencies, capture formats, validation gates, or external repository state.

## Evidence and provenance

The policy is based on the existing issue triage, security and privacy, dependency provenance, external coordination, and change-control requirements in `AGENTS.md`, together with the project maintainer's request. No upstream repository was contacted and no issue or pull request was created for this change.

## Validation

- Automated checks: `pass - git diff --check; documentation review`
- Live validation: `not_applicable`
- Capture or report: `none`

## Limitations and follow-up

The policy authorizes agents to prepare and, after explicit maintainer approval, carry out a narrowly scoped upstream contribution. It does not authorize automatic external communication or guarantee upstream acceptance. No runtime impact.
