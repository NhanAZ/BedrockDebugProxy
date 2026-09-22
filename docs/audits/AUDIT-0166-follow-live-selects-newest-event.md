# Audit AUDIT-0166 - Select the newest event during live follow

## Intent

Make Follow live identify the newest matching packet, not only the page that contains it.

## Scope

- Select the last row returned by each live tail refresh while Follow live is enabled.
- Load the selected event details and scroll its row into the table viewport.
- Avoid repeating the detail request when no newer sequence has arrived.
- Preserve the follow intent across the final refresh when the capture changes from open to closed.
- Update viewer validation, capture-viewer documentation, fix-history, and changelog text.

No capture format, packet forwarding, filtering semantics, or HTTP API behavior changed.

## Evidence and provenance

The live API already returns the newest matching page through its `tail=true` query. The defect was limited to the browser selection state after that response was rendered. The implementation uses the final row in the returned response and keeps the existing selected-event endpoint.

## Automated validation

- Focused viewer tests cover the embedded JavaScript hook and existing live tail endpoint behavior.
- `git diff --check` passed.
- `powershell -NoProfile -ExecutionPolicy Bypass -File tools/quality.ps1 -LowMemory` passed, including viewer JavaScript syntax validation, all tests, build, static analysis, and the vulnerability scan.

## Live validation

Pass via a clean committed build at `[local viewer URL]` using an open local capture. After appending sequence 5, Follow live selected row `#5`, loaded Details for sequence 5, and kept the row visible. The browser console had no errors or warnings. No public-server or real-client validation was required for this viewer-only change.

## Known limitations and follow-up

The selected event follows the newest row after active filters are applied. Turning Follow live off preserves the current page and selection so earlier events can be inspected.
