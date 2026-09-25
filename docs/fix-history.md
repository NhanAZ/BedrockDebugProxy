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
- Evidence and validation: failed capture [captures/session-20260914T105050Z](https://github.com/BedrockDebugProxy-Captures/session-20260914T105050Z/releases/tag/capture-archive), successful older Enchanted capture [captures/session-20260914T124716Z](https://github.com/BedrockDebugProxy-Captures/session-20260914T124716Z/releases/tag/capture-archive), [`docs/decisions/0005-deferred-login-packet-drain.md`](decisions/0005-deferred-login-packet-drain.md), and login-order permutation tests.
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
- Evidence and validation: capture [session-20260916T002518Z](https://github.com/BedrockDebugProxy-Captures/session-20260916T002518Z/releases/tag/capture-archive) (sensitive evidence archived privately), [`docs/research/protocol-1.26.50.md`](research/protocol-1.26.50.md), the Mojang `ItemUseInventoryTransaction` schema, and the focused protocol test. A new stamped The Hive report is still required.
- Status: fixed in the source tree. Exact-revision live validation is pending.

### FIX-0013 - Racing peer-close write errors

- Symptom: Mineville Zeqa could close the client-facing connection during normal termination while the peer forwarding loop concurrently returned `bridge.write_error` with `context.Canceled`, causing the report to classify the closed session as a blocking error.
- Root cause: each forwarding loop waited for `runHop` to mark coordinated shutdown after receiving the first error. The read and write loops could therefore classify the same connection close before the marker was visible to either loop.
- Implementation: the first terminal forwarding error now atomically claims shutdown before classification. Read-side close evidence remains recorded, while first or subsequent write-side `context.Canceled` and `net.ErrClosed` results are retained in `session.close` diagnostics without creating a blocking bridge error.
- Evidence and validation: Mineville Zeqa capture [captures/session-20260916T025636Z](https://github.com/BedrockDebugProxy-Captures/session-20260916T025636Z/releases/tag/capture-archive) at revision `9a7859c351b55b66a89cfa7e3448b809f5a48584` showed `bridge.write_error` racing a downstream context-cancelled termination. New shutdown race tests cover read-first and write-first ordering. A fresh stamped live report is required after this fix.
- Status: fix implemented; exact-revision live validation pending.

### FIX-0014 - Per-hop pre-spawn console summaries

- Symptom: after an Enchanted transfer, the capture contained hop 2 login and resource-pack packets but the terminal showed no corresponding `UPSTREAM` or `DOWNSTREAM` summaries.
- Root cause: the live reporter used one process-wide spawned flag. Once hop 1 reached gameplay, raw packet summaries were suppressed for every later transfer hop, including its pre-spawn exchange.
- Implementation: live output now tracks spawned state per hop, flushes the previous bucket when a new hop becomes active, and keeps the existing three-second summary cadence and capture path unchanged.
- Evidence and validation: Enchanted capture [captures/session-20260916T044421Z](https://github.com/BedrockDebugProxy-Captures/session-20260916T044421Z/releases/tag/capture-archive) recorded hop 2 resource-pack exchange and spawn while the operator reported missing terminal summaries. The regression test verifies pre-spawn summaries for two hops and continues to suppress post-spawn raw packets within each hop.
- Status: implemented; a fresh stamped The Hive baseline and the five-server release matrix are required because observable runtime logging changed.

### FIX-0015 - Concurrent URL resource-pack prefetch

- Symptom: after an Enchanted transfer, a backend with several URL-advertised resource packs could keep the upstream login exchange open long enough for the backend to kick the proxy client for `Timed out`, even though hop 2 eventually reached spawn.
- Root cause: the URL resource-pack adapter downloaded and parsed each advertised archive serially inside the upstream packet reader. The downstream listener could not receive the pack offer until every preceding URL download completed, extending the upstream-first buffering window.
- Implementation: URL pack downloads now use a bounded worker pool and commit successful results in the original `ResourcePacksInfo` order. Duplicate-offer suppression, UUID/version/size validation, URL removal from downstream copies, and the sequential downstream listener path are unchanged.
- Evidence and validation: Enchanted capture [captures/session-20260916T053620Z](https://github.com/BedrockDebugProxy-Captures/session-20260916T053620Z/releases/tag/capture-archive) selected `factions-spawn.factions.connect.enchanted.gg:19132`; hop 2 recorded seven URL pack archives, reached `session.spawned`, and then received the upstream `You were kicked: Timed out` message. The capture had zero dropped, truncated, decode, or write errors. The concurrency regression test uses a barrier-backed HTTP fixture to prove overlapping downloads and stable offer order.
- Status: implemented with focused automated tests; a new stamped live Enchanted transfer and The Hive baseline are required before release approval.

### FIX-0016 - Asynchronous bounded capture writer

- Symptom: Galaxite could send a dense `UpdateBlock` burst while the proxy synchronously wrote raw, decoded, and timing events. The forwarding path then stalled long enough for the client to receive `You moved too slow`, while the direct client did not show the same behavior.
- Root cause: every capture event held the recorder ordering lock through blob hashing, temporary-file publication, JSON encoding, and event-stream writes. High-volume server-to-client traffic therefore blocked unrelated client-to-server forwarding callbacks behind disk I/O.
- Implementation: capture events now receive their sequence at submission and enter one bounded ordered writer queue. The queue has blocking backpressure, a configurable record and byte budget, and a `never_drop` loss policy. `RecordReader` drains queued events before streaming a barrier-sensitive blob, `Close` drains accepted jobs before publishing the final manifest, and writer failures are exposed through `Recorder.Err` and the forwarding loop.
- Evidence and validation: Galaxite capture [captures/session-20260916T060203Z](https://github.com/BedrockDebugProxy-Captures/session-20260916T060203Z/releases/tag/capture-archive) at revision `320478f65db244c8b4d8be05e756e2c7774289db` retained 489,910 events and 93,790 blobs with no drops, but recorded 40,259 `UpdateBlock` packets and rare 596 to 738 ms capture/write stalls before the `You moved too slow` bridge write failure. The burst-order regression test verifies sequence continuity, blob integrity, manifest queue metadata, and closed-capture verification.
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

### FIX-0051 - The Hive Unfair Advantage ban investigation

- Symptom: a The Hive session through the read-only proxy received a temporary `Unfair Advantage` ban response even though the capture retained raw same-protocol traffic and showed no intentional gameplay mutation.
- Root cause: unresolved. The capture proves the server response, but it does not expose whether the trigger was a server-side blocklist or anti-cheat change, the `SrDuck84` upstream identity at the separate authenticated session boundary, client 1.26.51 behavior, or a combination.
- Implementation: none. This is an evidence and policy record, not a runtime fix.
- Evidence and validation: [`docs/research/hive-unfair-advantage-2026-09.md`](research/hive-unfair-advantage-2026-09.md), affected capture [session-20260920T135412Z](https://github.com/BedrockDebugProxy-Captures/session-20260920T135412Z/releases/tag/capture-archive), comparison capture [session-20260917T034859Z](https://github.com/BedrockDebugProxy-Captures/session-20260917T034859Z/releases/tag/capture-archive), Hive's current proxy policy, public Hive and Mojang release pages, and the community reports cited in the research note.
- Status: the technical cause remains unresolved. The maintainer later reported that Hive treated the punishment as a false positive and lifted it; the provided support-thread export confirms the review and lift but does not identify the trigger. The same `NhanAZ` to `SrDuck84` pattern appears in 22 earlier successful Hive captures, so the identity alone is not proven sufficient. Any future Hive session must follow the current read-only validation workflow and must not probe enforcement or imply a public policy change.

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

### FIX-0057 - Align authenticated ClientData metadata with Bedrock 1.26.51

- Symptom: the affected local Hive capture showed downstream ClientData game version `1.26.51` but authenticated upstream ClientData game version `1.26.50`, even though both sides used protocol 2193 and Android device metadata.
- Root cause: gophertunnel's authenticated Android login path assigns `protocol.CurrentVersion` to upstream ClientData, and the selected local protocol constant still advertised `1.26.50`.
- Implementation: update the selected game-version metadata to `1.26.51` while retaining protocol 2193 and the reviewed 1.26.50 packet definitions. Mojang's 1.26.51 release reports no packet changes.
- Evidence and validation: [`docs/research/protocol-1.26.51.md`](research/protocol-1.26.51.md), the affected closed capture, Mojang release commit `8667debe0adad27de6c692c88942c489a1193206`, focused metadata and proxy tests, and the project quality gate. This correction does not establish the cause of the Hive punishment and does not bypass any server enforcement.
- Status: implemented as a narrow metadata compatibility update. Authenticated vanilla-client and public-server validation remain pending and unauthorized by the current research record.

### FIX-0058 - Remove authenticated PlayFab identity from offline upstream logins

- Symptom: a 1.26.51 vanilla client connected through an unauthenticated local upstream produced a self-signed token with a `mid` claim copied from the downstream identity while its `ClientData.PlayFabId` remained the client value. The local ProxyFingerprint report flagged `auth_client_playfab_mismatch`; a direct vanilla login did not include `mid`.
- Root cause: the offline gophertunnel dial path cleared XUID and legacy title ID but retained the PlayFab identity fields introduced by the downstream authenticated login.
- Implementation: clear `IdentityData.PlayFabID` and `IdentityData.PlayFabTitleID` together with XUID and legacy title ID before calling `EncodeOffline`. The downstream display name, UUID, and ClientData remain available for the operator-controlled unauthenticated server.
- Evidence and validation: the local 1.26.51 ProxyFingerprint report, the `EncodeOffline` and `tokenClaims` source path, `TestClearXBLIdentityDataRemovesAuthenticationBoundFields`, and the focused proxy integration test. This change is limited to unauthenticated upstream dials.
- Status: implemented for local offline identity parity. It does not infer public-server anti-cheat behavior or bypass server policy.

### FIX-0059 - Preserve complete StartGame world metadata

- Symptom: a terminating bridge reconstructed the downstream `StartGame` packet from gophertunnel's compact `GameData` view, so fields not represented by that view were replaced with proxy defaults or omitted.
- Root cause: the upstream dialer decoded `StartGame` into `GameData`, which intentionally exposes only a subset of the packet. The bridge then called the listener's `StartGame`, whose serializer filled values such as generator, command, education, broadcast, level, and telemetry fields independently.
- Implementation: retain a copy of the complete decoded `packet.StartGame` in `GameData` for dialled connections. Listener `StartGame` reuses that packet when present and keeps the existing compact-data construction for callers that create `GameData` directly.
- Evidence and validation: the complete-packet regression test and the proxy loopback integration test use non-default `StartGame` fields and verify that the downstream client receives them unchanged. [`docs/research/start-game-parity.md`](research/start-game-parity.md) records the source boundary and remaining login-packet limitations.
- Status: implemented for decoded `StartGame` field parity. Dimension, Jigsaw, VoxelShapes, and ItemRegistry handling remain separate login-boundary behaviors and are not claimed to be byte-identical.

### FIX-0060 - Preserve startup Jigsaw and VoxelShapes packets

- Symptom: captures showed the proxy sending an empty `JigsawStructureData` and empty `VoxelShapes` downstream, followed later by the non-empty upstream packets. The client therefore received duplicate startup packets in a different boundary sequence from the server.
- Root cause: gophertunnel deferred these packet IDs during login without retaining them in `GameData`. The listener-side `StartGame` path always generated empty replacements, and the deferred upstream packets were later exposed to the bridge as normal traffic.
- Implementation: accept and retain both packet types across the bounded post-resource-pack login states. Listener `StartGame` now sends the retained decoded packets when available and keeps empty defaults only when the upstream did not supply them.
- Evidence and validation: The Hive and local loopback captures contain packet IDs 313 and 337 in the upstream login and two downstream occurrences before this fix. New unit and proxy loopback tests verify decoded contents and one startup delivery path. [`docs/research/start-game-parity.md`](research/start-game-parity.md) records the capture evidence and boundary.
- Status: implemented for the observed startup packet ordering. Packets arriving only after the connection is fully spawned remain normal forwarded traffic and require separate ordering evidence before changing the state machine.

### FIX-0061 - Preserve the upstream resource-pack stack

- Symptom: the upstream Hive capture carried a 1,449-byte `ResourcePackStack` with 24 texture-pack entries, while the downstream listener emitted a 1,446-byte reconstructed stack with 32 entries after adding gophertunnel's built-in exempted packs.
- Root cause: the dialled connection validated `ResourcePackStack` but discarded the complete packet. The listener then rebuilt a stack using the local resource-pack list, the current protocol version, a default experiment, and exempted packs.
- Implementation: retain a cloned `ResourcePackStack` in `GameData`, add a listener callback for an exact per-connection stack, and have the proxy supply the upstream stack before downstream resource-pack negotiation completes. The previous generated stack remains the fallback when no exact stack is available.
- Evidence and validation: [`docs/research/resource-pack-stack-parity.md`](research/resource-pack-stack-parity.md), the nested retention test, and the proxy loopback integration test with non-default stack metadata.
- Status: implemented for the exact stack boundary. `ResourcePacksInfo` offer-policy mirroring remains a separate documented limitation.

### FIX-0062 - Preserve upstream resource-pack offer metadata

- Symptom: the owner-controlled PocketMine capture [session-20260921T003441Z](https://github.com/BedrockDebugProxy-Captures/session-20260921T003441Z/releases/tag/capture-archive) carried an upstream `ResourcePacksInfo` payload whose `TexturePackRequired` flag was false, while the downstream listener-generated payload forced it to true.
- Root cause: the dialled connection decoded `ResourcePacksInfo` for pack delivery but discarded its offer metadata. The listener therefore used only its static required and vibrant-visuals policy.
- Implementation: retain a cloned upstream `ResourcePacksInfo` on dialled `GameData`, add a listener callback for per-connection offer metadata, and have the proxy preserve the metadata while rebuilding only the archive-backed `TexturePacks` entries that it can serve.
- Evidence and validation: the local capture above, the proxy loopback integration test with non-default optional-offer metadata, and the resource-pack negotiation research note.
- Status: implemented for decoded offer metadata. URL entries remain intentionally removed from downstream copies because the proxy delivers pack bytes over its own RakNet connection.

### FIX-0063 - Preserve raw dimension and item startup packets

- Symptom: the local capture [session-20260921T003441Z](https://github.com/BedrockDebugProxy-Captures/session-20260921T003441Z/releases/tag/capture-archive) contained an `ItemRegistry` packet with the same upstream and downstream size but a different wire hash. Decoding both packets produced equal semantic registries, so the difference was in typed re-serialization rather than the item values.
- Root cause: the terminating listener reconstructed `DimensionData` and `ItemRegistry` from decoded `GameData`. NBT maps inside item components can serialize in a different key order even when their decoded values are unchanged.
- Implementation: retain the complete raw packets and negotiated protocol ID during the dialled login exchange. The listener validates the packet ID and protocol before writing the retained bytes, and keeps the existing typed reconstruction as a fallback. Raw bytes are copied before the login buffer is released.
- Evidence and validation: the local capture above, the raw packet semantic probe showing 2,076 equal `ItemRegistry` entries, nested gophertunnel raw-retention and startup-writer tests, and the proxy loopback integration test that verifies exact downstream `ItemRegistry` bytes.
- Status: implemented for `DimensionData` and `ItemRegistry` when protocol IDs match. Other startup packets remain separate parity boundaries.

### FIX-0064 - Preserve raw world-start startup packets

- Symptom: the local startup probe found that typed re-encoding changed the wire bytes of `JigsawStructureData` and `StartGame` even when the decoded values were unchanged. `VoxelShapes` round-tripped identically in that capture, but remained part of the same startup boundary.
- Root cause: the terminating listener reused decoded packet structs for all startup world packets, allowing map serialization and optional-field encoding to diverge from the upstream bytes.
- Implementation: retain copied raw `JigsawStructureData`, `StartGame`, and `VoxelShapes` packets in `minecraft.GameData`. The listener validates packet IDs and the negotiated protocol before writing retained bytes and falls back to the decoded path when validation fails. Existing startup order and post-spawn forwarding are unchanged.
- Evidence and validation: focused probe of owner-controlled capture [captures/session-20260921T003441Z](https://github.com/BedrockDebugProxy-Captures/session-20260921T003441Z/releases/tag/capture-archive), nested raw-retention tests, and the root proxy loopback integration test that verifies exact bytes for all four retained world-start packet types.
- Status: implemented for the supported same-protocol pre-spawn boundary. Exact-revision real-client validation remains pending.

### FIX-0065 - Index per-connection NetworkSettings

- Symptom: the capture retained raw `NetworkSettings` packets, but `session.negotiated` did not expose the compression threshold, compression algorithm, or client-throttle values for either connection.
- Root cause: gophertunnel applied the settings directly to its encoder and decoder and then discarded the decoded packet. The listener and dialler therefore had no structured comparison surface.
- Implementation: retain a copy of the negotiated `NetworkSettings` on each connection and add bounded downstream and upstream metadata plus an explicit match boolean to `session.negotiated`. No wire behavior, compression policy, batching, or forwarding path is changed.
- Evidence and validation: owner-controlled capture [captures/session-20260921T003441Z](https://github.com/BedrockDebugProxy-Captures/session-20260921T003441Z/releases/tag/capture-archive) showed upstream threshold `1` and downstream threshold `256`, nested retention tests cover all fields, and the root loopback integration test verifies the structured comparison.
- Status: implemented as a capture-observability correction. Choosing a different downstream compression policy remains an open design question and is not inferred from this observation.

### FIX-0066 - Classify transport boundary state

- Symptom: transport payload blobs preserved application-batch bytes, but a capture could not show whether compression negotiation or packet encryption had completed for the connection that produced each payload.
- Root cause: the transport wrapper recorded operation and byte counts only. Gophertunnel applied NetworkSettings and encryption internally without exposing their current state to the capture callback.
- Implementation: add a per-payload metadata callback to the observed network wrapper, expose connection encryption state from the vendored connection, and attach bounded compression and encryption state to downstream and upstream transport events after their connection references become available. Raw payload blobs and existing event fields remain unchanged.
- Evidence and validation: the transport wrapper unit test verifies metadata and raw blob retention, the nested connection test covers the encryption boundary accessors, and the root loopback integration test observes metadata on both proxy channels. [`docs/research/capture-observability.md`](research/capture-observability.md) records the field meanings and remaining lower-layer boundary.
- Status: implemented as a capture-observability correction with no wire behavior change. Early login payloads may not have derived state because the connection has not exposed the relevant boundary yet.

### FIX-0067 - Avoid duplicate generated CreativeContent

- Symptom: owner-controlled capture [captures/session-20260921T005313Z](https://github.com/BedrockDebugProxy-Captures/session-20260921T005313Z/releases/tag/capture-archive) contained the upstream CreativeContent packet once at 49,283 bytes, followed downstream by a generated 2-byte CreativeContent packet and then the original 49,283-byte packet.
- Root cause: the terminating listener always emitted gophertunnel's empty CreativeContent compatibility packet when handling the downstream RequestChunkRadius. The upstream server packet could arrive after the spawn signal and was then forwarded separately.
- Implementation: add a per-GameData option for the terminating proxy to suppress only the listener-generated compatibility packet. The upstream CreativeContent packet remains on the normal forwarding and raw-capture path, and ordinary listener callers retain the existing default.
- Evidence and validation: the capture sequence above, the nested listener suppression test, and the root proxy loopback integration test that sends a non-empty upstream CreativeContent packet and matches its raw hash across both channels.
- Status: implemented for the observed terminating-bridge boundary. Direct vanilla-client validation remains pending. This change does not filter, mutate, or reorder an upstream packet.

### FIX-0068 - Capture transport state during login

- Symptom: transport payload events during the login sequence had no connection-derived compression or encryption state because the proxy stored each `*minecraft.Conn` only after the login boundary completed.
- Root cause: the observed transport wrapper records bytes before the listener or dialer returns its connection to the proxy, while gophertunnel did not expose the connection object at creation time.
- Implementation: add opt-in listener and dialer connection callbacks and use them to associate the connection with its hop before the first login packet. Existing per-payload metadata now reports the state that has actually been negotiated at that boundary, while the raw payload remains unchanged.
- Evidence and validation: the root loopback integration test now requires negotiated compression metadata on both channels before `session.negotiated`, proving that login-time payloads receive the state. The change is limited to capture observability and does not alter wire behavior.
- Status: implemented as a capture-observability correction. The first payload can still be `not_negotiated` before NetworkSettings is processed, and raw UDP or RakNet framing remains outside this observation boundary.

### FIX-0069 - Correct NetherNet capability research status

- Symptom: the ecosystem research note still described the observing network as RakNet-only and listed capability preservation as an unimplemented item, although the current wrapper is used with configurable transports and exposes the transport-specific packet methods.
- Root cause: the research note was not updated after the capability-preserving wrapper and transport selection path were implemented.
- Implementation: update the research note to distinguish confirmed wrapper behavior from the remaining unknown live NetherNet behavior. No runtime code or wire behavior changes.
- Evidence and validation: `internal/bedrock/network.go`, `internal/bedrock/network_test.go`, `internal/experience/target.go`, and the vendored NetherNet capability assertion show the supported path. The focused network tests pass.
- Status: documentation correction only. A live controlled NetherNet capture is still pending.

### FIX-0070 - Correct transport-state limitation metadata

- Symptom: newly opened manifests still claimed that transport payload compression and encryption state was not classified per event after the capture path began recording that state during login.
- Root cause: the CLI limitation string was not updated when transport metadata observability was implemented.
- Implementation: centralize the CLI limitation list and describe the narrower boundary that very early payloads may precede connection-state availability. Add a regression test for the wording.
- Evidence and validation: `internal/proxy/proxy_test.go` and `internal/bedrock/network_test.go` cover negotiated transport metadata, while the new CLI test prevents the stale limitation from returning. No forwarding or capture bytes change.
- Status: implemented as a capture metadata correction. Raw UDP and RakNet recovery frames remain outside the supported observation boundary.

### FIX-0071 - Ignore local runtime logs

- Symptom: the canonical validation build refused to stamp the current committed source because operator-created `proxy-agent-*.log` and `server.log` files appeared as unrelated untracked changes.
- Root cause: the repository ignored captures and build output but did not ignore the local runtime log names used by the documented development workflow.
- Implementation: add root-scoped ignore patterns for those generated logs. Existing files are preserved and are not removed or modified.
- Evidence and validation: the build script's dirty-tree guard reports these exact files before the change. No runtime code, capture data, or protocol behavior changes.
- Status: implemented as a workflow-only correction.

### FIX-0072 - Remove operator-specific LAN address examples

- Symptom: copy-paste validation and transfer examples used `192.168.1.10`, which may not exist on the operator's machine and causes the listener bind to fail.
- Root cause: documentation used one historical LAN address instead of a placeholder that requires the current host address.
- Implementation: replace the hardcoded address in README, validation, and release examples with `<LAN_IP>` while retaining wildcard and loopback examples where their semantics are intentional.
- Evidence and validation: the documented bind error occurs when an address is not assigned locally. This is documentation-only and does not change listener validation or network behavior.
- Status: implemented as a documentation correction.

### FIX-0073 - Compare parity lifecycle timing

- Symptom: the local ProxyFingerprint tool recorded lifecycle timing for each labeled session, but its comparison command ignored that data and reported only identity fields, packet order, and packet counts.
- Root cause: `ComparisonEngine` had no lifecycle comparison view even though `SessionRecord` retained ordered milestone timestamps and total duration.
- Implementation: ProxyFingerprint commit [`323aa03de`](https://github.com/NhanAZ/PocketMine-MP/commit/323aa03de) adds a versioned `lifecycle_timing` comparison containing duration deltas, event-prefix alignment, occurrence counts, and elapsed microseconds for each milestone. The original session reports remain unchanged.
- Evidence and validation: the plugin self-test and PHP syntax checks pass. The timing output is explicitly documented as an observation aid and does not infer server enforcement behavior.
- Status: implemented in the owner-controlled local parity tool. Existing reports can still be read, but comparisons generated before this change do not contain the new view.

### FIX-0074 - Bound local quality-gate memory use

- Symptom: running the full quality gate with the host's default Go parallelism can consume excessive memory on a development workstation.
- Root cause: `tools/quality.ps1` had no opt-in way to bound Go runtime workers and package build parallelism, even though the same checks are safe at lower concurrency.
- Implementation: add `tools/quality.ps1 -LowMemory`, which temporarily sets `GOMAXPROCS=2` and `GOFLAGS=-p=1` for child Go commands and restores the caller's environment afterward. The default and CI profiles remain unchanged.
- Evidence and validation: the low-memory quality gate, PowerShell syntax checks, and `git diff --check` pass. No runtime, protocol, capture, or binary behavior changes.
- Status: implemented as a local workflow safeguard.

### FIX-0075 - Preserve mixed resource-pack offer order

- Symptom: an upstream `ResourcePacksInfo` offer containing both URL-advertised and RakNet-delivered packs could lose an entry when the proxy replaced the completed upstream list with a cache-completion list.
- Root cause: `connectUpstream` treated `urlResourcePackCache.Packs()` as a complete resource-pack list, even though that method reports cache completion order and does not encode the complete `ResourcePacksInfo` offer order. The replacement could omit an entry that was not yet cached and also discarded the original offer order.
- Implementation: record the complete advertised cache-key order and merge prefetched URL packs with gophertunnel's normal resource-pack list. The merge prefers the validated archive for a URL key, retains non-URL entries, and preserves the offer order. Failed URL downloads still fall back to the normal gophertunnel pack entry and remain observable through existing warnings and raw evidence.
- Evidence and validation: source-path review of `internal/proxy/proxy.go` and `resource_pack_url.go`, `TestURLResourcePackCacheMergePreservesMixedOfferOrder`, and `TestRunnerPreservesMixedResourcePackOfferOrder`. The unit test exercises URL, RakNet, URL ordering, while the root loopback test verifies the listener's downstream metadata and stack order with a client configured to decline downloads. Focused proxy tests and the full low-memory quality gate must pass before commit.
- Status: implemented for mixed URL and RakNet offers. Real-client validation with an owner-controlled mixed offer remains pending.

### FIX-0076 - Correct resource-pack limitation metadata

- Symptom: newly opened manifests still claimed that the adapter could not mirror an upstream optional resource-pack policy after offer metadata preservation had been implemented.
- Root cause: the static limitation string was not updated with the `ResourcePacksInfo` retention and URL-archive boundary.
- Implementation: describe the preserved offer metadata and the intentional URL replacement with proxy-served archives. Failed URL retrieval remains observable through existing capture evidence and warnings.
- Evidence and validation: `internal/proxy/proxy.go` supplies retained offer metadata through the listener callback, `docs/resource-packs.md` documents the boundary, and `cmd/bedrock-debug-proxy/main_test.go` rejects the stale text and checks the corrected wording. No forwarding or packet bytes change.
- Status: implemented as a capture metadata correction. Mixed-offer real-client validation remains pending.

### FIX-0077 - Preserve validation report timestamps

- Symptom: `tools/create-validation-report.ps1` emitted capture `started_at` and `ended_at` using the host's locale after PowerShell parsed the analysis timestamps as `DateTime` values. The resulting report lost the original UTC suffix and fractional precision.
- Root cause: the report converted typed timestamps with the culture-sensitive `[string]` cast instead of formatting them explicitly.
- Implementation: add a timestamp formatter that emits invariant UTC ISO 8601 strings for `DateTime` and `DateTimeOffset` values, while preserving string inputs.
- Evidence and validation: the report generated from [captures/session-20260921T122228Z](https://github.com/BedrockDebugProxy-Captures/session-20260921T122228Z/releases/tag/capture-archive) previously contained `09/21/2026 12:22:28`; after the fix it must retain `2026-09-21T12:22:28.3304852Z` and the corresponding end timestamp. PowerShell syntax, the low-memory quality gate, and `git diff --check` are required before commit.
- Status: implemented as a deterministic validation-artifact correction. Existing reports are not rewritten in place.

### FIX-0078 - Bound advertised URL resource-pack bodies

- Symptom: the URL resource-pack adapter checked the advertised compressed size but the vendored URL reader could continue copying a response body that was larger than that size before metadata validation rejected it.
- Root cause: `resource.ReadURL` streamed the HTTP body directly into its temporary archive without a caller-provided byte limit. `Content-Length` is optional and cannot be treated as the response boundary.
- Implementation: add `resource.ReadURLLimited` and use the advertised pack size as the actual response-body limit. The reader rejects a declared or observed body larger than the limit before the archive is accepted, while the existing UUID, version, and size validation remains in place.
- Evidence and validation: the loopback URL-cache regression serves a chunked response containing a valid archive followed by extra bytes and confirms that no pack is returned. The vendored patch note and resource-pack documentation record the project-specific bounded reader. Full quality and diff checks remain required before commit.
- Status: implemented in the source tree with focused automated coverage. Real-client HTTP resource-pack validation remains pending.

### FIX-0079 - Cancel URL resource-pack fetches with the connection

- Symptom: an in-flight URL resource-pack fetch could continue after the downstream connection or upstream login context had been canceled, delaying shutdown and retaining HTTP and archive resources until the outer timeout.
- Root cause: the URL adapter checked its context before starting a download, but the vendored `resource.ReadURL` path used `http.Get` without a request context.
- Implementation: add `resource.ReadURLLimitedContext` and pass the per-connection cache context through the bounded URL reader. Request cancellation now stops the HTTP request and body read while preserving the existing raw offer and cache error evidence.
- Evidence and validation: the loopback cache regression holds an HTTP response open, cancels the cache context, and confirms the worker exits with a context-canceled error. The size-bound test and existing URL offer tests remain unchanged.
- Status: implemented in the source tree with focused automated coverage. Real-client HTTP resource-pack validation remains pending.

### FIX-0080 - Preserve per-pack resource-pack metadata

- Symptom: when a listener rebuilt downstream `ResourcePacksInfo` entries from archive-backed packs, it retained UUID, version, and size but dropped per-pack `ContentIdentity`, `SubPackName`, `HasScripts`, `AddonPack`, and `RTXEnabled` values from the upstream offer.
- Root cause: `Conn.handleClientToServerHandshake` cleared the retained offer's texture-pack list and reconstructed each entry from `resource.Pack`, whose archive object does not carry all wire-level offer metadata.
- Implementation: match served archives to the retained advertised entries by UUID, version, and compressed size, then copy the per-pack metadata while keeping downstream URL handling and archive delivery unchanged. Encrypted packs retain their current-connection content key, with the advertised key used as a fallback for archive entries without one.
- Evidence and validation: `TestRunnerPreservesMixedResourcePackOfferOrder` now supplies non-default metadata for URL and RakNet entries and verifies every preserved field across the proxy boundary. The resource-pack negotiation research note records the matching boundary.
- Status: implemented with focused automated coverage. Real-client encrypted and mixed-offer validation remains pending.

### FIX-0081 - Record client cache parity boundary

- Symptom: completed Windows client captures consistently showed `ClientCacheStatus=true` downstream and `false` upstream.
- Root cause: the listener learns the downstream cache value only after sending `ResourcePacksInfo`, while the upstream Dialer must send its value before receiving upstream `ResourcePacksInfo`. The proxy leaves `Dialer.EnableClientCache` at its false zero value.
- Implementation: add [`docs/research/client-cache-parity.md`](research/client-cache-parity.md) with raw packet evidence, the gophertunnel source path, and the startup ordering boundary. No runtime behavior is changed.
- Evidence and validation: [captures/session-20260921T122228Z](https://github.com/BedrockDebugProxy-Captures/session-20260921T122228Z/releases/tag/capture-archive) contains downstream ID 129 payload `01` and upstream ID 129 payload `00`; the closed session also records both booleans in `session.spawned`. This is a documentation-only research update.
- Status: confirmed parity difference, intentionally left as an explicit architecture boundary until an owner-controlled cache-enabled loopback can validate a safe redesign.

### FIX-0082 - Add protocol and local-source evidence for client cache ordering

- Symptom: the client-cache research note identified a startup dependency but did not cite the versioned Mojang packet reference or the owner-controlled server source that determines the local offer order.
- Root cause: the initial note relied on the vendored gophertunnel source and capture evidence only, leaving the protocol capability definition and local server sequencing implicit.
- Implementation: record the Mojang 1.26.50 protocol reference and the PocketMine `NetworkSession` and `ResourcePacksPacketHandler` source path in the client-cache research note. The note now distinguishes the confirmed capability definition and local ordering from platform-based inference.
- Evidence and validation: the official protocol reference describes packet 129 as a client-reported login capability, and the local PocketMine source sends `ResourcePacksInfo` before observing the client response. No runtime code, capture schema, or packet behavior changes.
- Status: implemented as a documentation-only provenance correction. A controlled cache-enabled local server test remains required before changing startup architecture.

### FIX-0083 - Forward the observed downstream chunk-radius request

- Symptom: an owner-controlled local capture showed downstream `RequestChunkRadius` `24/28` while the upstream connection sent the gophertunnel default `16/16`.
- Root cause: the upstream login and resource-pack exchange completes before the terminating bridge starts the downstream `StartGame` sequence, so the upstream dialer cannot synchronously wait for the downstream request without stalling login.
- Implementation: decode downstream packet ID 69 in the listener packet observer and write its typed `ChunkRadius` and `MaxChunkRadius` values to the paired upstream connection. The existing initial 16/16 request remains as the safe fallback. The in-tree gophertunnel dialer has an opt-in request provider used by deterministic loopback tests.
- Evidence and validation: `docs/research/chunk-radius-parity.md` records the capture payloads and timing boundary. The loopback integration test observes upstream 16/16 followed by 24/28, and focused malformed-decoder and dependency tests cover the new path.
- Status: implemented for observed request forwarding. The approved server radius after a later request remains server-controlled, and public-server validation is pending.

### FIX-0084 - Preserve chunk-radius forward failures

- Symptom: the listener packet callback attempted to forward a downstream `RequestChunkRadius` to the paired upstream connection but discarded a missing-connection or `WritePacket` error.
- Root cause: the request forwarding path was added as a best-effort callback with no capture error record when the upstream connection was unavailable or closed during the write.
- Implementation: retain the typed request forwarding helper and record `bridge.write_error` with the operation `forward RequestChunkRadius` when the write cannot be completed. The downstream raw packet and its original evidence remain unchanged.
- Evidence and validation: source review of the packet callback and the local chunk-radius parity capture identified the silent error path. Focused success and write-error tests cover typed value preservation and error propagation.
- Status: implemented as a capture-observability correction. The initial 16/16 request and the later-request ordering boundary remain unchanged.

### FIX-0085 - Preserve chunk-radius forwarding decode failures

- Symptom: the listener retained a raw downstream `RequestChunkRadius` body but silently returned when the strict typed forwarding decoder rejected truncation or trailing bytes.
- Root cause: the forwarding callback treated its validation error as a local control-flow result and did not emit a capture decode event.
- Implementation: record `packet.decode_error` with the packet ID, `RequestChunkRadius` type, operation, and underlying error before leaving the raw packet unchanged. The callback still does not mutate, suppress, or close the connection because of this observation-only failure.
- Evidence and validation: strict decoder tests cover truncation and trailing bytes; a recorder regression test verifies packet context and the manifest decode-error counter. The behavior is derived from the existing raw observer and capture schema.
- Status: implemented with focused automated coverage. The initial upstream 16/16 request and later-request ordering limitation remain documented in [`docs/research/chunk-radius-parity.md`](research/chunk-radius-parity.md).

### FIX-0086 - Preserve chunk-radius decode header context

- Symptom: the new malformed `RequestChunkRadius` decode event identified the packet type but omitted the sender and target sub-client header values that were already retained by the raw packet event.
- Root cause: the helper constructed a reduced `PacketInfo` from only the packet ID and typed name instead of carrying the full observed header context.
- Implementation: pass the complete `PacketInfo` into the decode-error recorder so the event preserves both sub-client routing bytes and decode status without changing the raw payload or forwarding decision.
- Evidence and validation: the capture schema defines `SenderSubClient` and `TargetSubClient` on `PacketInfo`, and the regression test now asserts non-zero values survive the event boundary.
- Status: implemented as a capture-context correction. No wire behavior, protocol ordering, or authentication behavior changes.

### FIX-0087 - Classify raw-forward decode failures

- Symptom: when same-protocol forwarding decoded a retained raw packet clone unsuccessfully, the bridge recorded the result as `bridge.read_error`, conflating a protocol decode failure with an underlying transport read failure.
- Root cause: the forwarding loop reused its generic read-error branch after `DecodePacketRead` returned an error, even though the original raw packet header and bytes were still available.
- Implementation: classify that branch as `packet.decode_error`, retain the packet ID and sub-client header context, and preserve the raw packet event. Transport errors and clone/storage errors continue using their existing bridge error kinds.
- Evidence and validation: the raw packet adapter exposes the original header separately from clone decoding, and deterministic recorder tests verify the decode-error event and manifest counter. No wire bytes or forwarding decisions change.
- Status: implemented with automated coverage. Real-client validation remains pending for the current stamped binary.

### FIX-0088 - Continue raw forwarding after derived decode failures

- Symptom: after a same-protocol packet clone failed to decode, the bridge recorded the failure but still terminated both forwarding loops even though the original raw packet remained available.
- Root cause: the raw forwarding loop treated every `DecodePacketRead` error as a terminal bridge error instead of separating the derived-view failure from the valid raw forwarding boundary.
- Implementation: record `packet.decode_error`, create a copied `packet.Unknown` derived view containing the packet ID and payload, and continue writing the original `PacketRead` through `WriteRawPacket`. Recorder failures, clone failures, and transport failures retain their existing terminal behavior.
- Evidence and validation: gophertunnel's packet decoder reports invalid trailing data while the raw adapter's clone decode returns the error and leaves the independently owned original header and payload available. The regression tests verify that boundary, the copied unknown fallback, and byte-preserving continuation through a loopback proxy session. No wire bytes are re-encoded or dropped.
- Status: implemented with automated coverage. Real-client validation remains pending for the current stamped binary.

### FIX-0089 - Preserve raw ResourcePackStack bytes

- Symptom: the downstream `ResourcePackStack` retained the upstream decoded values but could still be typed re-encoded, even when the complete upstream packet bytes were available.
- Root cause: `GameData` retained only the decoded stack and the listener callback had no raw packet or protocol boundary.
- Implementation: retain the complete raw stack and its source protocol ID, expose a narrow listener callback returning both values, validate the packet ID and protocol before writing the bytes, and keep the cloned typed stack as the fallback. The bridge still rebuilds `ResourcePacksInfo` entries from locally served archives.
- Evidence and validation: the gophertunnel regression tests verify exact raw response bytes and the proxy loopback integration test verifies decoded values and retained stack bytes across both proxy channels. The resource-pack parity research note records the source and boundary.
- Status: implemented with automated coverage. Real-client mixed and encrypted resource-pack validation remains pending for the current stamped binary.

### FIX-0090 - Use retained stack protocol provenance on the listener

- Symptom: the new raw `ResourcePackStack` path checked the listener's `GameData.RawProtocolID`, which is not populated by the listener callback before it sends the stack. A same-protocol connection therefore used typed re-encoding even when retained raw bytes were available.
- Root cause: the raw stack's source protocol ID was stored in dedicated listener connection state but the send condition still referenced the upstream-oriented `GameData` field.
- Implementation: select the raw path using `rawResourcePackStackProtocolID` and strengthen the regression fixture with an extra retained byte that typed serialization would not reproduce. Protocol mismatch and packet validation fallback behavior remain unchanged.
- Evidence and validation: the focused gophertunnel test now proves the output includes the retained trailing byte, and the proxy loopback integration remains green. This is a local source and test correction; no live server was contacted.
- Status: implemented with automated coverage. Real-client mixed and encrypted resource-pack validation remains pending for the current stamped binary.

### FIX-0091 - Forward the observed client cache capability

- Symptom: completed local vanilla-shaped captures advertised `ClientCacheStatus=true` downstream while the upstream Dialer sent `false`, even though both sessions used the same client capability boundary.
- Root cause: BedrockDebugProxy started the upstream login from the listener's dynamic resource-pack callback before the listener had consumed the downstream cache-status packet. The Dialer therefore used its false zero value.
- Implementation: add an opt-in listener mode that sends `PlayStatus(LoginSuccess)`, waits for `ClientCacheStatus`, and then runs the existing resource-pack callbacks. BedrockDebugProxy passes the observed boolean to `Dialer.EnableClientCache`; listeners without the mode retain the previous ordering.
- Evidence and validation: the client-cache parity capture and research note provide the confirmed wire difference and source ordering. The focused gophertunnel state-machine test covers both `ClientCacheStatus=false` and `true`, while the root loopback integration test enables the cache on its client, verifies the upstream connection reports it, and checks that `session.spawned` records parity on both sides.
- Status: implemented with automated local coverage, including a controlled loopback cache-blob and miss-response exchange. Real-client cache storage and reuse remain pending. Public-server policy or enforcement conclusions are out of scope.

### FIX-0092 - Validate client cache blob forwarding

- Symptom: the client-cache parity fix aligned the login capability but had no regression coverage for the later `LevelChunk` blob-cache and miss-response exchange.
- Root cause: the existing loopback fixture stopped after ordinary startup and gameplay packets, so it could not prove that cache-enabled metadata, client miss hashes, and server blob payloads survived the same-protocol raw forwarding path.
- Implementation: extend the owner-controlled loopback integration with a cache-enabled `LevelChunk`, a `ClientCacheBlobStatus` miss, and a `ClientCacheMissResponse`. Assert typed values, raw packet presence, and payload hash preservation on both channels.
- Evidence and validation: the deterministic proxy integration test exercises the complete local exchange and the capture scanner checks the new raw packet IDs. This does not model a real client's cache storage decisions or public-server behavior.
- Status: implemented with automated local coverage. A real vanilla-client local capture remains useful for cache hit and reuse behavior.

### FIX-0093 - Cancel early-login resource-pack work with the downstream connection

- Symptom: a downstream disconnect during login could leave an in-flight URL resource-pack fetch or upstream dial attached only to the proxy run context until the outer timeout.
- Root cause: the listener's dynamic resource-pack callback runs before `Listener.Accept` returns, while the observing network wrapper reports its context through `ConnectionContextFunc` only when it wraps the accepted connection. The proxy therefore saw no tracked downstream context during early login even though the listener `Conn` already owned the cancellable transport context.
- Implementation: prefer the tracked observed context after accept, and fall back to the listener connection's `Context()` during the login callback. This keeps upstream login and URL prefetch cancellation tied to the physical client without changing packet ordering or resource-pack selection.
- Evidence and validation: gophertunnel `Listener.createConn`, `Network.wrap`, and the proxy callback call order establish the timing boundary. Focused unit tests cover both the pre-`Accept` fallback and the post-`Accept` tracked context. The resource-pack documentation records the lifecycle boundary.
- Status: implemented with deterministic local coverage. No public server or external account was contacted for this change.

### FIX-0094 - Reject incomplete URL resource-pack offers before side effects

- Symptom: the URL resource-pack adapter could parse a valid prefix of a malformed or trailing `ResourcePacksInfo` body and start downloads even though the canonical packet decoder would later reject the complete body.
- Root cause: the auxiliary decoder checked for panics but did not verify that the packet reader consumed every byte.
- Implementation: require the `ResourcePacksInfo` reader to reach the end of the observed body before URL entries are scheduled. The raw packet observer remains unchanged, so malformed bytes and their decode outcome stay available as capture evidence.
- Evidence and validation: the strict decoder follows the existing `RequestChunkRadius` boundary check. A regression test appends a trailing byte to an otherwise valid empty offer and requires an error with no decoded value. No network or public server was contacted.
- Status: implemented with deterministic local coverage. URL-prefetch side effects are now gated on a complete typed offer.

### FIX-0095 - Preserve the NetherNet transport public key

- Symptom: the observing `net.Conn` wrapper preserved most optional transport methods but hid NetherNet's `PublicKey` capability from gophertunnel.
- Root cause: `observedConn` exposed `BatchHeader`, encryption state, packet reads, context, latency, and client GUID explicitly, but had no forwarding method for the authenticated transport key. Gophertunnel therefore could not perform its optional transport-to-login public-key comparison through the wrapper.
- Implementation: forward `PublicKey() *ecdsa.PublicKey` when the underlying transport provides it and return nil for transports without that capability. No packet bytes or authentication tokens are rewritten; gophertunnel's existing transport-to-login check is restored.
- Evidence and validation: the gophertunnel `publicKeyConn` boundary and go-nethernet `Conn.PublicKey` implementation establish the method contract. The bedrock network regression test now verifies pointer preservation through `observedConn`. No network or public server was contacted.
- Status: implemented with deterministic local coverage. A live NetherNet capture remains unverified.

### FIX-0096 - Retain transport status-ping responses

- Symptom: gophertunnel's dial and transfer paths received status-ping response bytes, including redirect-port metadata, but the observing network discarded them after returning from `PingContext`.
- Root cause: `Network.PingContext` delegated directly to the selected transport and had no capture event for the returned response or its failure timing.
- Implementation: record `transport.ping` with the requested address, response size, elapsed duration, operation error, and any returned response bytes. The event records the transport API boundary and does not claim raw UDP or RakNet request evidence.
- Evidence and validation: gophertunnel `Dialer.prepareDialAddress` consumes the returned pong to select a redirected port, while transfer routing uses the same network boundary. Network tests cover successful response blobs, partial response on error, and error metadata. No network or public server was contacted.
- Status: implemented with deterministic local coverage. Raw ping request and datagram-level evidence remain outside the current boundary.

### FIX-0097 - Capture transfer route-probe responses

- Symptom: the pre-reconnect transfer route probe called a RakNet transport directly, so its status-ping response bytes were discarded even though the later upstream dial used the observing `Network` boundary.
- Root cause: `probeTransferRoute` recorded only the route-probe outcome and constructed its transport outside `bedrock.Network.PingContext`.
- Implementation: route probes now use a `bedrock.Network` configured with the same local UDP address, client GUID, recorder, hop, and pending upstream connection identity. The existing `transfer.route_probe` lifecycle event remains, while the response and operation error are captured as `transport.ping` evidence.
- Evidence and validation: the transfer path retains the source address and client GUID before probing; the route-probe regression test uses a deterministic network stub and verifies both lifecycle and raw ping evidence. No network or public server was contacted.
- Status: implemented with deterministic local coverage. Raw UDP request and RakNet datagram-level evidence remain outside the current boundary.

### FIX-0098 - Preserve RakNet client GUID metadata

- Symptom: transport payload captures retained endpoints and Bedrock compression/encryption state but omitted the non-zero RakNet client GUID already exposed by the active connection.
- Root cause: the proxy's `transportMetadata` helper classified only compression and encryption state, even though gophertunnel exposes `RakNetClientGUID` for the same connection and transfer following already consumes it.
- Implementation: add optional `raknet_client_guid` metadata to transport payload events when the connection exposes a non-zero value. Other transports and early payloads omit the field; raw payload bytes and wire behavior remain unchanged.
- Evidence and validation: gophertunnel's `Conn.RakNetClientGUID` delegates to the transport's optional `ClientGUID` method, and the proxy already preserves that value for transfer routing. The local integration test now verifies GUID metadata on the dialled upstream channel; the listener-side connection may legitimately omit it when its transport does not expose a non-zero client GUID. No network or public server was contacted.
- Status: implemented as a capture-observability correction. The GUID is sensitive local transport identity metadata; raw UDP and RakNet recovery frames remain outside the current boundary.

### FIX-0099 - Record transport identity at connection boundaries

- Symptom: connection-open and connection-close events retained endpoints but not the optional RakNet GUID or public-key capability already visible on the underlying transport.
- Root cause: `Network.recordConnection` populated only generic lifecycle fields, while payload metadata was the first place to classify some connection state.
- Implementation: add bounded `data` metadata for a non-zero RakNet client GUID and for the presence of the optional public-key method and value. The public key itself is never serialized, and transport errors retain their existing lifecycle semantics.
- Evidence and validation: gophertunnel and the observing wrapper expose `ClientGUID` and `PublicKey` at the `net.Conn` boundary. The bedrock network regression test verifies both fields on `transport.connection_open`; no network or public server was contacted.
- Status: implemented as a capture-observability correction. The fields describe capabilities, not a transport fingerprint or raw datagram behavior.

### FIX-0100 - Document typed additional-packet boundary

- Symptom: the raw forwarding adapter's `PacketRead` API can return a typed packet without raw bytes after gophertunnel has consumed it during login or produced it as an additional packet during protocol conversion.
- Root cause: the dependency's login queue and conversion interface intentionally expose typed values after the original read boundary. The current protocol 2193 implementation uses identity conversion, so no additional converted packet is currently observed in the supported tree.
- Implementation: document the confirmed raw-observation boundary and the future multi-packet conversion limitation in [`docs/research/capture-observability.md`](research/capture-observability.md). No runtime behavior changes and no synthetic raw bytes are introduced.
- Evidence and validation: the in-tree `minecraft/raw_packet.go`, `minecraft/packet.go`, `minecraft/conn.go`, and `minecraft/protocol.go` source paths were reviewed. The current identity conversion and the packet-hook ordering are confirmed; future adapter usage remains unknown.
- Status: documentation-only research clarification. No live or public-server validation was performed.

### FIX-0101 - Document independent spawn acknowledgements

- Symptom: a local parity scan reported that the downstream server-to-client stream placed `ChunkRadiusUpdated` and `PlayStatusPlayerSpawn` before packets that appeared earlier on the upstream stream.
- Root cause: the proxy owns two independent login state machines. The dialled upstream connection consumes its own spawn acknowledgements, while the listener emits equivalent responses for the downstream client's separate `RequestChunkRadius` before ordinary forwarding begins.
- Implementation: update [`docs/research/start-game-parity.md`](research/start-game-parity.md) with the exact capture sequences, raw hashes, source handlers, and the independent-session classification. No runtime code, packet bytes, capture schema, or forwarding policy changed.
- Evidence and validation: owner-controlled capture [captures/session-20260921T122228Z](https://github.com/BedrockDebugProxy-Captures/session-20260921T122228Z/releases/tag/capture-archive) shows ID 70 and the spawn-status ID 2 with matching raw hashes on both channels. The in-tree `minecraft/conn.go` handlers and `internal/proxy/proxy.go:startGame` establish the consumption and generation boundary.
- Status: documentation-only clarification. A future state-machine change requires a deterministic fixture proving that consumed upstream acknowledgements can be substituted without delaying or duplicating the downstream spawn handshake.

### FIX-0102 - Mark client-cache mismatch evidence as historical

- Symptom: the client-cache research note described a pre-fix capture with downstream `true` and upstream `false` without clearly marking that the capture predates the deferred-offer implementation.
- Root cause: the note retained the original wire evidence but did not distinguish historical capture state from the current proxy path, allowing a later reader to mistake the old mismatch for a current regression.
- Implementation: label the capture as pre-FIX-0091, update the source-path description to name the current capability handoff, and state the current loopback parity evidence separately.
- Evidence and validation: FIX-0091 and FIX-0092 describe the implementation and deterministic cache exchange. This documentation-only correction does not change runtime behavior, packet bytes, or capture schema.
- Status: documentation-only provenance clarification. Real-client cache storage and reuse remain unverified.

### FIX-0103 - Mark chunk-radius mismatch evidence as historical

- Symptom: the chunk-radius research note used a capture created before FIX-0083 without explicitly stating that the later typed-forwarding path was absent from that capture.
- Root cause: the note preserved the original wire evidence but did not distinguish the historical initial-request mismatch from the current loopback behavior covered by the FIX-0083 regression test.
- Implementation: mark the capture as pre-FIX-0083, classify its values as historical, and link the deterministic 16/16-then-24/28 forwarding evidence in the note.
- Evidence and validation: FIX-0083 through FIX-0086 and the current `internal/proxy/proxy_test.go` loopback assertions establish the forwarding and failure-recording behavior. This documentation-only correction does not change runtime behavior, packet bytes, or capture schema.
- Status: documentation-only provenance clarification. The approved radius after a later request remains server-controlled and public-server validation is pending.

### FIX-0104 - Preserve equivalent raw resource-pack offers

- Symptom: the terminating listener retained decoded `ResourcePacksInfo` metadata but always serialized a new offer, leaving a byte-preserving gap for offers that required no URL or list transformation.
- Root cause: the upstream raw offer was not retained in `GameData`, and the listener had no safe way to select it after archive-backed pack preparation.
- Implementation: retain the complete raw `ResourcePacksInfo`, expose its source protocol through a narrow listener callback, and use the raw packet only after the prepared typed payload matches exactly. URL replacement, offer changes, protocol mismatches, and write failures keep the typed fallback.
- Evidence and validation: the existing resource-pack stack boundary and same-protocol raw-forwarding decision establish the safety constraints. Nested gophertunnel tests cover retention and exact reuse, and the root loopback integration test verifies downstream raw bytes. No capture schema or packet policy changed.
- Status: implemented with deterministic local coverage. Mixed URL and encrypted real-client resource-pack validation remains pending.

### FIX-0105 - Validate retained resource-pack stack fields

- Symptom: the listener selected retained `ResourcePackStack` bytes solely from protocol provenance, so a callback that prepared a different typed stack could still send stale retained fields.
- Root cause: the raw stack path validated only the packet ID and protocol. It did not compare the retained known fields with the stack selected for the current downstream connection.
- Implementation: decode the known stack fields before raw reuse and require them to match the prepared typed stack. Allow unread trailing bytes so fields introduced by a newer protocol revision remain observable and are forwarded unchanged. Use the existing typed path when the known fields differ or decoding fails.
- Evidence and validation: the listener callback boundary and retained-stack source path establish that the prepared value is the authority for the current connection. Focused nested tests cover exact raw reuse with trailing bytes and typed fallback for a changed stack. No capture schema or packet policy changed.
- Status: implemented with deterministic local coverage. Real-client validation remains pending for current stamped binaries.

### FIX-0106 - Clear recovered resource-pack cache failures

- Symptom: after a URL resource-pack download failed, a later RakNet fallback could store a valid matching archive while the per-connection cache still returned the stale URL error for that key.
- Root cause: `urlResourcePackCache.Store` replaced the pack map entry but did not remove the corresponding non-fatal error recorded by the URL prefetch worker.
- Implementation: delete the cached error when `Store` accepts a matching pack. The normal gophertunnel fallback remains unchanged, and the original URL failure stays visible in the operator warning and raw offer evidence.
- Evidence and validation: the `ResourcePackCache` contract defines load errors as non-fatal misses and successful stores as future cache hits. A focused regression test stores a valid pack after a synthetic prefetch error and verifies that `Load` returns the pack without an error.
- Status: implemented with deterministic local coverage. Real-client URL-failure and RakNet-fallback validation remains pending.

### FIX-0107 - Preserve transport and capture errors together

- Symptom: when an observed transport read or write returned an error while the capture recorder also failed, the wrapper returned only the capture error and a partial read could be reported as zero bytes.
- Root cause: `observedConn` replaced the underlying operation result whenever `recordPayload` failed, discarding the transport error and transferred byte count at the wrapper boundary.
- Implementation: return the original byte count and join the transport and capture errors for `Read`, `ReadPacket`, and `Write`. The raw payload and event error metadata remain unchanged.
- Evidence and validation: the transport wrapper is the shared boundary for all observed connections, and the recorder already stores the operation error in the event. A deterministic regression test uses a closed recorder plus independent read and write errors and verifies both errors and transferred counts remain visible.
- Status: implemented with deterministic local coverage. Live transport failure behavior remains unverified.

### FIX-0108 - Preserve connection-close errors during failed observation setup

- Symptom: when recording `transport.connection_open` failed, the wrapper closed the newly opened transport but returned only the capture error if the close also failed.
- Root cause: `Network.wrap` discarded the close result after the open-event recorder failure.
- Implementation: join the recorder and close errors while keeping the existing cleanup behavior.
- Evidence and validation: the wrapper owns the open-event failure path and the regression test uses a closed recorder plus an independent transport close error to verify both remain discoverable.
- Status: implemented with deterministic local coverage. Live transport-failure behavior remains unverified.

### FIX-0109 - Preserve typed Transfer rewrite failures without a nil raw read

- Symptom: a malformed or unreachable listener address during opt-in Transfer following could panic while handling the rewrite failure when the packet had already been decoded or the typed forwarding path was active.
- Root cause: the rewrite-error branch unconditionally called `rawRead.Release()`, even though `rawRead` is nil for typed forwarding and for raw reads that already exposed a decoded packet.
- Implementation: release the raw packet only when one is retained, then record `bridge.transfer_rewrite_error` and return the wrapped rewrite failure.
- Evidence and validation: the forwarding path sets `rawRead` to nil after a decoded raw read and never initializes it on typed forwarding. A deterministic regression test invokes the same error finalization with no raw packet and verifies the wrapped error and capture event.
- Status: implemented with deterministic local coverage. Live Transfer rewrite failures remain unverified.

### FIX-0110 - Propagate listener callback capture failures

- Symptom: a `RequestChunkRadius` decode or upstream-forwarding failure could be recorded unsuccessfully by the listener callback while the callback ignored the recorder error and allowed the session to continue.
- Root cause: gophertunnel's packet callback cannot return an error, but the proxy discarded the errors from `recordPacketDecodeErrorAtHop` and `recordNetworkErrorAtHop` instead of sending them to the run's `FailureSink`.
- Implementation: route both callback recorder errors through `setCaptureFailure`, which preserves the first capture error and cancels the run through the existing failure sink.
- Evidence and validation: the listener callback is the only boundary for these observed requests, and the recorder methods return failures when the capture stream is unavailable. A deterministic unit test verifies that callback recorder errors reach the failure sink without changing packet forwarding.
- Status: implemented with deterministic local coverage. Live capture-writer failure behavior remains unverified.

### FIX-0111 - Preserve lifecycle error-event write failures

- Symptom: `session.spawn_error`, `transfer.route_probe`, or `upstream.dial_error` could fail to enter the canonical capture while the underlying operation error remained the only visible result.
- Root cause: these error paths treated their lifecycle event as best effort and discarded the recorder error, even though the run already had a `FailureSink` for capture failures.
- Implementation: join lifecycle operation errors with recorder errors and send the capture error through the existing failure sink. Route-probe diagnostics remain non-fatal to the normal dial decision, but an incomplete route event now cancels the run.
- Evidence and validation: each path owns an explicit lifecycle event and receives either the run failure sink or a `Network` carrying it. Deterministic local tests cover route-probe capture failure and the shared callback propagation helper.
- Status: implemented with deterministic local coverage. Live recorder-failure behavior remains unverified.

### FIX-0112 - Preserve upstream close errors during setup failure

- Symptom: after resource-pack recording, downstream-pack preparation, or `upstream.connected` recording failed, `connectUpstream` closed the connection but returned only the original operation error.
- Root cause: the three cleanup branches discarded the observed connection's `Close` result, hiding transport or capture errors from the failed setup.
- Implementation: join the setup error with the connection close error through `closeAfterUpstreamError` while keeping cleanup ordering unchanged.
- Evidence and validation: `connectUpstream` owns the connection after a successful dial and all three branches close that connection before returning. A deterministic test verifies both operation and close errors remain discoverable.
- Status: implemented with deterministic local coverage. Live upstream cleanup failures remain unverified.

### FIX-0113 - Record the authenticated upstream identity boundary

- Symptom: a local server reported `XUID does not match (possible impersonation attempt)` when an authenticated Bedrock client connected through a proxy started with `--auth none`.
- Classification: confirmed local interoperability boundary, not a packet-forwarding defect. The offline upstream dial creates a new self-signed login and cannot carry the downstream Xbox identity into a server that verifies XUID.
- Evidence: closed capture [captures/session-20260921T121717Z](https://github.com/BedrockDebugProxy-Captures/session-20260921T121717Z/releases/tag/capture-archive) from stamped revision `d272805fd571525cca6f7473b166fb1129bafd48`. The upstream dial failed before `session.negotiated`; the error was recorded as `upstream.dial_error` after 1 minute 29.872 seconds. The existing authenticated boundary research and offline identity tests establish why copying or re-signing the authenticated identity would be unsafe.
- Action: document the boundary and the correct local test mode. Do not clear or forge identity checks to make an `auth none` upstream appear authenticated.
- Status: documentation-only clarification. No runtime behavior changed and no public-server validation was performed.

### FIX-0114 - Preserve typed Transfer capture failures without a raw packet

- Symptom: a recorder failure while writing the derived `packet.transfer_rewrite` view could call `Release` on a nil `PacketRead` and panic when `--follow-transfers` used typed forwarding or a raw read that had already exposed a decoded packet.
- Root cause: the forwarding path fixed nil cleanup for Transfer address rewrite failures but retained an unconditional raw-packet release in the preceding derived-view recording branch.
- Implementation: centralize the derived Transfer event write and release a retained raw packet only when one exists. The existing rewrite event and transfer error semantics remain unchanged.
- Evidence and validation: the raw-forward code explicitly sets `rawRead` to nil for typed packets and already-decoded reads. A deterministic regression test closes the recorder, invokes the derived-view path with no raw packet, and verifies that the recorder error is returned instead of panicking.
- Status: implemented with focused local coverage. Live Transfer capture-writer failure behavior remains unverified.

### FIX-0115 - Document deferred-login raw and decoded ordering

- Symptom: a local capture appeared to place a decoded `ServerBoundLoadingScreen` event after later raw packets, which could be mistaken for packet reordering or a missing raw-to-decoded pair.
- Root cause: gophertunnel invokes the raw packet hook when a packet arrives, then may defer the packet in its login state machine until the current expected state accepts it. The bridge consumes and records the decoded view later, while concurrent observer and bridge callbacks share one ordered recorder stream.
- Implementation: document the confirmed deferred-login ordering evidence and the confirmed, inferred, and unknown parts of the raw-to-decoded relationship in [`docs/research/capture-observability.md`](research/capture-observability.md). No packet bytes, forwarding policy, or capture schema changed.
- Evidence and validation: owner-controlled capture [captures/session-20260921T122228Z](https://github.com/BedrockDebugProxy-Captures/session-20260921T122228Z/releases/tag/capture-archive) at revision `d272805fd571525cca6f7473b166fb1129bafd48`, plus the in-tree `minecraft/packet.go`, `minecraft/conn.go`, and `minecraft/raw_packet.go` queue paths. The event stream and closed manifest were inspected locally.
- Status: documentation-only clarification. Raw and decoded events must not be paired by adjacent sequence numbers, and no public-server validation was performed.

### FIX-0116 - Retain only transferred bytes on partial writes

- Symptom: a partial transport write recorded a raw `transport.payload` blob containing the entire attempted buffer even though the transport reported that only a prefix was transferred.
- Root cause: `observedConn.Write` passed the caller's full payload to the recorder while separately storing the shorter transferred byte count.
- Implementation: slice the write payload to the reported transferred prefix before recording the raw blob. Keep `attempted_bytes`, `transferred_bytes`, and the underlying transport error unchanged.
- Evidence and validation: the `net.Conn` write contract permits returning fewer bytes with an error. A deterministic regression test uses a synthetic partial transport, verifies the blob contains only the transferred prefix, and verifies both byte-count metadata fields.
- Status: implemented with deterministic local coverage. Live partial-write behavior remains unverified.

### FIX-0117 - Account for accepted events lost after writer failure

- Symptom: when the asynchronous capture writer failed, it stopped after the failing job and left already accepted jobs in the queue without recording their loss in the manifest.
- Root cause: the writer returned immediately after setting `failed`, so queued jobs were neither persisted nor counted as dropped. A rollback racing a completed job could also reuse its committed sequence number.
- Implementation: drain accepted jobs after a writer failure, count the failed and drained jobs as dropped, add one explicit completeness limitation, and keep pending queue accounting consistent. Enqueue rollback now avoids rewinding a sequence that the writer already committed.
- Evidence and validation: deterministic recorder coverage closes the event stream, queues three accepted jobs, closes the recorder, and verifies zero persisted events, three dropped events, an incomplete manifest, and the explicit limitation. The focused capture tests pass under the Go race detector.
- Status: implemented with deterministic local coverage. Live disk-failure behavior remains unverified.

### FIX-0118 - Preserve physical flow identity during transfer overlap

- Symptom: a packet emitted by the old client connection after `Transfer` could receive the pending hop's downstream connection ID while the old physical flow was still active.
- Root cause: the listener observer used mutable pending state for every downstream callback. Transfer following creates that state before the old client flow has fully closed, so source and destination addresses were not used to select provenance.
- Implementation: retain connection ID and hop alongside every observed physical flow at transport-open time. Packet callbacks now match source and destination addresses to the flow that produced them, while retaining the existing dynamic fallback for callers without an identity-bearing flow.
- Evidence and validation: the loopback transfer regression sends `RequestChunkRadius` after the old hop's Transfer and verifies the downstream raw event remains `downstream-1`. A focused overlapping-flow observer test verifies both old and new identities. No packet bytes or transfer routing behavior changed.
- Status: implemented with deterministic local coverage. Real-client overlap timing remains unverified.

### FIX-0119 - Preserve omitted PlayFab ID in authenticated upstream ClientData

- Symptom: an authenticated local upstream report contained a synthetic `ClientData.PlayFabId` whose HMAC differed from the same session's authenticated token `mid`, while a direct vanilla report omitted the field. ProxyFingerprint classified the proxied session as `auth_client_playfab_mismatch`.
- Root cause: the forked gophertunnel dialer generated a random `PlayFabId` in `defaultClientData` before encoding every login, including authenticated dials where the downstream ClientData field was empty. After fallback generation was restricted to offline dials, the empty field was still serialized as `"PlayFabId":""`, which triggered a local parser warning even though the selected report treated it as absent.
- Implementation: make PlayFab fallback generation explicit and enable it only for offline dials. Authenticated dials preserve the downstream value, including an intentional empty field, and the ClientData JSON tag omits the key when the value is empty. No token claim, signature, device metadata, or forwarded packet bytes are rewritten.
- Evidence and validation: the closed local comparison reports from 2026-09-22, the `defaultClientData` and authenticated `login.Encode` source path, `TestAuthenticatedClientDataPreservesEmptyPlayFabID`, `TestAuthenticatedClientDataOmitsEmptyPlayFabIDFromJSON`, the first post-fix report from [captures/session-20260922T041204Z](https://github.com/BedrockDebugProxy-Captures/session-20260922T041204Z/releases/tag/capture-archive), and the follow-up report from [captures/session-20260922T043146Z](https://github.com/BedrockDebugProxy-Captures/session-20260922T043146Z/releases/tag/capture-archive). The follow-up has no `auth_client_playfab_mismatch` signal, records the field as absent, produces no PocketMine `PlayFabId` warning, and records `XUID match`.
- Status: implemented and validated on the owner-controlled local authenticated loopback. The authenticated Android metadata boundary and public-server policy remain outside this fix.

### FIX-0120 - Keep multiple Microsoft accounts behind one default run path

- Symptom: changing the Microsoft account required deleting the single shared `auth-token.json` cache, and adding a second account would have forced an account flag onto every normal `run` command.
- Root cause: the authentication cache had no local account registry, no selected default, and no explicit reauthentication operation. The provider's token lifetime was also easy to confuse with a local cache-retention policy.
- Implementation: add a small per-user account registry with the legacy cache preserved as the `default` account. Add top-level `login`, `accounts`, `use`, and `logout` commands. The first account becomes the default, ordinary `run --auth device` uses it automatically, and named commands are only needed when managing or switching accounts. Interactive `login` asks once for an optional local maximum cache age and treats Enter as no local limit. `login --max-age` remains available for scripts and forces device authentication after that age without changing Microsoft token handling.
- Evidence and validation: focused auth-cache tests cover legacy compatibility, account selection and replacement, safe names, local-age parsing, unchanged-cache timestamp preservation, interactive and non-interactive policy selection, and forced device authentication. Documentation records the storage paths, token privacy boundary, and official Microsoft token-lifetime references. Full quality-gate and manual device-login validation remain required before treating this runtime change as live-validated.
- Status: implemented with deterministic local coverage. Existing cached tokens and the Microsoft provider remain untouched unless the operator explicitly runs `login` or `logout`.

### FIX-0121 - Require a label for the first interactive account

- Symptom: a first device login could silently store the token under the synthetic `default` account, making it unclear which local profile would be used by later proxy runs.
- Root cause: the new account registry intentionally preserved the legacy `default` path, and the unlabeled login path reused that name even when no account had been configured.
- Implementation: prompt for an account username or local label during the first interactive `login` and first interactive `run --auth device`. A non-empty label is required and becomes the first default account. Existing installations with one or more accounts retain the low-friction Enter-for-default behavior. Non-interactive first runs fail with an actionable instruction instead of creating an unnamed profile.
- Evidence and validation: focused command tests cover the required prompt, shared input buffering, and the existing optional default prompt. Full quality-gate validation remains required before treating this runtime change as live-validated.
- Status: implemented with deterministic local coverage. Existing legacy caches remain compatible and are not renamed automatically.

### FIX-0122 - Rename legacy authentication caches explicitly

- Symptom: a legacy authentication file continued to appear as the ambiguous `default` account after named profiles were introduced.
- Root cause: compatibility kept `%AppData%\\BedrockDebugProxy\\auth-token.json` at the synthetic `default` label, but there was no way to migrate that cache without signing in again.
- Implementation: add `rename OLD_ACCOUNT NEW_ACCOUNT`. The command validates both labels, refuses an existing destination, moves the cache without changing token bytes, preserves the local age policy, and updates the selected default atomically with rollback on registry-save failure.
- Evidence and validation: account-store tests cover legacy migration, default preservation, token retention, and destination collision rejection. Full quality-gate validation remains required before treating this runtime change as live-validated.
- Status: implemented with deterministic local coverage. The label remains operator-supplied local metadata and is not verified against the token identity.

### FIX-0123 - Group and discover CLI help

- Symptom: the root help output listed every command in one flat block, making it difficult to find account-management versus capture-inspection workflows.
- Root cause: `printUsage` had no command categories, descriptions, or command-specific help route.
- Implementation: group the root command list by purpose and add `help [COMMAND]` with concise usage and behavior details for every command. The command remains read-only and does not alter proxy or capture behavior.
- Evidence and validation: focused CLI tests cover grouped help text, command-specific help, and unknown-command handling. Full quality-gate validation remains required before treating this documentation-facing runtime change as live-validated.
- Status: implemented with deterministic local coverage. Detailed flag descriptions remain on each command's existing flag set.

### FIX-0124 - Require CLI and documentation consistency checks

- Symptom: command help and documentation can drift when a command, flag, prompt, or user-visible comment changes independently of the implementation.
- Root cause: the agent guide required documentation updates for behavior changes but did not make root help, command help, and focused usage tests an explicit interface contract.
- Implementation: extend `AGENTS.md` with a CLI and documentation consistency rule. Public command syntax, prompts, defaults, exit behavior, examples, and user-visible comments must be updated together with help, focused documentation, and tests.
- Evidence and validation: documentation review and `git diff --check`. This is a workflow-only change with no runtime or capture impact.
- Status: implemented as a documentation-only maintenance rule.

### FIX-0125 - Restore CubeCraft as a maintainer-risk-gated release target

- Symptom: the release workflow and readiness checker omitted CubeCraft even when the maintainer wanted the private, owner-controlled test process to include it.
- Root cause: the workflow treated CubeCraft's public Bedrock proxy prohibition as an unconditional project exclusion instead of separating the public policy warning from the maintainer's internal risk decision.
- Implementation: add CubeCraft as the sixth release-matrix label with its direct endpoint recorded as an initial reference, require an explicit `risk_acknowledged=pass` check, retain the public policy warning, prohibit recording private server discussions, and keep the normal read-only capture path. No runtime forwarding or packet behavior changed.
- Evidence and validation: CubeCraft's [Rules](https://help.cubecraft.net/en/article/cubecraft-rules-1403lij/) and [Allowed Mods and Clients](https://www.cubecraft.net/threads/allowed-mods-and-clients.228596/) were reviewed on 2026-09-22. This change updates the release docs, agent guide, validation catalog, changelog, and readiness checker only.
- Status: implemented as a documentation and release-workflow change. No CubeCraft connection was performed and no permission or policy exception is claimed.

### FIX-0126 - Surface the selected account in run output and viewer headers

- Symptom: after multiple Microsoft accounts were added, the terminal and capture viewer did not identify which selected account started a session.
- Root cause: account selection existed only in the local account store and was not copied into capture metadata or the viewer overview.
- Implementation: record the selected account label as `options.values.account_name`, print it when an authenticated run starts, and add an Account metric to the viewer header. The value is local operator metadata and is never a token or identity verification claim.
- Evidence and validation: focused viewer coverage verifies the manifest value and Account header marker. The CLI path uses the same default account already selected by `run`. Full quality gate and browser smoke validation remain required for the viewer-facing change.
- Status: implemented with deterministic local coverage. Real-client account identity parity remains outside this presentation change.

### FIX-0127 - Select the newest event during live follow

- Symptom: Follow live moved the viewer to the newest page, but the table still showed that page from its first row and did not identify the newest packet in the live stream.
- Root cause: live refresh used the API tail offset but preserved the current event selection instead of selecting the last row returned by the tail query.
- Implementation: when Follow live is enabled, each successful refresh selects the last matching row returned, loads its details, and scrolls that row into the event table viewport. Repeated refreshes avoid reloading details when the newest sequence is unchanged.
- Evidence and validation: viewer JavaScript coverage checks the newest-event hook, while browser smoke validation must append complete live events and confirm the newest row remains selected and visible without changing filters.
- Status: implemented with viewer-only impact. The canonical capture and packet forwarding paths are unchanged.

### FIX-0128 - Keep page scroll stable during live follow

- Symptom: while inspecting Decoded view - Event details, each Follow live refresh moved the page back toward the event table and interrupted manual scrolling.
- Root cause: newest-event selection used `scrollIntoView`, which was allowed to scroll the page-level document in addition to the bounded event table.
- Implementation: scroll the newest or keyboard-selected row by adjusting only the event table wrapper's `scrollTop`, leaving the page and inspector scroll position unchanged.
- Evidence and validation: viewer JavaScript coverage requires the table-only scroll hook. Browser smoke validation must keep the details panel in view while a live event is appended and confirm that the page scroll position remains stable.
- Status: implemented with viewer-only impact. The selected newest row remains visible inside the event table without moving the outer page.

### FIX-0129 - Preserve details layout during packet-heavy live follow

- Symptom: on a high-rate capture, Follow live repeatedly collapsed Decoded view - Event details to a loading line before expanding it again, so the page scrollbar moved and manual inspection was interrupted.
- Root cause: each newest-event request cleared the details DOM before its response arrived. The resulting layout changes triggered browser scroll anchoring while the table was also being refreshed.
- Implementation: keep the current details and blob preview visible while Follow live loads the next event, then restore the page scroll position after rendering the replacement. Manual event selection keeps the existing loading feedback.
- Evidence and validation: viewer JavaScript coverage requires the preserve-details and page-scroll hooks. Browser smoke validation must append events at a packet-heavy cadence while the details panel is in view and confirm its content does not collapse or move the outer page.
- Status: implemented with viewer-only impact. The selected event can change while its previous details remain visible until the new response is ready.

### FIX-0130 - Report progress during closed-capture verification

- Symptom: after automatic artifacts finished, the terminal only printed a single verification message and appeared idle for several minutes while a large capture was scanned. The live viewer also had no indication that verification was still active.
- Root cause: capture verification was synchronous and exposed no progress callbacks. The run command and live viewer therefore had no intermediate state to publish after the recorder and artifact writer closed.
- Implementation: add coarse verification milestones for scanned events and verified blobs, print them to the console, and publish the same counters through the live viewer session API. The run viewer opts into external verification ownership so HTTP polling does not start a competing blocking verification pass. The viewer keeps polling until verification completes, then returns to its normal event status.
- Evidence and validation: capture verification tests cover the start, manifest, and completion milestones. Viewer tests cover externally owned progress and the final verified state. Browser smoke validation must observe progress text in a live viewer while a large closed capture is verified and confirm that the terminal remains active.
- Status: implemented with capture verification and viewer status impact. The capture schema and verified result remain unchanged.

### FIX-0131 - Accept large client cache blob status lists

- Symptom: the Galaxite release capture recorded one `ClientCacheBlobStatus` decode error when the client sent 1,302 miss hashes. The session otherwise connected, spawned, forwarded traffic, and completed artifacts.
- Root cause: the vendored gophertunnel protocol helper applied its generic 1,024-entry slice safety bound to both cache hash lists. The Bedrock packet schema has VarInt counts and the reviewed BedrockProtocol implementation does not define a 1,024-entry protocol limit.
- Implementation: add a bounded `FuncSliceOfLenWithMaxLen` helper and use a packet-specific 16,384-entry bound for `ClientCacheBlobStatus`. All ordinary slices retain the 1,024-entry bound. Raw packet capture and decode-error reporting remain unchanged.
- Evidence and validation: Galaxite capture `fe9d3e3b-aa80-4976-a6ec-31db99dd1db1`, [`docs/research/galaxite-client-cache-blob-status-limit.md`](research/galaxite-client-cache-blob-status-limit.md), and `TestClientCacheBlobStatusAllowsObservedHashCount`.
- Status: implemented with focused protocol regression coverage. A fresh stamped Galaxite capture and the full release matrix are required before release approval.

### FIX-0132 - Recover repeated RakNet connection acceptance

- Symptom: a Galaxite upstream dial aborted with `unexpected additional CONNECTION_REQUEST_ACCEPTED packet` after the dialer had already sent `NEW_INCOMING_CONNECTION`.
- Root cause: the dialer treated a repeated acceptance as fatal. RakNet acceptance is not itself reliable, so a server can retransmit it when the first client response was lost or delayed.
- Implementation: re-send `NEW_INCOMING_CONNECTION` with the repeated acceptance timestamp and keep the established connection open. The change is isolated to the vendored go-raknet dialer and does not alter Bedrock packet payloads.
- Evidence and validation: capture `1de1bf87-bc59-47fe-949d-6f0eee6ea9fc` at revision `730e6b972dd16c21a94c3b9c2e2119d68e2ea7ea`, the immutable upstream handler at [`ea813dc668b5a2a2767cc5577bf77869c965f27a`](https://github.com/Sandertv/go-raknet/blob/ea813dc668b5a2a2767cc5577bf77869c965f27a/handler.go), and `TestDialerResendsNewIncomingConnectionAfterRepeatedAcceptance`.
- Status: implemented with deterministic transport coverage. A fresh stamped Galaxite session is required. The later retry `9aeeab8b-46ac-499f-b42f-68b328b7a883` failed independently during the Minecraft authentication-chain request and does not validate or invalidate this transport fix.

### FIX-0133 - Allow slow active resource-pack exchanges to finish

- Symptom: the Galaxite session continued receiving and requesting resource-pack chunks for five minutes, then the proxy timed out before upstream login completed.
- Root cause: the absolute five-minute per-hop deadline was shorter than the observed 460-chunk exchange. The capture showed 186 chunks arriving with ongoing requests when the deadline fired, so the connection was active rather than stalled.
- Implementation: increase the bounded upstream login and resource-pack deadline to fifteen minutes. Packet ordering, chunk contents, and capture error reporting remain unchanged.
- Evidence and validation: Galaxite capture [captures/session-20260922T151050Z](https://github.com/BedrockDebugProxy-Captures/session-20260922T151050Z/releases/tag/capture-archive) at revision `a992d6c37f07ae04afb651ce2130712f9e4b209e`, [`docs/research/galaxite-resource-pack-flow.md`](research/galaxite-resource-pack-flow.md), and the existing timeout error coverage.
- Status: implemented with documentation updates. A fresh stamped Galaxite session must reach upstream login, spawn, and shutdown before release approval.

### FIX-0134 - Pipeline upstream resource-pack chunk requests

- Symptom: an active featured-experience resource-pack exchange requested one chunk, waited for its response, and left the server's available response window idle. The fifteen-minute deadline fixed the premature timeout but not the serial transfer cost.
- Root cause: the pinned gophertunnel upstream download loop was serial. The listener's sequential downstream delivery is a separate path and was not the cause of the upstream wait.
- Implementation: allow up to 100 validated upstream chunk requests in flight for one pack, replenish the window as responses arrive, and retain out-of-order fragments in a bounded reorder buffer. Validate chunk indices, request membership, bounds, and exact lengths, and clone response bytes before buffering. Keep downstream sequencing and pacing unchanged.
- Evidence and validation: Galaxite capture [captures/session-20260922T151050Z](https://github.com/BedrockDebugProxy-Captures/session-20260922T151050Z/releases/tag/capture-archive), the source and implementation notes in [`docs/research/resource-pack-pipeline.md`](research/resource-pack-pipeline.md), [gophertunnel PR #509](https://github.com/Sandertv/gophertunnel/pull/509), and `resource_pack_pipeline_test.go`. A fresh stamped The Hive baseline and Galaxite session are required because upstream transfer scheduling changed.
- Status: implemented with focused nested gophertunnel coverage. Real-server validation remains pending.

### FIX-0135 - Accept implicit resource-pack directory declarations

- Symptom: CubeCraft encrypted resource packs recorded `resource_pack.decrypt_error` events because `contents.json` declared directories such as `animation_controllers` and `texts` that were not present as explicit ZIP directory entries.
- Root cause: the decoder required every manifest path to match a file-map entry. ZIP archives commonly omit directory entries, while the manifest still lists directories before their keyed child files.
- Implementation: treat a manifest path as an implicit directory only when the archive contains descendants below it and the directory entry has an empty key. Keep child-file keys unchanged, exclude directory declarations from file counts, and continue rejecting genuinely missing paths or keyed directories.
- Evidence and validation: CubeCraft capture [captures/session-20260922T165238Z](https://github.com/BedrockDebugProxy-Captures/session-20260922T165238Z/releases/tag/capture-archive) at revision `3a77b024ef44f37479f925ed1cba0aacddffc448`, the encrypted archive directory listings, `TestDecryptArchiveAllowsImplicitDirectoryDeclarations`, `TestDecryptArchiveRejectsMissingDirectoryDeclaration`, `TestDecryptArchiveRejectsKeyedDirectoryDeclaration`, and a temporary direct decrypt check over all five captured CubeCraft packs.
- Status: implemented with focused resource-pack regression coverage. A fresh stamped six-server matrix is required before release approval.

### FIX-0136 - Preserve best-effort transfer probe failures without blocking validation

- Symptom: an Enchanted transfer session reached hop 2 and spawn, but its best-effort route probe and the actual dial status ping recorded DNS timeout errors as blocking validation failures.
- Root cause: route-hint pings were retained as ordinary transport errors even though the route probe is explicitly non-fatal and the normal dial continued to a successful upstream connection.
- Implementation: mark route-hint pings as best effort in their structured metadata, retain their operation errors, render them as warnings, and classify only those marked errors as non-blocking in sanitized reports. Other status-ping failures remain blocking.
- Evidence and validation: Enchanted capture [captures/session-20260922T164833Z](https://github.com/BedrockDebugProxy-Captures/session-20260922T164833Z/releases/tag/capture-archive) at revision `3a77b024ef44f37479f925ed1cba0aacddffc448`, where hop 2 connected and spawned after the 750 ms DNS timeout, plus focused network, proxy, analysis, and report-generation checks.
- Status: implemented with explicit evidence classification. A fresh stamped Enchanted session is required before release approval.

### FIX-0137 - Capture upstream resource packs without a local listener

- Symptom: capturing the packs offered during an upstream login required starting a local proxy listener and joining it with a separate Minecraft client.
- Root cause: the resource-pack capture path was only reachable through `Runner.Run`, which accepts a local client before dialing upstream. In the pinned gophertunnel fork, `DialContextNetwork` also waits for the full client login state, which is reached only after world startup. Skipping an explicit `Conn.DoSpawn` call alone therefore does not return at the resource-pack boundary.
- Implementation: add `run --packs-only` and a runner path that uses the existing upstream dialer, packet observer, URL-pack adapter, archive recorder, and optional in-session decryption. Add an opt-in gophertunnel dial completion signal after the server `ResourcePackStack` is accepted and the final `PackResponseCompleted` is flushed. Packs-only closes at that boundary without starting the proxy's local listener or requesting world spawn. The capture explicitly records the mode and its lack of world-spawn evidence.
- Evidence and provenance: local loopback regressions `TestRunPacksOnlyCapturesUpstreamWithoutListener` and `TestDialerResourcePackExchangeOnlyStopsBeforeWorldStartup` exercise the one-pack Bedrock exchange and a server `StartGame` immediately after the pack boundary. They verify the archive SHA-256, capture integrity, and absence of client world-spawn packets. The external references are gophertunnel `v1.61.0` at `283a5a97dfe65da94bcc0b401807f6aefa9e72ee`, bedrocktool revision `85d5cfe1545c8d853be2859144a8357539ffc0f2`, and the maintainer-authorized private EasySpecter revision `6939f8aa1ad70fc694d370e3b4b33e6d3e49482c`. No third-party source was copied. See [`docs/research/resource-pack-fetch-mode.md`](research/resource-pack-fetch-mode.md).
- Status: the focused loopback test passes; the full project quality gate must be rerun after the final diff. No public-server session is implied by that test. A fresh stamped The Hive baseline and the complete release matrix are required before release approval.

### FIX-0138 - Synchronize concurrent upstream resource-pack state

- Symptom: a Hive run terminated with `fatal error: concurrent map writes` in `Conn.handleResourcePackDataInfo` while resource-pack responses were arriving.
- Root cause: the per-pack download goroutine removed completed entries from `awaitingPacks` under `Conn.packMu`, while the next `ResourcePackDataInfo` handler inserted into the same map without the mutex. The bounded download pipeline permits these operations to overlap.
- Implementation: guard the `downloadingPacks` lookup and the atomic move into `awaitingPacks` with `Conn.packMu`. Add a regression test that starts many pack-info handlers concurrently and verifies every pack reaches the awaiting state without corrupting either map.
- Evidence and validation: maintainer-provided Hive console trace dated 2026-09-23, the handler and downloader critical sections in vendored gophertunnel, `TestResourcePackDataInfoTransitionsConcurrentPacksSafely`, and focused nested-module tests including `go test -race`. No capture or account data was added to the repository.
- Status: automated focused validation passes. A newly built binary must complete a fresh The Hive session before this runtime change can be treated as live-validated or used to continue release approval.

### FIX-0139 - Share the upstream resource-pack request window

- Symptom: the `0.3.0-rc.3` Galaxite capture sent 460 `ResourcePackChunkRequest` packets in about 10.5 seconds but recorded only 97 `ResourcePackChunkData` responses during the five-minute session. The dial ended with `context canceled`, not the configured timeout, and packet decoding reported no errors.
- Root cause: the 100-request bound was local to each pack downloader. Multiple `ResourcePackDataInfo` packets start independent downloaders, so their windows could multiply across one connection and create a 460-request burst.
- Implementation: add one connection-wide 100-slot request window. A downloader acquires a slot before writing a chunk request and releases it when the matching response is accepted or the write fails. Per-pack ordering, bounded reorder buffers, response validation, downstream sequential delivery, and pacing remain unchanged.
- Evidence and provenance: captures [captures/session-20260923T044137Z](https://github.com/BedrockDebugProxy-Captures/session-20260923T044137Z/releases/tag/capture-archive) at revision `1ede568486768047acc5d52b721f595b50907fcc` and [captures/session-20260922T153826Z](https://github.com/BedrockDebugProxy-Captures/session-20260922T153826Z/releases/tag/capture-archive) at revision `29fbfd8e0664be20cb3929f5916f961e4e07b66c`; [`docs/research/resource-pack-pipeline.md`](research/resource-pack-pipeline.md); and the bounded client request window described by [gophertunnel PR #509](https://github.com/Sandertv/gophertunnel/pull/509). The request burst is confirmed. The server-side reason for its low response count remains inferred.
- Status: focused tests verify that multiple packs share the cap and that accepted responses replenish it. Exact-revision The Hive and Galaxite live validation remain pending.

### FIX-0140 - Bound larger client packet batches

- Symptom: the CubeCraft session reached spawn, then closed after the downstream listener reported `decode batch: number of packets exceeds max=812` for a client-originating batch.
- Root cause: the vendored decoder's default 812-packet limit is enabled for listener connections. The proxy's client-facing listener inherited that server-oriented cap, so one accumulated client batch above 812 was rejected before normal forwarding.
- Implementation: make the decoder's positive packet-count limit configurable and expose it through `minecraft.ListenConfig`. Keep 812 as the default and retain the listener's finite check. Set only the proxy's client-facing listener to 16,384 packets per batch; do not change dialer behavior or packet bytes.
- Evidence and provenance: local closed CubeCraft capture [captures/session-20260923T104518Z](https://github.com/BedrockDebugProxy-Captures/session-20260923T104518Z/releases/tag/capture-archive), tested revision `bfc19102a312cbe5b0a18c9432c37afcbb446b24`, and `docs/research/client-batch-packet-limit.md`. The local decoder rejection is confirmed. Why the client accumulated this batch is unknown, and the capture does not establish a server kick or ban.
- Status: focused nested gophertunnel tests and the full low-memory project quality gate pass. Fresh stamped The Hive and CubeCraft sessions are required to establish live behavior; the six-server release matrix remains separate.

### FIX-0141 - Release the RakNet source address before a transfer probe

- Symptom: Linux release CI could not bind the transfer route probe's local UDP address, and the following next-hop login failed on the same address.
- Root cause: the normal RakNet close waits for the peer to acknowledge closure. Transfer following immediately probes and dials the next hop using that connection's source address, so the previous socket can still own the port.
- Implementation: add an explicit immediate-close capability through go-raknet, gophertunnel, and the observing connection wrapper. Use it only after forwarding `Transfer`, and strengthen the loopback test to require a successful route probe.
- Evidence and provenance: GitHub Actions run `35946284078` for commit `18e66b72daf1e01d9aa9a6ee0cfb43d68be31a52` showed `bind: address already in use` on Linux. The transfer decision and Enchanted handshake note already require reuse of the same local UDP address and client GUID. The focused loopback transfer test and go-raknet close test pass after the correction.
- Status: Windows CI passed; the Linux quality job found an independent network-settings metadata race tracked by FIX-0142. The transfer fix awaits the next exact-revision CI run and live validation. Fresh stamped The Hive and the full six-server release matrix remain required.

### FIX-0142 - Synchronize negotiated network settings metadata

- Symptom: the Linux race-enabled quality job reported concurrent access to `Conn.networkSettings` while an upstream packet reader handled `NetworkSettings` and the proxy captured per-payload connection metadata.
- Root cause: the dial-side handler assigned the settings pointer while the public getter read and copied it without synchronization.
- Implementation: protect both dial-side and listener-side settings publication with a read-write mutex, and return a value copy under a read lock. Add a concurrent getter and negotiation regression test.
- Evidence and provenance: GitHub Actions run `35947403349` for commit `d8724a45feed4889e92616fdb845ec9f2d2f50fa` reported the race in `TestRunnerForwardsBidirectionalPacketsAndCapturesSession` and `TestRunPacksOnlyCapturesUpstreamWithoutListener`. The write and read paths are `Conn.handleNetworkSettings` and `Conn.NetworkSettings` in the in-tree gophertunnel fork.
- Status: the focused concurrency test and full local race-enabled quality gate pass. Exact-revision CI remains pending. The fix changes synchronization only; packet contents and negotiation order remain unchanged.

### FIX-0143 - Classify remote disconnect writes as peer termination

- Symptom: Enchanted captures recorded a server-to-client `Disconnect` packet and closed the upstream connection, then classified later writes to that closed connection as blocking `bridge.write_error` events.
- Root cause: gophertunnel surfaces a received Bedrock `Disconnect` as a typed `minecraft.DisconnectError`. The bridge shutdown classifier handled context cancellation and closed sockets but not this explicit remote-close error. The downstream `RequestChunkRadius` callback also recorded the same typed close as a forwarding failure.
- Implementation: recognize wrapped `minecraft.DisconnectError` as a peer-termination cause when classifying bridge shutdown and omit the downstream chunk-radius callback error when it is caused by that same remote close. Raw packet evidence and session-close diagnostics remain available; other write failures remain blocking.
- Evidence and provenance: local captures [captures/session-20260924T152209Z](https://github.com/BedrockDebugProxy-Captures/session-20260924T152209Z/releases/tag/capture-archive) at revision `bfe0741fa4cc9fb74199bde0624ce38c16d4af76` and [captures/session-20260923T062327Z](https://github.com/BedrockDebugProxy-Captures/session-20260923T062327Z/releases/tag/capture-archive) at revision `bfc19102a312cbe5b0a18c9432c37afcbb446b24` each contain upstream packet ID 5 before connection closure and later bridge write errors. The packet message text was not inspected. The handling is based on the vendored gophertunnel `Conn.receive` and `Conn.closeErr` paths.
- Status: focused regression tests cover wrapped remote-disconnect errors and first-write shutdown classification. Fresh stamped Enchanted and The Hive sessions are required; this does not prevent a server from disconnecting the session or make a failed manual normal-session check pass.

### FIX-0144 - Bound optional upstream RakNet pings

- Symptom: an Enchanted validation capture stopped after `upstream.dial_start` without recording an upstream transport connection. The capture ended before a ping-completion event, so it does not independently prove which dial sub-step was pending.
- Root cause: the upstream dialer left `PingTimeout` at zero for initial connections. The vendored gophertunnel dialer then reused the full 15-minute login and resource-pack context for its optional address-discovery ping. The vendored go-raknet documentation explicitly warns that `PingContext` can wait indefinitely without a deadline.
- Implementation: bound initial upstream address-discovery pings to five seconds, matching go-raknet's `Ping` convenience method. Transfer-hop pings retain their existing 750 ms bound. A failed ping remains recorded and gophertunnel continues to the original address as before.
- Evidence and provenance: capture [captures/session-20260924T161139Z](https://github.com/BedrockDebugProxy-Captures/session-20260924T161139Z/releases/tag/capture-archive) at revision `29795bbef1f266c9915d194e2d31410abfc5fbbb` ended after `upstream.dial_start` with no upstream transport connection. Source review of `third_party/gophertunnel/minecraft/dial.go` confirms a zero ping timeout inherits the caller deadline; `third_party/go-raknet/dial.go` documents the need for a bounded context and uses five seconds in `Ping`. This is a source-supported likely cause, not a claim that the capture isolated the ping stage.
- Status: focused proxy regression coverage and the full low-memory quality gate pass. Fresh exact-revision live validation remains required; the interrupted Enchanted session is incomplete.

### FIX-0145 - Use the Experience selector for CubeCraft release validation

- Symptom: the release matrix documented CubeCraft's endpoint as its upstream target, which made it appear that CubeCraft must bypass the Experience selector.
- Root cause: the docs confused a target's Featured server category and endpoint reference with the selector supported by the proxy. The selector resolves the Minecraft services catalog, which contains both Featured and Creator Experiences.
- Implementation: document `experience:CubeCraft` as the CubeCraft release selector, preserve `play.cubecraft.net:19132` as an endpoint reference, and state that `experience:` is shared by Featured and Creator Experiences. The CubeCraft risk gate and read-only validation requirements are unchanged.
- Evidence and validation: `internal/experience/target.go` resolves names from the Minecraft services discovery catalog through the shared `experience:` path and reports failures as "list featured and creator experiences"; `internal/experience/target_test.go` accepts `EXPERIENCE:CubeCraft`. Documentation review and `git diff --check`; no runtime code or live connection changed.
- Status: documentation-only correction. Existing validation reports and their tested revisions remain unchanged.

### FIX-0146 - Name the RakNet acceptance wait on cancellation

- Symptom: a canceled RakNet dial while waiting for `CONNECTION_REQUEST_ACCEPTED` reported only `context canceled`, hiding the handshake stage from the capture error.
- Root cause: the dialer's acceptance-wait cancellation branch returned the context error without operation context.
- Implementation: wrap that cancellation with `await connection request acceptance`. No packet, retry, timeout, or connection behavior changes.
- Evidence and validation: source review of `third_party/go-raknet/dial.go` confirms that this branch runs after sending `CONNECTION_REQUEST` and waiting for the acceptance signal. The Enchanted capture [captures/session-20260925T013400Z](https://github.com/BedrockDebugProxy-Captures/session-20260925T013400Z/releases/tag/capture-archive) contains a canceled hop 3 dial without transport-open evidence, but does not prove that this specific branch produced the observed error. `tools/format.ps1` and `tools/quality.ps1 -LowMemory` passed. Enchanted captures [captures/session-20260925T022556Z](https://github.com/BedrockDebugProxy-Captures/session-20260925T022556Z/releases/tag/capture-archive) and [captures/session-20260925T023226Z](https://github.com/BedrockDebugProxy-Captures/session-20260925T023226Z/releases/tag/capture-archive) completed six and four hops; the new cancellation diagnostic was not exercised.
- Status: implemented as a diagnostic improvement. The underlying cause of the earlier Enchanted stall remains unconfirmed.

## Ordering safety contract

The deferred-packet change is deliberately narrow. `expectedIDs` remains the only authority for what the current login state accepts. The queue scan selects the earliest packet that is currently expected, leaves packets for later states queued, and never mutates or reorders bytes on the wire. The regression tests cover canonical resource-pack order, featured early `ResourcePacksInfo`, a later-phase packet queued ahead of the current phase, and all six permutations of the three login packets.

This is a compatibility guard, not permission to assume a universal Bedrock sequence. Before changing an expected set, inspect the packet definitions, the relevant capture, and the server-specific validation flow. A new server observation must add a focused test and update this index or a linked research note.

## Open and pending items

- Generate a stamped binary from the current commit and run The Hive minimum baseline.
- Re-run the direct Enchanted transfer and resource-pack flow with the same binary.
- Complete the exact-revision matrix for Galaxite, Lifeboat, Mineville Zeqa, Enchanted, and CubeCraft after the The Hive baseline before release. CubeCraft requires maintainer risk acknowledgement first.
- If a server fails, record the capture and classify the failure before changing the generic state machine.
