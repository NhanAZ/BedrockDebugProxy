# AUDIT-0036 - Themed filter dropdowns

- Commit subject: `style: add themed filter dropdowns`
- Change type: `runtime`
- Runtime impact: `no - viewer presentation and interaction only; proxy forwarding and capture behavior are unchanged`

## Intent

Replace browser-native filter option popups with controls that use the capture viewer's dark, square, normal-weight visual language. Native popup styling is controlled by the browser or operating system and cannot be made consistent with the viewer stylesheet.

## Scope

- Add custom listbox controls for Direction, Packet, Kind, and Channel.
- Preserve each native select as the synchronized value source used by existing filter queries and change listeners.
- Support pointer selection, outside-click closing, focus styling, and Arrow, Home, End, Enter, Space, and Escape keyboard interaction.
- Add the themed trigger and option styles to the embedded viewer stylesheet and cover their presence in the viewer asset test.
- Intentionally unchanged: packet forwarding, capture storage, protocol decoding, resource-pack handling, filter query parameters, and live polling.

## Evidence and provenance

- The supplied viewer screenshot shows the closed controls styled by the project and the opened native popup rendered with unrelated browser styling.
- Browser-native option popup appearance is not portable or fully controllable with CSS. The custom listbox is a local presentation layer and does not alter captured data.
- No external code, font, dependency, or asset is introduced.

## Validation

- Automated checks: pass - `node --check internal/viewer/assets/app.js`, `go test ./internal/viewer`, `tools/format.ps1`, `tools/quality.ps1`, and `git diff --check` all passed. The quality gate also completed the full test suite, build, static analysis, vulnerability scan, and third-party license inventory.
- Live validation: pass - local Chrome smoke check opened the Direction listbox, confirmed the themed dark option panel, selected an option with the pointer, and verified that the synchronized filter value updated.
- Capture or report: `none`

## Limitations and follow-up

Browser viewport size, zoom, and local font availability still affect the exact dimensions and glyph rasterization. A browser smoke check should confirm pointer and keyboard behavior on the target machine.
