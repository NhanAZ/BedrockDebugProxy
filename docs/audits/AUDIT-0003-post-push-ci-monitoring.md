# AUDIT-0003 - Require post-push CI monitoring

- Commit subject: `Require post-push CI monitoring`
- Change type: `documentation | process`
- Runtime impact: `no`

## Intent

Make the post-push workflow explicit so a local quality pass cannot be mistaken for a completed GitHub Actions
run. The agent must inspect the exact pushed commit, wait for every required job, and fix any failure before
declaring the change complete.

## Scope

Update `AGENTS.md` and `CONTRIBUTING.md` with the post-push CI loop. The loop requires checking the Actions run for
the exact commit, waiting for both jobs in `quality.yml`, reading failure logs, and using a new audited commit for
any correction.

No runtime code, protocol behavior, capture schema, CLI behavior, dependency, or workflow YAML changes are included.

## Evidence and validation

- Before this policy change, GitHub Actions run 26 for commit `c9a7318a2f018e688b6d9553558946920cf8a6ae` completed successfully.
- Both `Quality gate` and `Windows compatibility` were observed as completed successfully in the repository's Actions UI.
- Automated validation for this documentation-only change: `git diff --check`.
- Live validation: `not_observed`; runtime behavior is unchanged.

## Limitations and follow-up

This policy depends on repository access and a reachable GitHub Actions service. If Actions is unavailable, record the
run as unavailable rather than treating local checks as proof of CI success. Every future push must repeat the exact
commit check.
