# AUDIT-0040 - Render pending decode status as gold

- Commit subject: `style: distinguish pending packet decodes`
- Change type: `tooling`
- Runtime impact: `no`

## Intent

Make the capture viewer distinguish intentionally preserved `not_yet_decoded` raw packet evidence from actual decode failures. The pending status now uses the requested gold `#FFAA00` presentation.

## Scope

Updated the embedded viewer stylesheet and row renderer, viewer asset regression checks, the user-facing changelog, and the fix-history index. Packet capture, decoding, filtering semantics, and the canonical capture schema are unchanged.

## Evidence and provenance

`internal/bedrock/observer.go` records `not_yet_decoded` for raw frames at `post_decrypt_decompress_frame`. `internal/proxy/proxy.go` records decoded views separately and emits decode or view errors through distinct events. The previous viewer styling treated every non-`decoded` status as red.

## Validation

- Automated checks: `go test ./internal/viewer`, `node --check internal/viewer/assets/app.js`, and `./tools/quality.ps1` passed. `git diff --check` passed.
- Live validation: `pass` via a clean local `view --no-open` build, browser rendering of raw pending-decode rows, event selection, and a console check with no errors or warnings.
- Capture or report: `none`

## Limitations and follow-up

The `Problems only` filter still includes pending-decode rows because they are review targets, although their color no longer implies failure. No runtime forwarding or capture behavior changed.
