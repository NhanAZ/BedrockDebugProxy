# Fix history

This is the agent-facing index of investigated bugs and compatibility fixes. It is a navigation aid, not a replacement for Git history, architecture decisions, research notes, captures, or validation reports. Git remains the authoritative source for the complete commit history.

## How to use this index before changing code

1. Search this file for the reported symptom, server, packet family, and error text.
2. Open the linked decision or research note and inspect the cited capture or validation report when it exists.
3. Classify the report as a known fix, a regression, a known limitation, a duplicate, or a new root cause. Do not create a second implementation for a known fix.
4. If the evidence shows a different root cause, add a new `FIX-NNNN` entry in the same coherent change as the fix and link the old entry as related history.
5. Keep server-specific observations separate. A passing capture on one server is not evidence that another server uses the same packet order, timing, optional packets, transport, or resource-pack policy.

Every new entry should state the symptom, root cause, implementation commits, evidence, validation status, limitations, and follow-up work. The per-commit record in [`docs/audits/`](audits/) explains the exact change that introduced or updated an entry.

## Compatibility and bug index

### FIX-0001 - Experience destination resolution

- Symptom: default or named experience destinations could resolve to the wrong address or fail to select the intended featured experience.
- Root cause: destination resolution and the default experience path did not share one explicit selector boundary.
- Implementation: `42ae317`, `68b9795`.
- Evidence and validation: [`docs/research/experience-routing.md`](research/experience-routing.md); local experience routing tests.
- Status: fixed. Recheck the resolver when adding or renaming an Experience.

### FIX-0002 - Device authentication reuse and logout

- Symptom: repeated runs required unnecessary device-code authentication, and switching accounts had no supported CLI action.
- Root cause: the device credential cache had no operator-facing lifecycle command or one-click login URL.
- Implementation: `eb8eab0`, `483614f`, `88ab196`.
- Evidence and validation: auth cache tests and CLI checks. The cached token remains sensitive and is never committed.
- Status: fixed. Account switching requires the explicit logout command before the next device login.

### FIX-0003 - URL resource-pack offers

- Symptom: servers offering HTTP(S) resource packs could fall back to incomplete delivery or disconnect before the client received the advertised pack.
- Root cause: URL offers were not reconstructed and delivered through the same session-bound pack path as protocol chunk downloads.
- Implementation: `112ceb4`, `d46a10f`, `dd8845e`.
- Evidence and validation: [`docs/resource-packs.md`](resource-packs.md), [`docs/research/galaxite-resource-pack-flow.md`](research/galaxite-resource-pack-flow.md), and captured pack archives.
- Status: fixed for the supported current-connection flow. Offline retrieval, key guessing, and arbitrary remote pack download remain out of scope.

### FIX-0004 - Coordinated shutdown and upstream cancellation

- Symptom: after the downstream client left, an upstream read or recorder worker could continue and print a misleading error.
- Root cause: coordinated shutdown did not cancel all upstream work at the same connection boundary, and recorder-close races were logged at error level.
- Implementation: `3f32962`, `813e19f`, `5f553fa`.
- Evidence and validation: loopback shutdown tests and closed-capture verification.
- Status: fixed. The final recorder warning is expected when a read finishes after the recorder closes; it is not a capture-integrity failure by itself.

### FIX-0005 - Featured-experience transfer following

- Symptom: a server `Transfer` ended the first hop, but the client remained on a loading screen because the next backend was not reached or its RakNet route was not ready.
- Root cause: the proxy did not follow transfers, or the next hop lost the route and handshake state required by the featured backend.
- Implementation: `9878bb5`, `ab23673`, `d1929d4`.
- Evidence and validation: [`docs/decisions/0003-opt-in-transfer-following.md`](decisions/0003-opt-in-transfer-following.md), [`docs/research/enchanted-transfer-handshake.md`](research/enchanted-transfer-handshake.md), and the successful Enchanted transfer capture cited there.
- Status: fixed for the opt-in `--follow-transfers` RakNet path. The option is intentionally opt-in and each hop has its own connection and pack exchange.

### FIX-0006 - Featured-experience spawn without `ChunkRadiusUpdated`

- Symptom: Enchanted and Lifeboat could send `PlayStatusPlayerSpawn` without the radius acknowledgement that the upstream library expected, leaving the dialer blocked.
- Root cause: the connection finalizer required two signals even though the observed featured flow completed with one.
- Implementation: `0df6b63`.
- Evidence and validation: [`docs/research/bedrock-ecosystem.md`](research/bedrock-ecosystem.md) and `minecraft/conn_featured_experience_test.go`.
- Status: fixed for the observed optional-packet flow. The patch does not rewrite or suppress the packet.

### FIX-0007 - Slow resource-pack and transfer bounds

- Symptom: a slow but active featured resource-pack exchange could time out, while an unreachable transfer backend could leave the client waiting indefinitely.
- Root cause: the upstream login and transfer route had no useful bounded wait aligned with observed pack pacing.
- Implementation: `c94d105`, `0fc7ef8`.
- Evidence and validation: [`docs/resource-packs.md`](resource-packs.md), [`docs/validation.md`](validation.md), and timeout regression tests.
- Status: fixed with bounded waits. A timeout is recorded as an observable failure and does not silently discard the capture.

### FIX-0008 - Same-protocol raw forwarding

