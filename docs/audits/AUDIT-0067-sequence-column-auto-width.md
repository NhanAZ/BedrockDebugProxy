# AUDIT-0067 - Size the event sequence column from visible values

- Commit subject: `fix: size event sequence column dynamically`
- Change type: `viewer`
- Runtime impact: `no`

## Intent

Keep full sequence values visible in the event stream when a capture reaches five or more digits.

## Scope

Derive the first table column width from the longest sequence on the current page. Update the viewer asset contract, changelog, and fix history. No capture, protocol, transport, or server behavior changes.

## Evidence and provenance

The reported screenshot showed values such as `#934...` because the fixed 62-pixel desktop and 52-pixel narrow column widths clipped longer sequence text. The change preserves the existing table layout and only adjusts the visible Seq column width.

## Validation

- Automated checks: `git diff --check`, the project quality gate, viewer HTTP tests, JavaScript syntax validation, and a rebuilt binary.
- Live validation: browser smoke on a clean committed viewer using the Hive capture `session-20260920T135412Z`; no new Hive connection is required or authorized.
- Capture or report: the local Hive capture remains unchanged and sensitive.

## Limitations and follow-up

The width is calculated from the currently visible page, so it can change when navigating between pages. It has a small safety allowance for the hash prefix and cell padding.
