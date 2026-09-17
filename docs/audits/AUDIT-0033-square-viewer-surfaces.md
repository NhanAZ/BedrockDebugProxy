# AUDIT-0033 - Make capture viewer surfaces square

- Commit subject: `style: make viewer corners square`
- Change type: `documentation | runtime | protocol | capture | tooling | process`
- Runtime impact: `no`

## Intent

Make the capture viewer use square corners throughout the interface as requested. This is a visual-only change and does not alter packet forwarding, capture records, protocol decoding, or resource-pack handling.

## Scope

- Remove corner rounding from the logo, status indicator, filter controls, search field, toggle controls, panels, pager buttons, and metadata cards in `internal/viewer/assets/app.css`.
- Update the user-facing changelog and fix-history index.
- Preserve all layout, typography, color, interaction, accessibility, and loopback-serving behavior.

## Evidence and provenance

- The pre-change stylesheet contained nine independent `border-radius` declarations across the viewer surfaces.
- The requested visual direction is a square, non-rounded interface.
- No external source or third-party code is introduced.

## Validation

- Automated checks: `tools/format.ps1`, `go test ./internal/viewer`, `tools/quality.ps1`, `git diff --check`, and a stylesheet assertion found zero `border-radius` declarations.
- Live validation: `not_observed`
- Capture or report: `none`

## Limitations and follow-up

The viewer still uses shadows, gradients, and borders for hierarchy, but no surface has rounded corners. No runtime impact.