- Symptom: decoding and re-encoding an otherwise unchanged packet could alter optional-field encoding, batching, or other wire details.
- Root cause: the bridge used the typed read/write path even when both sides negotiated the same protocol.
- Implementation: `7691681`, `96d8a01`.
- Evidence and validation: [`docs/decisions/0004-same-protocol-raw-forwarding.md`](decisions/0004-same-protocol-raw-forwarding.md), packet-preservation tests, and raw forwarding timing evidence.
- Status: fixed. Typed decoding remains a derived observation; raw packet bytes remain authoritative.

### FIX-0009 - Console and shutdown observability

- Symptom: high-volume resource-pack traffic spammed the console, and orderly shutdown looked like an unexplained error or appeared to stop before capture verification finished.
- Root cause: operator summaries were emitted for every small batch and shutdown stages were not named consistently.
- Implementation: `a1ff5fe`, `5f553fa`, `2818a67`.
- Evidence and validation: logger tests, PocketMine color comparison, and closed-capture terminal runs.
- Status: fixed. Detailed evidence remains in the capture, while the console reports progress through forwarding stop, connection close, recorder close, artifact completion, and verification.

### FIX-0010 - Deferred login packets across valid server orderings

- Symptom: Enchanted hop 2 sent `ResourcePacksInfo` before `PlayStatus`; the offer was deferred and the client stayed on the resource-pack loading screen until the upstream wait expired.
- Root cause: changing the login state did not re-check packets already received for a later state. A later-phase packet could also sit ahead of the packet that advances the current state.
- Implementation: `a600bbf`, `7d94229`, `7a64a0d`.
- Evidence and validation: failed capture `captures/session-20260914T105050Z`, successful older Enchanted capture `captures/session-20260914T124716Z`, [`docs/decisions/0005-deferred-login-packet-drain.md`](decisions/0005-deferred-login-packet-drain.md), and login-order permutation tests.
- Status: fixed in tests and validated on Enchanted with the earlier state-aware implementation. The latest queue-scan implementation still needs an exact-revision The Hive baseline and a fresh five-server release matrix.

### FIX-0011 - Bedrock 1.26.50 protocol 2193 support

- Symptom: a client or server running Bedrock 1.26.50 could use packet IDs and layouts that the 1.26.45 protocol model did not represent correctly.
- Root cause: the latest stable gophertunnel release predates Bedrock 1.26.50, while its 1.26.50 feature branch diverges before later stable fixes and cannot safely replace the selected tree wholesale.
- Implementation: this coherent protocol-update commit; see [`AUDIT-0008`](audits/AUDIT-0008-protocol-1.26.50.md) and Git history.
- Evidence and validation: [`docs/research/protocol-1.26.50.md`](research/protocol-1.26.50.md), Mojang's released schemas, reviewed gophertunnel revisions, and focused metadata, packet-pool, and binary-body tests.
- Status: implemented with the full project quality gate and nested gophertunnel tests passing. A stamped exact-revision The Hive report remains required before the runtime change is complete.

### FIX-0012 - Packed item-use hand alignment

- Symptom: a Bedrock 1.26.50 `PlayerAuthInput` packet containing item interaction decoded with shifted position and optional fields, left four zero bytes, and produced a bridge decode error at the end of an otherwise active The Hive session.
- Root cause: the 1.26.50 `UseItemTransactionData` model added the `Hand` byte between `HotBarSlot` and `HeldItem`, but the packed `PlayerInventoryAction` reader and writer did not include that field.
- Implementation: this follow-up fix adds the missing read and write operation and a round-trip regression test.
- Evidence and validation: capture `session-20260916T002518Z` (local ignored evidence), [`docs/research/protocol-1.26.50.md`](research/protocol-1.26.50.md), the Mojang `ItemUseInventoryTransaction` schema, and the focused protocol test. A new stamped The Hive report is still required.
- Status: fixed in the source tree. Exact-revision live validation is pending.

### FIX-0013 - Racing peer-close write errors

- Symptom: Mineville Zeqa could close the client-facing connection during normal termination while the peer forwarding loop concurrently returned `bridge.write_error` with `context.Canceled`, causing the report to classify the closed session as a blocking error.
- Root cause: each forwarding loop waited for `runHop` to mark coordinated shutdown after receiving the first error. The read and write loops could therefore classify the same connection close before the marker was visible to either loop.
- Implementation: the first terminal forwarding error now atomically claims shutdown before classification. Read-side close evidence remains recorded, while first or subsequent write-side `context.Canceled` and `net.ErrClosed` results are retained in `session.close` diagnostics without creating a blocking bridge error.
- Evidence and validation: Mineville Zeqa capture `captures/session-20260916T025636Z` at revision `9a7859c351b55b66a89cfa7e3448b809f5a48584` showed `bridge.write_error` racing a downstream context-cancelled termination. New shutdown race tests cover read-first and write-first ordering. A fresh stamped live report is required after this fix.
- Status: fix implemented; exact-revision live validation pending.

### FIX-0014 - Per-hop pre-spawn console summaries

- Symptom: after an Enchanted transfer, the capture contained hop 2 login and resource-pack packets but the terminal showed no corresponding `UPSTREAM` or `DOWNSTREAM` summaries.
- Root cause: the live reporter used one process-wide spawned flag. Once hop 1 reached gameplay, raw packet summaries were suppressed for every later transfer hop, including its pre-spawn exchange.
- Implementation: live output now tracks spawned state per hop, flushes the previous bucket when a new hop becomes active, and keeps the existing three-second summary cadence and capture path unchanged.
- Evidence and validation: Enchanted capture `captures/session-20260916T044421Z` recorded hop 2 resource-pack exchange and spawn while the operator reported missing terminal summaries. The regression test verifies pre-spawn summaries for two hops and continues to suppress post-spawn raw packets within each hop.
- Status: implemented; a fresh stamped The Hive baseline and the five-server release matrix are required because observable runtime logging changed.

