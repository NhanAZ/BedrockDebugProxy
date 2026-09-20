# AUDIT-0060 - Minecraft viewer presentation

- Commit subject: `style: finish viewer presentation`
- Change type: `tooling`
- Runtime impact: `no`

## Intent

Finish the capture viewer presentation using the supplied Minecraft formatting-code palette, responsive layout, page scrollbar styling, and pixel SVG toggle glyphs.

## Scope

Replace custom CSS colors with the supplied Minecraft foreground and background values, remove RGB/HSL alpha blends and gradients, use opaque palette surfaces, remove decorative shadows, use compact square scrollbars that match the palette for inner controls and the document root, reflow the viewer for narrow screens, bundle the maintainer-supplied pixel SVG glyphs for dropdown and Blob bytes toggles, and update HTML and manifest theme metadata to `#000000`. Capture decoding and evidence retention are unchanged.

## Evidence and provenance

The reported palette table defines the allowed foreground and background values. The previous stylesheet contained custom navy and pastel literals, alpha-blended RGB overlays, gradient backgrounds, and CSS-drawn toggle geometry. The new stylesheet uses palette-backed variables only, with bright text colors on dark surfaces to preserve readability, and serves the supplied SVG glyphs from the same embedded viewer bundle.

## Validation

- Automated checks: `go test ./internal/viewer`, `node --check internal/viewer/assets/app.js`, `git diff --check`, and `./tools/quality.ps1` passed before commit. The viewer asset test rejects non-palette color literals and gradient or RGB/HSL functions, checks compact scrollbar declarations, verifies the supplied SVG glyph responses and CSS references, and verifies representative text/background contrast ratios of at least 4.5.
- Live validation: passed on the clean build at `http://127.0.0.1:65227/2bf42f8ca0756b9eb9a2f5c5f47988c9/`. Browser smoke covered the overview, filters, expanded Packet menu with the supplied pixel glyph, event rows, selected-row state, matching page scrollbar, and panel controls at the tablet-style viewport. The asset test covers both SVG glyph responses and the phone breakpoint rules at 640px.
- Capture or report: `none`

## Limitations and follow-up

Supplied raster favicon assets are retained unchanged because this change scopes the palette contract to the HTML and CSS viewer presentation. If the asset artwork must also be recolored to the palette, it requires a separate image-asset decision and visual review.
