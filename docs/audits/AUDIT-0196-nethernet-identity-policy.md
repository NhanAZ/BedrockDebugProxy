# AUDIT-0196 - Record NetherNet v1.0.23 identity policy

- Commit subject: `docs: record go-nethernet v1.0.23 behavior`
- Change type: `documentation`
- Runtime impact: `no`

## Intent

Document the behavior and evidence for the go-nethernet v1.0.23 dependency update in this pull request, including the unresolved difference between upstream behavior and the current Mojang guide.

## Scope

Update the Unreleased changelog and direct-dependency provenance, and add a focused research note. This commit changes no product code or dependency versions.

## Evidence and provenance

The note links the go-nethernet v1.0.23 tag and its identity tests, the immutable upstream change `b5719958958a1d211334984ba520c1eb16fab640`, and the Mojang guide at commit `c5a73c3b275f3c32366f0c2048643f80e6a13ad8`. The guide and upstream implementation disagree about temporal validation of server identity tokens. No private chat transcript or capture is included.

## Validation

- Automated checks: `tools/format.ps1` passed; `tools/quality.ps1` passed; `go test ./...` passed from the go-nethernet v1.0.23 module; `git diff --check` passed. GitHub Actions passed on the dependency commit `8d6fee87432f07437734eab64f7b0b0188e66905`; checks for this documentation commit remain pending until pushed.
- Live validation: `pending`
- Capture or report: `none`

## Limitations and follow-up

This documentation-only commit has no runtime impact. The dependency update in the preceding Dependabot commit changes runtime behavior and still requires a stamped build, a closed capture, and a machine-readable report for the exact tested revision before the pull request can be approved or merged.
