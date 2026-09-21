# Changelog

All notable changes to BedrockDebugProxy are documented in this single file. Each release keeps a concise, curated
summary here, while GitHub's compare view and the commit audit records provide the complete history and evidence.
Detailed rationale remains in [`docs/audits/`](docs/audits/).

## [Unreleased]

### Added

- Added a read-only local capture viewer with paginated event filters, decoded JSON details, bounded blob hex previews, loopback-only serving, and per-run access URLs. Open captures can be followed with `view --live`, and `run --web-viewer` starts the live browser view automatically without putting browser work on the forwarding path.

### Changed

- `session.negotiated` capture events now distinguish the protocol implementation version from the game version and device OS reported in each connection's `ClientData`, including explicit mismatch booleans for local parity investigations.
- `session.negotiated` capture events now include bounded compression and client-throttle settings for both connections, with an explicit match flag for per-session parity analysis.
- Updated the compiled Bedrock game-version metadata to 1.26.51 while retaining protocol 2193 and the reviewed 1.26.50 packet definitions. Mojang's 1.26.51 release reports no packet changes, so authenticated upstream login metadata now advertises 1.26.51 without changing forwarded packet bytes.
- Unauthenticated upstream dials now preserve the downstream client's identity instead of falling back to gophertunnel's default `Steve` identity. Xbox-authenticated dials continue to use the identity from the authentication chain.
- Unauthenticated upstream dials now remove forwarded XUID and PlayFab identity claims before creating a self-signed login, preventing authenticated identity data from being mixed into the offline token while retaining the downstream name and UUID.
- Terminating bridges now retain and reuse the complete decoded upstream `StartGame` packet when sending world-start data downstream, instead of replacing fields outside the compact `GameData` view with proxy defaults.
- Startup `JigsawStructureData` and `VoxelShapes` packets are now retained from the upstream login and sent once with their original decoded contents instead of being duplicated around the generated spawn sequence.
- Terminating bridges now retain and reuse the complete upstream `ResourcePackStack`, preserving its pack order, base game version, experiments, and stack flags instead of appending local exempted packs.
- Terminating bridges now preserve decoded upstream `ResourcePacksInfo` offer metadata while rebuilding downstream pack entries from archive-backed resources that the proxy can serve.
- Terminating bridges now retain and reuse raw `DimensionData` and `ItemRegistry` startup packets when the downstream protocol matches, avoiding wire differences caused by typed re-serialization of NBT maps.
- Terminating bridges now also retain and reuse raw `JigsawStructureData`, `StartGame`, and `VoxelShapes` startup packets when the downstream protocol matches, preserving the observed world-start bytes while keeping typed fallback behavior.
- The capture viewer retains the OFL-licensed fan-made font assets without contacting a font CDN. Visible text uses regular Monocraft and can resolve a separately installed Minecraft Ten before the embedded fallback.
- The capture viewer now uses one 12px normal-weight scale for all visible text, removes forced uppercase styling, and keeps labels in natural casing.
- The capture viewer now stacks the event stream, decoded details, and raw evidence in one scrollable column instead of splitting the workspace into two columns.
- The capture viewer now uses square corners throughout its interface, including status indicators, controls, panels, and metadata cards.
- The capture viewer now uses themed custom filter dropdowns with keyboard navigation instead of browser-native option popups, while keeping the original filter values and queries unchanged.
- The capture viewer now renders intentionally pending raw packet decodes in gold (`#FFAA00`) instead of using the error red reserved for decode failures.
- The capture viewer now uses only the Minecraft formatting-code palette, removes all gradients and blended color functions, and keeps viewer text on high-contrast dark surfaces.
- The capture viewer now uses compact four-pixel palette scrollbars without native arrow buttons in filter menus and the event stream.
- The capture viewer now gives the outer page scrollbar the same compact four-pixel palette styling as the inner scrollbars.
- The capture viewer now uses the supplied pixel SVG glyphs for dropdown and Blob bytes toggles instead of CSS-drawn chevrons.
- The capture viewer now replaces native JSON tree disclosure markers with the supplied pixel glyph instead of the browser triangle.
- The capture viewer now uses square, palette-colored custom checkboxes instead of native rounded checkbox controls.
- The capture viewer now reflows filters, overview metrics, event controls, and detail metadata for narrow screens without forcing a page-wide minimum width.
- The capture viewer now lets event details grow with their content instead of clipping or collapsing the panel behind an internal vertical scrollbar.
- The capture viewer now uses one page-level scrollbar so the header, filters, event stream, details, and raw evidence share the same scroll context.
- The capture viewer now keeps Blob bytes collapsed by default and provides an accessible toggle for expanding the raw hexdump when needed.
- The capture viewer now includes the supplied favicon, touch icon, and dark themed web manifest assets.
- The capture viewer header now uses the same supplied logo artwork as its favicon package.
- The capture viewer now places Events, Blobs, Duration, and Generator in one compact header row instead of a separate overview strip.
- The capture viewer header no longer displays the verification pill or capture identifier, leaving the compact overview as the only header metadata.
- The capture viewer now renders each panel eyebrow and title on one line with a hyphen separator to reduce heading height.
- The capture viewer header now keeps only the BedrockDebugProxy name and removes its decorative logo and secondary viewer title.
- The capture viewer now uses a shorter overview header, a wider Time column, and a compact in-panel result status instead of a footer.
- The search field no longer shows a decorative magnifier, and its Ctrl F hint now uses the same pixel font as the rest of the viewer.
- The event pager now accepts a page number so large captures can be opened directly without stepping through every page.
- The event stream keeps wheel and keyboard scrolling but hides its nested scrollbar for a cleaner right edge.
- The event stream now uses the same thin, theme-colored scrollbar as the Packet and Kind filter menus.
- Packet filter menus now expand to fit longer packet names before applying a viewport-safe maximum width.
- Packet filter menus now use a thin, theme-colored scrollbar instead of the browser default.
- The event stream now sizes the Seq column from the visible sequence numbers so long values do not collapse into ellipses.
- Long event Details messages now wrap onto additional lines instead of being truncated with an ellipsis.
- Live capture refreshes now preserve the open filter menu's scroll position while its options are rebuilt.
- The compact capture overview now vertically centers each metric label and value within the header row.
- The contributor validation workflow now requires viewer asset, HTTP, JavaScript, and browser smoke checks in the same change whenever a web-facing change could affect the page.

