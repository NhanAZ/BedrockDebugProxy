# AUDIT-0050 - Inline viewer panel headings

- Commit subject: `style: inline viewer panel headings`
- Change type: `tooling`
- Runtime impact: `no`

## Intent

Reduce vertical space used by panel headings while preserving clear section and title text.

## Scope

Render Ordered stream - Events, Decoded view - Event details, and Raw evidence - Blob bytes on one line. Panel controls, event data, capture storage, and protocol behavior are unchanged.

## Evidence and provenance

The reported screenshots showed each panel heading as two stacked lines. An explicit separator span keeps the relationship readable without depending on generated CSS text.

## Validation

- Automated checks: `go test ./internal/viewer`, `node --check internal/viewer/assets/app.js`, `./tools/quality.ps1`, and `git diff --check` passed before commit.
- Live validation: passed on the clean build at `[local viewer URL]`. Browser smoke confirmed Ordered stream - Events and Decoded view - Event details render on one line, the accessibility tree exposes Raw evidence - Blob bytes in the same inline structure, and the pager, Copy JSON, and Blob toggle controls remain aligned and usable. No browser errors or warnings were visible during smoke test.
- Capture or report: `none`

## Limitations and follow-up

The minimum page width remains unchanged. At the supported minimum width, long localized or generated labels may still ellipsize according to the existing panel controls.
