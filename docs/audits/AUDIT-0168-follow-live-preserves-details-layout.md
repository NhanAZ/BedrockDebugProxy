# Audit AUDIT-0168 - Preserve details layout during packet-heavy live follow

## Intent

Keep Decoded view - Event details usable while Follow live handles a high-rate capture.

## Scope

- Keep the current details and blob preview visible while the newest live event request is pending.
- Restore the page scroll position after the replacement details render.
- Extend viewer regression coverage and update the user-facing history.

No capture format, packet forwarding, filtering semantics, or HTTP API behavior changed.

## Evidence and provenance

The current live capture exposed the remaining failure after table-only scrolling was added. Each poll selected a new event and immediately replaced the details tree with a one-line loading message. At packet-heavy rates, that repeated collapse and expansion changed the page layout and interrupted manual scrolling.

## Automated validation

- Viewer asset coverage requires `preserveDetails`, `pageScrollTop`, and `restorePageScroll`.
- `node --check internal/viewer/assets/app.js` passed.
- `go test ./internal/viewer` passed.
- `tools/quality.ps1 -LowMemory` passed, including formatting, tests, build, static analysis, vulnerability scan, tracked-capture checks, and whitespace checks.
- `git diff --check` passed.

## Live validation

Browser smoke validation used a local live capture that appended 200 events at 250 ms intervals, alternating compact and nested decoded data. With Follow live enabled and Decoded view in view, the selected row advanced from event 153 to event 200 while the outer `scrollY` remained `576.8`, the Details panel top remained `107.6`, and its height remained `536.8`. The Details content stayed rendered as `Sequence#200`, and the browser reported no console errors.

No public-server or real-client validation is required for this viewer-only change.

## Known limitations and follow-up

Follow live may display the previous event's details briefly while the newest event request is in flight. This is intentional to preserve the operator's reading context.