### Fixed

- Fixed the page-level footer overlapping the viewer inspector when decoded details extend beyond the initial viewport. The footer now stays below the complete workspace, including Blob bytes.

## [0.2.1] - 2026-09-17

Supported Bedrock version: 1.26.50 (protocol 2193).

[Published release](https://github.com/NhanAZ/BedrockDebugProxy/releases/tag/v0.2.1) | [All changes since v0.2.0](https://github.com/NhanAZ/BedrockDebugProxy/compare/v0.2.0...v0.2.1)

### Added

- Added weekly Dependabot checks for the root Go module and GitHub Actions, with grouped non-major Go updates and no automatic merge.
- Added a weekly Mojang protocol release watcher that opens a deduplicated investigation issue without changing protocol code automatically.

### Changed

- Clarified maintainer review-comment guidance for bot-authored dependency pull requests, including concise natural wording, bot-neutral language, and punctuation.
- Documented the server-policy disclaimer, AI-assisted maintenance, protocol research, and resource-pack encryption review in the project guidance.
- Release notes now link to the relevant section in this changelog and to GitHub's compare view for the complete release commit history.

### Fixed

- Aligned Bedrock 1.26.50 sub-chunk height-map decoding with the released upstream schema, including row-length validation while preserving raw forwarding.

### Documentation and maintenance

- Updated dependency provenance and protocol baselines to record the merged gophertunnel 1.26.50 correction.

## [0.2.0] - 2026-09-16

Supported Bedrock version: 1.26.50 (protocol 2193).

[Published release](https://github.com/NhanAZ/BedrockDebugProxy/releases/tag/v0.2.0) | [All changes since v0.1.0](https://github.com/NhanAZ/BedrockDebugProxy/compare/v0.1.0...v0.2.0)

### Added

- Added Bedrock 1.26.50 support with protocol 2193 metadata and focused packet coverage.

### Changed

- Capture writes now use one bounded, ordered asynchronous writer with blocking backpressure and a never-drop policy, keeping disk work off packet-forwarding callbacks while preserving event order and raw evidence.
- URL-advertised resource packs are prefetched concurrently with bounded workers and committed in the original offer order.
- Pre-spawn packet summaries are shown independently for each transfer hop.

### Fixed

- Packed `PlayerAuthInput` inventory actions now preserve the 1.26.50 `Hand` field during decoding and encoding.
- Coordinated peer-close handling no longer reports expected cancellation and closed-connection races as blocking bridge errors.

### Documentation and maintenance

- Documented the upstream fair-play contribution process for defects found in third-party dependencies.
- Release workflow documentation now requires synchronization with the private source-history backup after a published release.

## [0.1.0] - 2026-09-14

[Published release](https://github.com/NhanAZ/BedrockDebugProxy/releases/tag/v0.1.0)

- Initial public release.
