# Changelog

All notable changes to BedrockDebugProxy are documented in this single file. Each release keeps a concise, curated
summary here, while GitHub's compare view and the commit audit records provide the complete history and evidence.
Detailed rationale remains in [`docs/audits/`](docs/audits/).

## [Unreleased]

### Fixed

- Reject inconsistent capture manifest paths before verification or export, preventing an archive from omitting the event stream that was verified.
- Index the final live-viewer events before publishing verification completion. Display an empty blob as zero bytes instead of reporting that no blob exists.
- Preserve unregistered authentication cache files when renaming an account.
- Reject fractional-second authentication cache ages instead of silently truncating them, which could turn a subsecond limit into an unlimited cache.
- Publish the signaling shutdown cause before waking credential requests, so a closing connection cannot return missing credentials without an error.

### Changed

- Use Go 1.27.3 to reject compromised intermediate certificates (GO-2026-6809).
- Update klauspost/compress to 1.18.7 for the S2 dictionary security fix.

- Shortened the README and corrected stale resource-pack and validation guidance.

## [0.4.1] - 2026-10-19

Supported Bedrock version: 1.26.52 (protocol 2193).

[Published release](https://github.com/NhanAZ/BedrockDebugProxy/releases/tag/v0.4.1) | [All changes since v0.4.0](https://github.com/NhanAZ/BedrockDebugProxy/compare/v0.4.0...v0.4.1)

### Changed

- Official releases now accept full authenticated CLI sessions across the six-server matrix. Reports distinguish CLI evidence from retail-client observations and still require spawn, stable traffic, verified captures, the Enchanted pack-bearing transfer, and CubeCraft risk acknowledgement.

### Fixed

- Preserve the previous upstream GUID and UDP source address when a client reconnects quickly after Transfer. Wait for the preceding socket and bounded route probe before dialing the next hop.

- Mark optional transfer-dial status pings as route hints, matching the pre-reconnect probe. Retain ping failures without making a successful next-hop session fail validation.

- Finish downstream resource-pack delivery before completing upstream pack negotiation and starting its world. This prevents a slow initial pack download from leaving upstream gameplay waiting behind the loading client.

- Preserve resource-pack stream read errors instead of reporting only an invalid ZIP or accepting a complete-looking archive from a failed HTTP response. Temporary archive cleanup still runs on failure.

## [0.4.0] - 2026-10-15

Supported Bedrock version: 1.26.52 (protocol 2193).

[Published release](https://github.com/NhanAZ/BedrockDebugProxy/releases/tag/v0.4.0) | [All changes since v0.3.0](https://github.com/NhanAZ/BedrockDebugProxy/compare/v0.3.0...v0.4.0)

### Added

- Golems have seeded, adjustable personalities and optional bounded autonomous behaviour. Agents can control online players interactively, observe nonblocking movement, change direction, stop, and resume control across CLI processes. Direct actions pause autonomy. Personality affects exploration, pace, greetings and reactions to damage.

- Added an optional Golem CLI and development-server companion for multi-player plugin tests, including movement, block and entity interactions, inventories, forms, permissions, event observations, assertions, and trusted PHP test scripts. JSON Lines requests and retained simulation journals support AI agents without a dashboard or Minecraft client. This workflow requires the pinned Golem runtime and does not establish network or retail-client compatibility.

### Changed

- Experience resolution now connects NetherNet destinations that advertise an HTTPS hostname and port, including CubeCraft's current join response. It checks the server's NetherNet status endpoint, skips the unsupported RakNet status ping for HTTPS routes, and retains the existing ID-based signaling paths.
- Updated the vulnerability scanner to handle the Go 1.27 Linux source graph without crashing, while preserving the full source-analysis gate.
- Generated login skins now include valid arm-size and color defaults for strict Bedrock server parsers. Supplied client skin metadata remains preserved.
- Upstream login metadata now advertises the stable Bedrock 1.26.52 game hotfix while retaining protocol 2193 and the reviewed packet codec. The protocol watcher compares the latest reviewed Mojang schema independently of the client game version.
- Updated the preferred Go toolchain to 1.27.2 and the networking dependency security patches required by the current vulnerability scan.
- Resource-pack capture now accepts offers whose final download URL is empty, preserving offer order without counting downloaded archives again through the cache.
- Transfer following now retains a completed server transfer when the client closes before the server-to-client forwarding loop returns.
- Resource-pack decryption now preserves plaintext root manifests and icons even when contents.json declares keys for those metadata files.
- The Mojang protocol watcher now reads formatted release protocol numbers and preserves resolved metadata and UTC publication times in its issue tables.
- Empty resource-pack download requests now preserve the active exchange instead of failing login or replacing the download queue.
- Fixed-width NBT arrays now reject overflow and truncated input before allocating their advertised payload size.
- Concurrent RakNet handshakes now synchronize their shared request cache to prevent data races and process crashes.
- Updated go-nethernet to v1.0.24 with SDP candidate, fingerprint, and cancellable SCTP startup fixes. Its existing server-identity temporal-validation policy remains documented in [the JWT research note](docs/research/nethernet-server-identity-jwt.md).
- Resource-pack archives now prefer a root `manifest.json` over nested files with the same name, so a nested placeholder cannot override the pack identity.
- RakNet dial cancellations while awaiting `CONNECTION_REQUEST_ACCEPTED` now retain that handshake stage in the reported error.

## [0.3.0] - 2026-09-25

Supported Bedrock version: 1.26.51 (protocol 2193).

[Published release](https://github.com/NhanAZ/BedrockDebugProxy/releases/tag/v0.3.0) | [All changes since v0.2.1](https://github.com/NhanAZ/BedrockDebugProxy/compare/v0.2.1...v0.3.0)

### Added

- Added a read-only local capture viewer with paginated event filters, decoded JSON details, bounded blob hex previews, loopback-only serving, and per-run access URLs. Open captures can be followed with `view --live`, and `run --web-viewer` starts the live browser view automatically without putting browser work on the forwarding path.
- Added `run --packs-only` to capture resource-pack offers and archives from a real upstream login without opening a local listener or requesting world spawn. Its opt-in dial completion boundary waits for the final resource-pack response to flush, then closes before world startup. Packet evidence and archives remain in the canonical capture, and decryption stays opt-in.

### Changed

- Release validation now uses `experience:CubeCraft` for CubeCraft. The Experience selector covers both Featured and Creator Experiences, while the endpoint remains a reference.
- The release workflow now includes CubeCraft as a sixth, maintainer-risk-gated validation target. The normal read-only capture path is required, and a missing risk acknowledgement keeps the release incomplete. No private server discussion is recorded.
- Owner-controlled server validation follows the maintainer's direct internal request without requiring private authorization messages in reports or source documentation.
- Microsoft authentication now supports multiple named local accounts with one default account used automatically by `run`. The first interactive login requires a non-empty account username or local label, while later unlabeled logins may press Enter to reuse the selected default. `login`, `accounts`, `use`, `rename`, and `logout` cover account management without adding account flags to normal proxy commands. `rename` converts the legacy `default` cache to a named profile without reauthenticating. Interactive `login` asks once for an optional local cache age and Enter means no local limit. The `--max-age` form remains available for scripts and does not alter Microsoft token expiry.
- The CLI help is now grouped by proxy, authentication accounts, capture inspection, and information commands. `help COMMAND` provides a short purpose and usage for each command so the workflow is discoverable without memorizing a flat command list.
- Authenticated runs now print the selected local account label and store it in the capture manifest so the viewer can show the same account in its overview header. The label is local metadata, not token identity verification.
- Follow live now selects the newest matching event after each live refresh, scrolls only the event table, and keeps the current Details content while the new event loads so packet-heavy sessions do not collapse or jump the page.
- Large capture verification now reports scanned event and blob counts to the terminal and the live viewer while the closed capture is being checked, so a multi-minute verification is visibly active.
- Writes that fail because the remote peer already sent a Bedrock `Disconnect` are treated as connection termination during shutdown. The raw packet and session-close diagnostics remain available, while unrelated write failures still block a passing validation report.
- Optional upstream RakNet address-discovery pings now have a five-second deadline instead of inheriting the full upstream login and resource-pack timeout. Transfer-hop pings retain their existing 750 ms deadline.
- LAN listener examples now use an explicit `<LAN_IP>` placeholder instead of an operator-specific address.
- The quality gate now offers `-LowMemory` to bound local Go worker and package parallelism without changing the default CI profile.
- The owner-controlled ProxyFingerprint parity tool now compares lifecycle milestone timing in addition to identity, packet-order, and packet-count observations.
- Capture manifests now describe early transport-state availability accurately instead of claiming compression and encryption state is never classified per event.
- Capture manifests now describe the current resource-pack offer boundary accurately. Upstream offer metadata is preserved while URL entries are rebuilt from proxy-served archives.
- Transport payload events now include bounded per-connection compression and encryption state when that state is observable, while retaining the original raw application-batch blob.
- Transport payload metadata is now available during the login sequence as soon as the corresponding connection object exists, while the first pre-negotiation payload remains explicitly marked as not negotiated.
- `session.negotiated` capture events now distinguish the protocol implementation version from the game version and device OS reported in each connection's `ClientData`, including explicit mismatch booleans for local parity investigations.
- `session.negotiated` capture events now include bounded compression and client-throttle settings for both connections, with an explicit match flag for per-session parity analysis.
- Terminating bridges now wait for the downstream `ClientCacheStatus` before starting dynamic resource-pack preparation and pass that observed capability to the upstream connection.
- `ClientCacheBlobStatus` decoding now accepts large observed hash lists with a packet-specific bounded limit while keeping the generic slice safety bound unchanged.
- RakNet upstream dials now recover from a repeated `CONNECTION_REQUEST_ACCEPTED` by re-sending `NEW_INCOMING_CONNECTION` instead of aborting a potentially recoverable handshake.
- Featured-experience upstream dials now allow fifteen minutes for an active resource-pack exchange, based on a Galaxite capture that delivered 186 of 460 chunks during the previous five-minute bound.
- Upstream resource-pack downloads now share one bounded 100-chunk request window across the connection and reorder validated responses before archive assembly. This prevents the limit from multiplying across several packs, while downstream delivery remains sequential and paced.
- Resource-pack queue state is now synchronized across pack download workers, preventing concurrent-map crashes when one pack finishes as another begins.
- Updated the compiled Bedrock game-version metadata to 1.26.51 while retaining protocol 2193 and the reviewed 1.26.50 packet definitions. Mojang's 1.26.51 release reports no packet changes, so authenticated upstream login metadata now advertises 1.26.51 without changing forwarded packet bytes.
- Unauthenticated upstream dials now preserve the downstream client's identity instead of falling back to gophertunnel's default `Steve` identity. Xbox-authenticated dials continue to use the identity from the authentication chain.
- Unauthenticated upstream dials now remove forwarded XUID and PlayFab identity claims before creating a self-signed login, preventing authenticated identity data from being mixed into the offline token while retaining the downstream name and UUID.
- Authenticated upstream dials now preserve an omitted downstream `ClientData.PlayFabId` instead of synthesizing a second random identifier, and omit the empty JSON field to match the vanilla wire shape. Offline dials retain their existing fallback generation.
- Terminating bridges now retain and reuse the complete decoded upstream `StartGame` packet when sending world-start data downstream, instead of replacing fields outside the compact `GameData` view with proxy defaults.
- Startup `JigsawStructureData` and `VoxelShapes` packets are now retained from the upstream login and sent once with their original decoded contents instead of being duplicated around the generated spawn sequence.
- Terminating bridges now retain and reuse the complete upstream `ResourcePackStack`, preserving its pack order, base game version, experiments, and stack flags instead of appending local exempted packs.
- Terminating bridges now preserve decoded upstream `ResourcePacksInfo` offer metadata while rebuilding downstream pack entries from archive-backed resources that the proxy can serve.
- URL-prefetched resource packs are now merged with non-URL packs in the upstream offer order instead of replacing the complete list and dropping RakNet-delivered entries.
- URL-prefetched resource packs now bound the HTTP response body to the advertised compressed size before parsing, including responses without a trustworthy Content-Length.
- URL resource-pack prefetch now rejects malformed or trailing `ResourcePacksInfo` bytes before starting auxiliary downloads, while retaining the original raw packet evidence.
- URL-prefetched resource-pack requests now stop when their connection context is canceled, so closed sessions do not keep CDN fetches alive through the outer timeout.
- A successful RakNet resource-pack fallback now clears a stale URL-prefetch error for the same cache key, allowing later cache lookups to use the recovered archive.
- Early-login resource-pack and upstream work now uses the listener connection context before the observing wrapper is accepted, so a client disconnect during login cancels in-flight work promptly.
- The observing transport wrapper now preserves NetherNet's authenticated public-key capability so gophertunnel can retain its transport-to-login identity check.
- Transport status-ping responses are now retained as raw `transport.ping` evidence, including partial responses and operation errors.
- Transfer route probes now use the observing network boundary, retaining their status-ping responses alongside the route-probe result.
- Transport payload events now retain non-zero RakNet client GUID metadata when the connection exposes it.
- Transport connection lifecycle events now retain bounded RakNet GUID and public-key capability metadata without storing public-key material.
- Observed transport reads and writes now preserve both the underlying transport error and any capture-recorder error, including partial byte counts.
- Partial transport writes now retain only the bytes reported as transferred in the raw payload blob while preserving attempted and transferred byte counts in event metadata.
- Accepted capture events that cannot be persisted after a writer failure are now counted as dropped in the manifest instead of disappearing silently.
- Transfer-overlap packet evidence now keeps the physical flow's connection ID and hop instead of using the pending reconnect state.
- Connection wrapping now preserves a transport close error when the initial connection-open capture also fails.
- Transfer rewrite failures now preserve the captured error path when no raw packet read is retained, avoiding a nil dereference on typed or already-decoded forwarding paths.
- RequestChunkRadius listener callbacks now propagate error-event write failures to the session failure sink instead of continuing with an incomplete capture.
- Spawn, upstream-dial, and transfer-route lifecycle errors now preserve failures from writing their diagnostic events and stop incomplete captures through the existing failure sink.
- Upstream setup failures now retain transport or capture errors returned while closing the connection during cleanup.
- Transfer rewrite view failures now preserve the recorder error without dereferencing a missing retained raw packet.
- Downstream resource-pack offers now retain per-pack content identity, sub-pack, script, addon, and RTX metadata while still rebuilding entries from proxy-served archives.
- Terminating bridges now retain and reuse raw `DimensionData` and `ItemRegistry` startup packets when the downstream protocol matches, avoiding wire differences caused by typed re-serialization of NBT maps.
- Terminating bridges now forward each observed downstream `RequestChunkRadius` to the paired upstream connection while retaining the existing 16/16 initial fallback.
- Chunk-radius forwarding failures are now recorded as capture bridge errors instead of being silently ignored when the paired upstream connection is unavailable or closed.
- Transfer-following now closes the previous upstream RakNet socket before probing and dialing the next hop with its reused UDP source address, preventing a source-port bind conflict while preserving ordinary close behavior.
- Negotiated `NetworkSettings` metadata can now be read safely while the login packet reader updates it, preserving race-free connection snapshots.
- Malformed observed chunk-radius requests now record packet decode errors with typed packet context while retaining their raw evidence.
- Chunk-radius decode errors now retain sender and target sub-client header context in addition to the raw packet.
- Same-protocol raw-forward decode failures now appear as packet decode errors instead of generic bridge read errors.
- Same-protocol packets whose derived decode fails now remain raw-forwarded with an explicit unknown view instead of closing the bridge.
- Terminating bridges now retain and reuse raw `ResourcePackStack` startup packets when the downstream protocol matches, with a typed fallback when it does not.
- Terminating bridges now retain and reuse raw `ResourcePacksInfo` offer packets when the prepared downstream offer is equivalent, while keeping typed rebuilding for URL replacement or offer changes.
- Raw `ResourcePackStack` forwarding now uses the retained source protocol provenance on the listener path, ensuring the same-protocol raw branch is actually selected.
- Raw `ResourcePackStack` forwarding now validates the known decoded stack fields before reusing retained bytes, while preserving trailing fields that this protocol revision does not understand.
- Terminating bridges now also retain and reuse raw `JigsawStructureData`, `StartGame`, and `VoxelShapes` startup packets when the downstream protocol matches, preserving the observed world-start bytes while keeping typed fallback behavior.
- Terminating bridges no longer add an empty compatibility `CreativeContent` packet when the upstream server's own packet will be forwarded after spawn, avoiding the observed duplicate startup packet while retaining the upstream raw bytes.
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
- Validation reports now preserve capture start and end timestamps as invariant UTC ISO 8601 values instead of locale-formatted strings.
- Resource-pack decryption now accepts manifest entries for implicit directories while continuing to validate missing descendants and keyed-directory declarations.
- Best-effort transfer route pings now retain their errors as warning evidence and do not block a report when the normal next-hop connection succeeds. Other status-ping errors remain blocking.
- Validation report generation now treats a null integrity-issue list as empty, matching the analyzer's JSON representation for captures without verifier issues.

### Fixed

- The proxy listener now accepts client batches of up to 16,384 packets instead of closing a session at the vendored decoder's 812-packet default. Other listeners retain the default, and the proxy still enforces a finite batch bound.
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