### FIX-0015 - Concurrent URL resource-pack prefetch

- Symptom: after an Enchanted transfer, a backend with several URL-advertised resource packs could keep the upstream login exchange open long enough for the backend to kick the proxy client for `Timed out`, even though hop 2 eventually reached spawn.
- Root cause: the URL resource-pack adapter downloaded and parsed each advertised archive serially inside the upstream packet reader. The downstream listener could not receive the pack offer until every preceding URL download completed, extending the upstream-first buffering window.
- Implementation: URL pack downloads now use a bounded worker pool and commit successful results in the original `ResourcePacksInfo` order. Duplicate-offer suppression, UUID/version/size validation, URL removal from downstream copies, and the sequential downstream listener path are unchanged.
- Evidence and validation: Enchanted capture `captures/session-20260916T053620Z` selected `factions-spawn.factions.connect.enchanted.gg:19132`; hop 2 recorded seven URL pack archives, reached `session.spawned`, and then received the upstream `You were kicked: Timed out` message. The capture had zero dropped, truncated, decode, or write errors. The concurrency regression test uses a barrier-backed HTTP fixture to prove overlapping downloads and stable offer order.
- Status: implemented with focused automated tests; a new stamped live Enchanted transfer and The Hive baseline are required before release approval.

### FIX-0016 - Asynchronous bounded capture writer

- Symptom: Galaxite could send a dense `UpdateBlock` burst while the proxy synchronously wrote raw, decoded, and timing events. The forwarding path then stalled long enough for the client to receive `You moved too slow`, while the direct client did not show the same behavior.
- Root cause: every capture event held the recorder ordering lock through blob hashing, temporary-file publication, JSON encoding, and event-stream writes. High-volume server-to-client traffic therefore blocked unrelated client-to-server forwarding callbacks behind disk I/O.
- Implementation: capture events now receive their sequence at submission and enter one bounded ordered writer queue. The queue has blocking backpressure, a configurable record and byte budget, and a `never_drop` loss policy. `RecordReader` drains queued events before streaming a barrier-sensitive blob, `Close` drains accepted jobs before publishing the final manifest, and writer failures are exposed through `Recorder.Err` and the forwarding loop.
- Evidence and validation: Galaxite capture `captures/session-20260916T060203Z` at revision `320478f65db244c8b4d8be05e756e2c7774289db` retained 489,910 events and 93,790 blobs with no drops, but recorded 40,259 `UpdateBlock` packets and rare 596 to 738 ms capture/write stalls before the `You moved too slow` bridge write failure. The burst-order regression test verifies sequence continuity, blob integrity, manifest queue metadata, and closed-capture verification.
- Status: implemented with automated tests; a new stamped Galaxite run and the five-server release matrix are required before release approval.

### FIX-0017 - Bedrock 1.26.50 sub-chunk height-map rows

- Symptom: decoded `SubChunk` height maps were represented as a flat 272-byte slice, so row-length prefixes were interpreted as height values and malformed row sizes could not be reported.
- Root cause: the local 1.26.50 composite retained the pre-`v1.62.0` flat decoder instead of the released upstream `HeightMap` structure.
- Implementation: `fix: align 1.26.50 subchunk height maps` adds the 16 by 16 `[z][x]int8` type, validates each varuint32 row length, and uses it for both optional sub-chunk height maps. Raw same-protocol forwarding remains unchanged.
- Evidence and validation: Mojang `bedrock-protocol-docs` `v1.26.50` at `475bd72ed89036af4eb18426774ef3b953de7603`, gophertunnel `v1.62.0` at `7a556a07335b663744b50d38062636ad8283f314`, and upstream correction `709768d`. `TestSubChunkHeightMapRoundTripUsesRows` verifies wire row prefixes and value preservation. A fresh stamped The Hive baseline remains required.
- Status: implemented with focused automated coverage; exact-revision live validation pending.

### FIX-0018 - Square capture viewer surfaces

- Symptom: the capture viewer used rounded corners on its logo, status indicator, controls, panels, pager buttons, and metadata cards even though the intended interface style is square.
- Root cause: the viewer stylesheet retained several independent `border-radius` declarations from the earlier visual design.
- Implementation: removed every `border-radius` declaration from the viewer stylesheet. Layout, colors, typography, controls, and capture behavior remain unchanged.
- Evidence and validation: a stylesheet search confirms that `internal/viewer/assets/app.css` contains no corner-radius declarations. Automated viewer tests and a local browser check cover the rebuilt interface.
- Status: implemented with no forwarding, capture, protocol, or resource-pack impact.

### FIX-0019 - Optional local Mojang font preference

