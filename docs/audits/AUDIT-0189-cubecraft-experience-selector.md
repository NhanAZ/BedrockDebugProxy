# Audit AUDIT-0189 - Document CubeCraft Experience selector

- Commit subject: `Document CubeCraft Experience selector`
- Change type: `release documentation correction`
- Runtime impact: `no`

## Intent

Make the release instructions use the existing `experience:` selector for CubeCraft while retaining its Featured server category and endpoint reference.

## Scope

- Update the release target matrix, release command guidance, validation matrix, and agent guide to use `experience:CubeCraft`.
- Clarify that `experience:` resolves both Featured and Creator Experiences and that the category column does not determine selector syntax.
- Keep the CubeCraft risk acknowledgement and read-only validation requirements unchanged.
- Correct the earlier fix-history wording to identify the endpoint as a reference, not the release selector.
- No source, CLI behavior, capture, validation report, or report revision changed.

## Evidence and provenance

`internal/experience/target.go` routes `experience:<name>` through Minecraft services and its resolver error names both featured and creator experiences. `internal/experience/target_test.go` accepts `EXPERIENCE:CubeCraft`. The README also documents that the selector resolves both categories. No external source or private session data was used.

## Validation

- `git diff --check` passed.
- Searched the Markdown documentation for CubeCraft endpoint instructions. The active release guidance now uses `experience:CubeCraft`; the endpoint appears only as a reference. Historical audit records remain unchanged.
- No tests were run because the change is documentation-only and does not alter runtime behavior.

## Live validation status

Not applicable. No server connection was initiated, and existing report contents and `tested_revision` values were not changed.

## Limitations and follow-up

The endpoint reference can change as Minecraft services update routing. The selector resolves the current destination at runtime. CubeCraft's existing maintainer risk gate still applies before any connection.
