# AUDIT-0057 - Expand packet filter menu labels

- Commit subject: `style: expand packet filter menu labels`
- Change type: `tooling`
- Runtime impact: `no`

## Intent

Keep packet names readable when the custom Packet filter menu is open.

## Scope

Use the menu content's intrinsic width instead of a fixed 220px minimum, keep the trigger width as the floor, and cap the menu at 480px or the available viewport width. Filter values and query behavior are unchanged.

## Evidence and provenance

The reported menu truncated `AvailableActorIdentifiers` with an ellipsis. The custom dropdown already renders each option as text, so the narrow menu width was the presentation root cause rather than missing packet data.

## Validation

- Automated checks: `go test ./internal/viewer`, `node --check internal/viewer/assets/app.js`, `git diff --check`, and `./tools/quality.ps1` passed before commit.
- Live validation: passed on the clean build at `http://127.0.0.1:53539/55da6a4938636f1a4edbc30be17edc7b/`. Browser smoke confirmed the expanded Packet menu exposes the full `AvailableActorIdentifiers` label without an ellipsis and remains bounded by the viewport-safe width. No browser errors or warnings were visible during smoke test.
- Capture or report: `none`

## Limitations and follow-up

Names longer than the viewport-safe 480px cap can still be clipped by the menu boundary. The cap prevents the dropdown from creating page-level horizontal overflow.
