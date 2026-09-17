# AUDIT-0039 - Single-column viewer workspace

- Commit subject: `style: stack viewer workspace panels`
- Change type: `runtime`
- Runtime impact: `no - viewer presentation only; proxy forwarding and capture behavior are unchanged`

## Intent

Give the local capture viewer one readable evidence column instead of placing the event stream and inspection panels side by side.

## Scope

- Stack the event stream and inspector as full-width rows in a scrollable workspace.
- Keep decoded details above raw evidence within the inspector.
- Add a regression contract for the single-column layout and reject the previous two-column tracks.
- Update the user-facing changelog and fix-history index.

## Evidence and provenance

- The pre-change stylesheet defined a two-column workspace with separate minimum widths and repeated that layout at the narrow viewport breakpoint.
- The change only affects `internal/viewer/assets/app.css` and its presentation test. No capture data, HTTP route, packet handling, or font asset changes are included.

## Validation

- Automated checks: pass - `tools/format.ps1`, `go test ./internal/viewer`, `tools/quality.ps1`, and `git diff --check` all passed. The quality gate also completed the full test suite, build, static analysis, vulnerability scan, and third-party license inventory.
- Live validation: pass - exact binary revision `44b3e0e7f227316f2c49d71b64c6f34d8e0b3d6e` served closed capture `session-20260917T034859Z` in Chrome. The event panel and inspector had the same full width and different vertical positions, with decoded details above raw evidence. Workspace overflow was `auto`, and the direction filter interaction remained functional.
- Capture or report: `none`

## Limitations and follow-up

The workspace uses an internal scrollbar when the stacked panels exceed the viewport. Panel sizing and scrollbar placement can vary with browser zoom, operating-system scaling, and viewport dimensions.
