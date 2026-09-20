# AUDIT-0069 - Preserve filter menu scroll during live refresh

- Commit subject: `fix: preserve filter menu scroll during live refresh`
- Change type: `viewer`
- Runtime impact: `no`

## Intent

Keep an open Packet or Kind filter at the operator's current scroll position while a live viewer refreshes its option list.

## Scope

Capture each custom menu's `scrollTop` before live option synchronization and restore it after rebuilding the option nodes. The filter values, event queries, capture data, and polling cadence are unchanged.

## Evidence and provenance

The live viewer calls `loadSession` during its one-second poll. That path calls `syncOptions` for every filter, which rebuilds the custom menu with `replaceChildren`. Rebuilding the scrollable menu reset an open long list to the first option on every poll.

## Validation

- Automated checks: `git diff --check`, the project quality gate, viewer HTTP tests, JavaScript syntax validation, and a rebuilt binary.
- Live validation: browser smoke must keep a long Packet or Kind menu open while a live capture poll completes and confirm that the visible option range does not jump to the first option.
- Capture or report: no capture data is changed or added.

## Limitations and follow-up

If the refreshed option list becomes shorter than the previous scroll position, the browser clamps `scrollTop` to the new maximum. This is expected and preserves the nearest valid position.
