# AUDIT-0178 - Record Hive appeal outcome

- Commit subject: `Record Hive appeal outcome`
- Change type: `documentation`
- Runtime impact: `no`

## Intent

Update the Hive incident record with the maintainer-reported appeal result so it no longer implies that the account remains punished or that the earlier status itself prevents a routine release baseline. Keep the technical cause and public server policy clearly separate from the appeal outcome.

## Scope

Update `docs/research/hive-unfair-advantage-2026-09.md` and its `FIX-0051` entry in `docs/fix-history.md`. No runtime behavior, public Hive policy statement, CLI, capture, or validation-report schema changes.

## Evidence and provenance

The maintainer reports that Hive characterized the punishment as a false positive and lifted it after review. The provided support-thread export confirms that the Appeals & Security Team reviewed the case and lifted the punishment, but does not identify a detector or technical trigger. The export, private correspondence, contact details, and punishment ID are not stored in the repository. The outcome is recorded as account-level context, not proof of a technical root cause or a policy change.

## Validation

- Automated checks: `tools/quality.ps1 -LowMemory` passed. This included all Go tests, build, static analysis with 0 issues, vulnerability scan with 0 reachable vulnerabilities, tracked-artifact checks, and whitespace validation. The scan noted 4 imported-package and 2 required-module findings that the code does not appear to call.
- Live validation: `not_observed`
- Capture or report: none; the exact-revision Hive release baseline remains separate follow-up work.

## Limitations and follow-up

The original enforcement signal remains unknown. A future Hive session must follow the current read-only validation workflow and must not attempt to probe enforcement. No runtime impact.