- Symptom: the capture viewer could use only the embedded fan-made fonts even when the operator had a separately installed Mojangles or Minecraft Ten font available for personal use.
- Root cause: the viewer's `@font-face` sources referenced only the redistributable fallback files.
- Implementation: the heading and content faces now try the exact locally installed `Mojangles` and `Minecraft Ten` family names before the embedded OFL fallbacks. No proprietary font file is copied into the repository, executable, or license bundle.
- Evidence and validation: local name-table inspection identified the supplied files as `Mojangles` and `Minecraft Ten`. The official [Minecraft Usage Guidelines](https://www.minecraft.net/en-us/usage-guidelines) identify fonts as part of the Minecraft brand and prohibit redistribution of game files. A viewer HTTP regression test verifies both `local()` sources.
- Status: implemented with viewer-only impact. Exact appearance depends on the font being installed and available to the browser.

### FIX-0020 - Uniform normal-weight viewer typography

- Symptom: viewer text used several small and large font sizes and bold weights across headings, filters, metadata, packet rows, JSON, hex evidence, and the footer.
- Root cause: the stylesheet assigned independent type sizes and weights to each interface role, and headings fell back to an embedded bold face when the optional local font was unavailable.
- Implementation: normalize visible text to 14px, set all viewer text roles to normal weight, and use regular Monocraft as the heading fallback. Local Mojangles and Minecraft Ten preferences remain opt-in through browser `local()` lookup.
- Evidence and validation: a stylesheet audit found all prior size and weight declarations. The viewer HTTP regression test rejects the previous non-uniform sizes and bold weights while checking the local font sources. No forwarding, capture, protocol, or resource-pack behavior changes.
- Status: implemented with viewer-only impact. Browser rendering and local font availability remain platform-specific.

### FIX-0021 - Themed filter dropdowns

- Symptom: filter controls matched the viewer when closed, but opening a native select displayed a browser-owned popup with unrelated colors, spacing, and typography.
- Root cause: native option popups are rendered by the browser or operating system and do not consistently honor the viewer stylesheet.
- Implementation: filter selects now retain their native elements as synchronized value sources while exposing themed listbox controls with keyboard navigation, pointer selection, focus state, and outside-click closing. Existing query construction, reset behavior, and change listeners continue to use the native values.
- Evidence and validation: the viewer asset test checks the embedded themed dropdown selectors. `node --check internal/viewer/assets/app.js` checks the custom control script syntax. No forwarding, capture, protocol, or resource-pack behavior changes.
- Status: implemented with viewer-only impact. Popup dimensions and text rasterization remain subject to the browser viewport and installed font availability.

### FIX-0022 - Smaller unified viewer typography

- Symptom: the capture viewer still appeared oversized, and headings, labels, and status controls could use a different locally installed face or forced uppercase presentation.
- Root cause: the stylesheet retained a separate heading face, repeated 14px declarations, and uppercase transforms for several visible roles.
- Implementation: visible viewer text now uses one regular Monocraft content face, with optional local Minecraft Ten lookup before the embedded fallback. A shared 12px size variable replaces repeated size declarations, and forced uppercase styling plus all-caps eyebrow literals were removed.
- Evidence and validation: the viewer stylesheet audit found the separate heading face, repeated 14px declarations, and four uppercase transforms. The viewer HTTP regression test now checks the single-face, 12px, normal-weight, and natural-casing contract. Browser smoke validation covers the rebuilt page and themed controls.
- Status: implemented with viewer-only impact. Browser zoom, operating-system scaling, and glyph fallback remain platform-specific.

### FIX-0023 - Single-column viewer workspace

- Symptom: the capture viewer split the event stream and inspection panels into two side-by-side columns, reducing the usable width of each evidence view.
- Root cause: the workspace grid defined independent event and inspector columns and retained a breakpoint override with a second two-column definition.
- Implementation: the workspace now has one full-width column with explicit scrollable rows for the event stream and inspector. The inspector still stacks decoded details above raw evidence, and the data flow and controls remain unchanged.
- Evidence and validation: the stylesheet contained two workspace column definitions before this change. The viewer HTTP regression test now asserts the single-column track, stacked row sizing, and workspace scrolling while rejecting the prior two-column tracks. Browser smoke validation covers the rebuilt layout and filter interaction.
- Status: implemented with viewer-only impact. The workspace scroll position and panel heights depend on the browser viewport.

### FIX-0024 - Gold pending-decode viewer status

- Symptom: raw packet rows with the intentional `not_yet_decoded` status were rendered in the same red style as actual decode problems, making preserved raw evidence look like a failure.
- Root cause: the viewer treated every non-`decoded` status as a red problem presentation without distinguishing the observer's pre-decode raw-capture marker.
- Implementation: classify `not_yet_decoded` rows in the embedded viewer script and render their row accent and summary text with the gold `#FFAA00` color. Other problem statuses keep the existing red styling.
- Evidence and validation: `internal/bedrock/observer.go` intentionally writes `not_yet_decoded` at `post_decrypt_decompress_frame`, while decode failures are represented by separate `packet.decode_error` or view-error events. Viewer asset tests assert the gold selectors and status classification, and JavaScript syntax validation remains required.
- Status: implemented with viewer-only impact. The `Problems only` filter still includes pending-decode rows because they remain useful review targets, but their presentation no longer implies a decode failure.

### FIX-0025 - Expand event details without an internal scrollbar

- Symptom: the selected event details panel constrained its decoded metadata and tree to a fixed grid track, forcing a second vertical scrollbar inside the panel.
- Root cause: the inspector and detail content used fractional rows and `overflow: auto`, so content was clipped to the panel viewport instead of contributing to the workspace height.
- Implementation: let the inspector and detail panel use content-sized rows, allow the detail content to expand naturally, and keep the workspace as the scroll boundary. The raw evidence panel and event table retain their bounded evidence views.
- Evidence and validation: the previous stylesheet contained the fractional inspector rows and `.detail-content` internal overflow. Viewer asset tests reject those declarations and browser smoke validation checks a selected raw event in the expanded panel.
- Status: implemented with viewer-only impact. Very large decoded trees can still make the workspace long, which is intentional so the evidence remains reachable without a nested detail scrollbar.

### FIX-0026 - Preserve the expanded event-details track

- Symptom: after the first event-details expansion change, the details panel could collapse to a thin header when the workspace grid was constrained by the browser viewport.
- Root cause: the workspace's content-sized row and the inspector's auto tracks still allowed grid items with `min-height: 0` to contribute zero intrinsic height, so the outer scroll boundary had no reliable minimum for the details content.
- Implementation: size the inspector and details rows from their max-content height, keep the detail panel and content at max-content minimums, and leave the workspace as the only scroll boundary. The raw evidence panel remains a separate bounded view.
- Evidence and validation: the viewer asset regression test now requires max-content sizing and rejects the zero-minimum inspector rule. Browser smoke validation must inspect the details panel at a constrained viewport and confirm that its metadata and decoded fields remain visible without an internal scrollbar.
- Status: implemented with viewer-only impact. Very large decoded trees intentionally increase the outer workspace scroll height.

### FIX-0027 - Use one page-level viewer scrollbar

- Symptom: the red scrollbar began below the header and filters because the workspace owned the vertical scroll context, making the page feel split into a fixed control area and a separate scrollable body.
- Root cause: `body` was height-constrained with `overflow: hidden`, while `.workspace` used `overflow: auto` as an inner scroll container.
- Implementation: let the body grid grow with its content and own the vertical scrollbar. The workspace now keeps a bounded event-stream row, expands the inspector below it, and leaves its overflow visible so the header, filters, details, and raw evidence share one page-level scroll context.
- Evidence and validation: the viewer asset regression test now requires the page-sized body tracks and the bounded workspace row, while rejecting the previous body-only grid constraint. Browser smoke validation must confirm that the scrollbar starts at the top of the page and that selected event details remain reachable below the event stream.
- Status: implemented with viewer-only impact. The event table retains its local table overflow for its paginated rows.

### FIX-0028 - Collapse Blob bytes by default

- Symptom: selecting an event with a large blob made the raw hexdump occupy most of the page even when the operator was inspecting decoded details.
- Root cause: the raw evidence `<pre>` was always visible and its panel expanded to the full preview height for every selected blob.
- Implementation: keep the Blob bytes panel collapsed on initial load, preserve the byte count and representation in the header, and add an accessible toggle that expands or collapses the hexdump without changing captured bytes.
- Evidence and validation: the viewer HTTP test checks the collapsed HTML contract, themed toggle CSS, and JavaScript state synchronization. Browser smoke validation must select an event, confirm Blob bytes stays collapsed, then expand it and confirm the hexdump is visible.
- Status: implemented with viewer-only impact. The operator's expanded or collapsed state is local to the current page and is not persisted in the capture.

### FIX-0029 - Add viewer favicon package

- Symptom: the local capture viewer used the browser's generic tab icon and did not advertise installable app metadata.
- Root cause: the embedded viewer had no favicon, touch icon, or web manifest assets.
- Implementation: embed the supplied favicon package, serve its standard icon variants and manifest through the token-scoped viewer handler, and link them from the document head. The manifest points back to the token-scoped viewer root and uses the viewer's dark theme.
- Evidence and validation: the viewer asset test checks the head links, manifest response, content types, and PNG/ICO signatures. Browser smoke validation must confirm the page loads the favicon and manifest without console errors.
- Status: implemented with viewer-only impact. The icons are static presentation assets and do not affect captures, packet handling, or viewer data.

### FIX-0030 - Use favicon package logo in viewer brand mark

- Symptom: the browser tab used the supplied favicon while the viewer header still displayed the older CSS gradient logo.
- Root cause: the header brand mark was a CSS-only decoration and was not connected to the favicon package.
- Implementation: serve the package's original logo asset and render it in the header at the existing 38px mark size.
- Evidence and validation: the viewer asset test checks the header image, logo response, PNG signature, and CSS image rendering rule. Browser smoke validation must confirm the header mark matches the favicon artwork.
- Status: implemented with viewer-only impact. The logo is static presentation content and does not affect captures, packet handling, or viewer data.

### FIX-0031 - Compact capture overview in viewer header

- Symptom: the capture viewer used a separate overview strip for four summary metrics, consuming a full horizontal row before the filters and event stream.
- Root cause: the overview section was laid out as a standalone page grid row with tall metric cards.
- Implementation: move the existing overview section into the topbar, render its four metrics as compact horizontal label/value pairs, reduce the page grid to its four actual children, and keep auto page tracks content-sized so the footer remains after the complete workspace. Existing metric IDs, values, accessibility label, and capture data remain unchanged.
- Evidence and validation: the viewer HTTP test verifies that the overview is inside the header and that the compact topbar grid is present. Browser smoke validation must confirm the brand, metrics, status, and capture ID share one row without clipping.
- Status: implemented with viewer-only impact. Long generator values and capture IDs may ellipsize in narrow windows.

### FIX-0033 - Remove verification and capture identity from viewer header

- Symptom: the verification pill and long capture identifier occupied the right side of the compact header even though they were not needed for the operator's packet review.
- Root cause: the header retained the status and capture identity presentation from the earlier full-width layout.
- Implementation: remove the status and capture identifier elements, their CSS, and their JavaScript updates. The overview metrics now use the freed header column without changing the capture manifest or session API.
- Evidence and validation: the viewer HTTP test rejects the removed element IDs and selectors, and browser smoke validation confirms only the brand and compact metrics remain in the header.
- Status: implemented with viewer-only impact. Capture identity and verification remain available in the capture manifest and API for programmatic inspection.

### FIX-0032 - Keep viewer footer below the inspector

- Symptom: the page footer could appear between decoded details and Blob bytes when the inspector content overflowed the initial viewport.
- Root cause: the body and workspace grid tracks were allowed to shrink below the content-sized inspector, so the footer was placed at the viewport track while the inspector continued painting below it.
- Implementation: size the page tracks and workspace from their max-content height so the footer follows the full workspace while retaining the page-level scrollbar.
- Evidence and validation: the viewer asset test checks the content-sized page grid contract. Browser smoke validation confirms the footer remains below Blob bytes after selecting an event and expanding the details.
- Status: implemented with viewer-only impact. Long decoded events still increase the page scroll height by design.

### FIX-0034 - Inline viewer panel headings

- Symptom: each viewer panel used two stacked heading lines, consuming unnecessary vertical space.
- Root cause: the eyebrow and panel title were rendered as a vertical block.
- Implementation: render each pair as one inline title row with an explicit hyphen separator for Ordered stream, Decoded view, and Raw evidence. Panel controls and data are unchanged.
- Evidence and validation: viewer HTTP tests check inline heading markers and CSS, and browser smoke validation confirms all three headings remain readable on one line.
- Status: implemented with viewer-only impact.

### FIX-0035 - Simplify viewer brand header

- Symptom: the viewer header used a decorative logo and a secondary Capture Viewer title, taking space away from the capture overview.
- Root cause: the brand block retained the earlier two-line identity treatment after the header was compacted.
- Implementation: remove the header logo and secondary title, keep BedrockDebugProxy as the sole visible brand name, and retain the favicon assets for browser metadata.
- Evidence and validation: the viewer HTTP test rejects the removed header elements and title text, while browser smoke validation confirms the name and overview metrics remain aligned.
- Status: implemented with viewer-only impact.

### FIX-0036 - Compact viewer controls and result status

- Symptom: the overview header left excess vertical space, Time values were truncated, the result count occupied a separate footer, and the search field showed an unwanted magnifier.
- Root cause: the header retained its earlier minimum height, the Time column was too narrow for full millisecond timestamps, viewer status was coupled to footer metadata, and the search affordance included a decorative glyph.
- Implementation: reduce the overview header height, widen the Time column to 132px, move the live result status into the event panel heading, remove the footer and magnifier, and style the Ctrl F hint with the shared pixel font. The web manifest now uses BedrockDebugProxy as its app name.
- Evidence and validation: viewer HTTP tests check the compact page grid, status placement, removed footer and glyph, wider Time column, manifest name, and shared pixel-font declaration. Browser smoke validation confirms the header, table, in-panel count, and search hint remain readable.
- Status: implemented with viewer-only impact.

### FIX-0037 - Use the shared pixel font for the search hint

- Symptom: the Ctrl F hint used a non-pixel sans-serif fallback while the rest of the viewer used the pixel font.
- Root cause: the search hint had an explicit Space Grotesk font override.
- Implementation: use the shared `var(--font-content)` family for the Ctrl F hint so every visible viewer role follows one pixel-font stack.
- Evidence and validation: the viewer asset test now requires the shared font declaration for the hint and rejects the separate Space Grotesk rule. Browser smoke validation confirms the hint matches the search field typography.
- Status: implemented with viewer-only impact.

### FIX-0038 - Direct event page entry

- Symptom: large captures required repeated previous or next clicks to reach a distant event page.
- Root cause: the pager displayed the current page as text and exposed no direct page-entry control.
- Implementation: replace the current-page text with a bounded number input, keep the total page count beside it, and load the requested offset on Enter or change while clamping invalid values to the available range.
- Evidence and validation: the viewer HTTP test checks the page input markers and JavaScript page-jump handler. Browser smoke validation confirms the input remains aligned with the existing pager controls and accepts direct page navigation.
- Status: implemented with viewer-only impact.

### FIX-0039 - Center compact overview metrics

- Symptom: the compact header's metric text appeared raised within the overview row compared with the BedrockDebugProxy name.
- Root cause: the metric flex rows used baseline alignment, which placed their single-line content at the cross-axis start instead of centering it in the stretched header cells.
- Implementation: use center alignment for each metric row while keeping the existing compact header height and horizontal layout.
- Evidence and validation: the viewer asset test continues to cover the compact header layout, and browser smoke validation confirms the metric labels and values are vertically centered beside the brand name.
- Status: implemented with viewer-only impact.

### FIX-0040 - Hide the event stream scrollbar chrome

- Symptom: the event stream displayed a second vertical scrollbar inside the panel's right edge.
- Root cause: the paginated table wrapper retained native scrollbar chrome even though the viewer already communicates page movement through the pager and remains wheel-scrollable.
- Implementation: hide the table wrapper's scrollbar track and thumb across Firefox, legacy Windows behavior, and Chromium while preserving `overflow: auto` scrolling.
- Evidence and validation: the viewer asset test checks the cross-browser scrollbar-hiding declarations. Browser smoke validation confirms the event rows remain available and the nested scrollbar is no longer visible.
- Status: implemented with viewer-only impact.

### FIX-0041 - Expand packet filter menu labels

- Symptom: longer packet names in the Packet filter menu were truncated even when there was room to show the full label.
- Root cause: the custom menu used a fixed minimum width tied to the trigger instead of sizing itself from its option content.
- Implementation: size the menu from its max-content width, keep it at least as wide as the trigger, and cap it at a viewport-safe 480px maximum.
- Evidence and validation: the viewer asset test checks the intrinsic menu width and viewport cap. Browser smoke validation confirms long packet names remain readable in the expanded Packet menu.
- Status: implemented with viewer-only impact.

### FIX-0042 - Theme packet filter scrollbars

- Symptom: the expanded Packet and Kind menus used a thick, bright native scrollbar that did not match the viewer controls.
- Root cause: the custom menu had no scrollbar styling, so each browser supplied its default track and thumb.
- Implementation: use a thin six-pixel Chromium scrollbar with a transparent track and `--line-bright` thumb, plus the Firefox `scrollbar-width` and `scrollbar-color` equivalents.
- Evidence and validation: the viewer asset test checks the cross-browser themed scrollbar declarations. Browser smoke validation confirms the dropdown scrollbar is thin and aligned with the existing dark palette.
- Status: implemented with viewer-only impact.

### FIX-0043 - Theme event stream scrollbar

- Symptom: the event stream's internal scrollbar was hidden, so its bounded table area had no visible drag affordance beside the rows.
- Root cause: the earlier cleanup hid `.table-wrap` scrollbar chrome even though the table remains an independently scrollable viewport.
- Implementation: restore the scrollbar and apply the same thin six-pixel track, transparent background, `--line-bright` thumb, and muted hover color used by the Packet and Kind menus. Wheel and keyboard scrolling remain unchanged.
- Evidence and validation: the viewer asset test checks the themed `.table-wrap` declarations and rejects the previous hidden-scrollbar rules. Browser smoke validation confirms the event stream scrollbar matches the filter menus.
- Status: implemented with viewer-only impact.

### FIX-0044 - Constrain viewer colors to the Minecraft palette

- Symptom: the viewer used custom navy, pastel, alpha-blended, and gradient colors outside the requested Minecraft formatting-code palette, which also made some text and surface combinations difficult to audit for contrast.
- Root cause: the stylesheet had accumulated independent color literals, translucent RGB overlays, shadows, and gradient backgrounds instead of a finite palette contract.
- Implementation: map viewer tokens to the supplied Minecraft foreground and background colors, remove gradients and RGB/HSL color functions, replace blended surfaces with opaque palette colors, remove decorative shadows, and keep text colors bright against dark palette backgrounds. The manifest and theme-color metadata now use `#000000`.
- Evidence and validation: the viewer asset test rejects non-palette CSS color literals and gradients, checks the manifest and HTML theme color, and the clean browser smoke test confirms the opaque palette renders across the viewer without console errors.
- Status: implemented with viewer-only impact.

### FIX-0045 - Minimize viewer scrollbars

- Symptom: filter menus and the event stream used more scrollbar chrome than the surrounding pixel-style controls needed.
- Root cause: the browser scrollbar declarations used six-pixel tracks, bright bordered thumbs, and left native arrow buttons visible.
- Implementation: use four-pixel square scrollbars, palette-colored tracks and thumbs, no thumb border, and hidden WebKit scrollbar buttons for the filter menus and event stream.
- Evidence and validation: the viewer asset test checks the compact cross-browser declarations. Browser smoke validation confirms the Packet menu and event stream retain usable scrolling without native arrow buttons.
- Status: implemented with viewer-only impact.

### FIX-0046 - Reflow viewer for narrow screens

- Symptom: the viewer kept a 780-pixel page minimum and exposed a horizontal filter scrollbar when opened in a narrow browser viewport.
- Root cause: the desktop grid and single-line filter row were not allowed to reflow below the desktop width.
- Implementation: remove the global page minimum width, wrap filters below 1100 pixels, stack the header and overview metrics below 640 pixels, let event controls wrap, and keep only the event table's content scrollable on mobile.
- Evidence and validation: the viewer asset test checks the responsive media queries and narrow table contract. Browser smoke validation covers desktop, tablet, and phone viewport layouts with no document-level horizontal overflow.
- Status: implemented with viewer-only impact.

### FIX-0047 - Theme page scrollbar

- Symptom: the outer page scrollbar remained native gray and thicker than the Packet, Kind, and event stream scrollbars.
- Root cause: only the inner scrollable regions had custom scrollbar declarations.
- Implementation: apply the same four-pixel square palette scrollbar to the document root and body, including hidden WebKit arrow buttons.
- Evidence and validation: the viewer asset regression test checks the global scrollbar declarations. Browser smoke validation confirms that the far-right page scrollbar matches the inner controls.
- Status: implemented with viewer-only impact.

### FIX-0048 - Use supplied pixel toggle glyphs

- Symptom: dropdown and Blob bytes toggles used CSS borders and transforms instead of the supplied pixel glyph artwork.
- Root cause: the viewer drew its chevrons as generic CSS geometry and did not bundle the requested SVG assets.
- Implementation: embed the supplied `glyph_e031.svg` and `∨.svg` artwork under ASCII-safe bundle names, serve both as same-origin SVG assets, and use them for the dropdown and Blob bytes toggle indicators.
- Evidence and validation: the viewer asset test checks both SVG responses, MIME types, and CSS references. Browser smoke validation confirms the dropdown and Blob bytes controls render the bundled pixel glyphs.
- Status: implemented with viewer-only impact.

### FIX-0049 - Replace native JSON disclosure markers

- Symptom: expanded JSON event details still showed the browser's native triangle marker beside nested values.
- Root cause: `<details>` summaries retained the platform disclosure marker even after the custom dropdown and Blob bytes indicators were replaced.
- Implementation: hide the native summary markers and render the bundled `glyph_e031.svg` asset for open and closed JSON tree nodes, rotating the closed state for the collapsed direction.
- Evidence and validation: the viewer asset test checks the cross-browser marker removal declarations and glyph reference. Browser smoke validation must inspect an expanded event JSON tree and confirm that no native triangle remains.
- Status: implemented with viewer-only impact.

### FIX-0050 - Square the filter checkboxes

- Symptom: the native checkbox inside the Problems only control retained a rounded platform shape that did not match the viewer's square controls.
- Root cause: the checkbox input used browser-native appearance and accent coloring.
- Implementation: remove native checkbox appearance, render a twelve-pixel square border with a palette-colored checked fill, and apply the same geometry to the hidden live-follow control.
- Evidence and validation: the viewer asset test checks the custom checkbox declarations and checked-state marker. Browser smoke validation confirms the Problems only control keeps its native label behavior while the checkbox uses square corners.
- Status: implemented with viewer-only impact.

### FIX-0052 - Size the event sequence column from visible values

- Symptom: larger sequence numbers such as `#93461` were truncated to `#934...` in the event stream.
- Root cause: the first table column used a fixed width that was smaller than the rendered sequence text once a capture reached five or more digits.
- Implementation: the viewer sets a CSS custom property from the longest visible sequence number after each page load, and the Seq header uses that value with a small padding allowance at desktop and narrow breakpoints.
- Evidence and validation: the viewer asset test checks the dynamic sizing hook and CSS contract. Browser smoke validation must open a large capture on a page containing a five-digit sequence and confirm the complete value is visible without an ellipsis.
- Status: implemented with viewer-only impact. The width follows the current page and may change when navigating between pages.

### FIX-0053 - Wrap long event details

- Symptom: long error summaries in the Details column were clipped with an ellipsis, hiding the address and reason returned by the server.
- Root cause: every event table cell inherited single-line ellipsis styling, including the Details column.
- Implementation: mark Details cells explicitly and allow their text to wrap at normal boundaries or inside long unbroken values while preserving ellipsis behavior for the other compact columns.
- Evidence and validation: the viewer asset test checks the Details class and wrapping declarations. Browser smoke validation must open the Hive ban event on the large capture and confirm its full error summary continues on additional lines without the ellipsis.
- Status: implemented with viewer-only impact. A wrapped Details message increases that event row's height.

### FIX-0054 - Preserve filter menu scroll during live refresh

- Symptom: opening the Packet or Kind filter in a live viewer and scrolling down caused the menu to jump back to its first option during the periodic refresh.
- Root cause: each live `loadSession` refresh rebuilt the custom menu with `replaceChildren`, which reset its scroll position before the next poll completed.
- Implementation: capture the custom menu's current `scrollTop` before rebuilding its options and restore it after active-option synchronization.
- Evidence and validation: the viewer asset test checks the scroll-preservation hook. Browser smoke validation must keep a long filter menu open through a live refresh and confirm its visible option range remains stable.
- Status: implemented with viewer-only impact. The selected filter value and available options still follow the latest session response.

### FIX-0055 - Preserve downstream identity on unauthenticated upstream dials

- Symptom: a local unauthenticated upstream server saw the proxied client as gophertunnel's default `Steve` identity instead of the downstream client's identity.
- Root cause: the listener exposed the downstream `ClientData` to the asynchronous upstream dial callback, but the callback discarded the accompanying `login.IdentityData`. The upstream `minecraft.Dialer` therefore generated default identity data when no token source was configured.
- Implementation: pass the callback identity through `connectUpstream` and set `minecraft.Dialer.IdentityData`. Authenticated dials still replace this value with the identity obtained from the Xbox authentication chain.
- Evidence and validation: [`docs/research/local-identity-parity.md`](research/local-identity-parity.md), the upstream identity regression assertion in `internal/proxy/proxy_test.go`, and a local PocketMine direct-versus-proxy loopback run.
- Status: fixed for the unauthenticated local path. The authenticated path still requires a real vanilla-client validation; this change does not infer public-server anti-cheat behavior.

### FIX-0056 - Distinguish protocol and ClientData versions in negotiation evidence

- Symptom: `session.negotiated` exposed the gophertunnel protocol package version but did not show the game version or device OS reported inside each login's `ClientData`, making a local parity mismatch easy to miss.
- Root cause: the event used `Conn.Proto().Ver()` for its two game-version fields and omitted the login metadata comparison even though both connection snapshots were available immediately afterward.
- Implementation: retain the existing protocol fields and add downstream/upstream `ClientData` game-version and device-OS fields plus explicit match booleans. The analysis summary exposes the same optional values while remaining compatible with older captures.
- Evidence and validation: focused analysis and proxy integration assertions cover both matching and mismatching metadata shapes. This is a capture-observability change and does not alter forwarded packet bytes.
- Status: implemented with no wire-behavior change.

## Ordering safety contract

The deferred-packet change is deliberately narrow. `expectedIDs` remains the only authority for what the current login state accepts. The queue scan selects the earliest packet that is currently expected, leaves packets for later states queued, and never mutates or reorders bytes on the wire. The regression tests cover canonical resource-pack order, featured early `ResourcePacksInfo`, a later-phase packet queued ahead of the current phase, and all six permutations of the three login packets.

This is a compatibility guard, not permission to assume a universal Bedrock sequence. Before changing an expected set, inspect the packet definitions, the relevant capture, and the server-specific validation flow. A new server observation must add a focused test and update this index or a linked research note.

## Open and pending items

- Generate a stamped binary from the current commit and run The Hive minimum baseline.
- Re-run the direct Enchanted transfer and resource-pack flow with the same binary.
- Complete the exact-revision matrix for Galaxite, Lifeboat, Mineville Zeqa, and Enchanted after the The Hive baseline before release.
- If a server fails, record the capture and classify the failure before changing the generic state machine.
