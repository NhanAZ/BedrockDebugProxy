# Resource-pack download pipeline research

## Question

Why did a featured-experience resource-pack exchange take several minutes, and can the proxy reduce that wait without changing packet contents or delivery policy?

## Confirmed local behavior

The pinned gophertunnel base handled an upstream `ResourcePackDataInfo` by requesting one chunk and waiting for its response before sending the next request. The downstream listener also sends one chunk at a time. Those two paths are separate. The downstream pacing is a deliberate transport-safety boundary, while the upstream one-at-a-time request loop unnecessarily leaves the server's response window idle.

The Galaxite capture [captures/session-20260922T151050Z](https://github.com/BedrockDebugProxy-Captures/session-20260922T151050Z/releases/tag/capture-archive) at revision `a992d6c37f07ae04afb651ce2130712f9e4b209e` recorded 197 upstream chunk requests and 186 responses during the five-minute deadline. The exchange was still active when the deadline fired. The capture contains no decode or resource-pack integrity error that explains the delay.

## Primary source

The [gophertunnel resource-pack pipeline PR](https://github.com/Sandertv/gophertunnel/pull/509) describes the vanilla client behavior as a bounded request window of up to 100 chunks, replenished as responses arrive, with a reorder buffer for responses that arrive out of order. The PR also keeps a downstream send delay because removing it without send-side congestion control can increase RakNet loss. The PR was open on 2026-09-22 and is recorded as research evidence, not as an adopted dependency revision.

## Implementation decision

BedrockDebugProxy applies only the upstream part of that behavior. Each connection shares a maximum of 100 outstanding chunk requests across all offered packs, honors each server-advertised chunk size, and replenishes one connection-wide slot after each accepted response. Chunk indices, request membership, bounds, and exact fragment lengths are validated. Response bytes are copied before entering the bounded reorder buffer. A pack is appended in the original offer sequence only after all its fragments are contiguous and the archive parses successfully.

The downstream listener remains sequential and retains its existing pacing. No packet is rewritten, dropped, injected, replayed, or reordered on the wire. Captures continue to record every request and response, including malformed or unexpected fragments that terminate the connection.

## Galaxite regression and correction

The `0.3.0-rc.3` Galaxite capture [captures/session-20260923T044137Z](https://github.com/BedrockDebugProxy-Captures/session-20260923T044137Z/releases/tag/capture-archive) at revision `1ede568486768047acc5d52b721f595b50907fcc` recorded 19 `ResourcePackDataInfo` packets and 460 outgoing `ResourcePackChunkRequest` packets in about 10.5 seconds, but only 97 `ResourcePackChunkData` responses during the five-minute session. The dial ended as `context canceled`, not as the configured fifteen-minute timeout, and the capture reported no packet decode errors. A prior Galaxite capture at revision `29fbfd8e0664be20cb3929f5916f961e4e07b66c` recorded 460 chunk requests and 460 chunk responses, then reached `session.spawned` about 2 minutes 45 seconds after the upstream dial began.

Source inspection confirmed that every accepted `ResourcePackDataInfo` starts a separate downloader, while the 100-request window was enforced independently by each downloader. When a server offers multiple packs during the same exchange, those per-pack windows can multiply into a much larger connection-wide burst. The revised implementation places the bounded semaphore on the connection and releases a slot only after a requested response is accepted or a request write fails. The local capture supports the excessive aggregate request window as the regression. It does not show the server's internal handling, so whether rate limiting or another server-side behavior caused the low response count remains inferred.

The global limit follows the bounded client request window described by the upstream [gophertunnel resource-pack pipeline PR #509](https://github.com/Sandertv/gophertunnel/pull/509). The PR is an implementation reference, not proof of Galaxite's internal policy. A new stamped Galaxite run is required to verify the correction.

## Confidence and limits

- Confirmed: the old upstream loop was serial; the per-pack pipeline allowed requests to multiply across packs; and the connection-wide window is now capped at 100 with bounded per-pack reorder buffers.
- Inferred: the vanilla-sized request window should reduce idle time for servers that tolerate concurrent chunk requests.
- Unknown: the server-side reason for the low response count, whether the five-minute session was canceled by the operator or by the client, and the optimal server-side response pacing for every featured experience. The 100-request maximum is a bounded compatibility choice, not a claim that every server uses that exact value.

Focused tests cover the chunk-count bound, the initial request window, window replenishment, requested-index validation, and copying of decoder-owned response bytes. A fresh stamped The Hive baseline and a fresh Galaxite session at the exact revision remain required before release approval.

## Concurrency incident

On 2026-09-23, a live Hive session terminated with a Go `fatal error: concurrent map writes` in `Conn.handleResourcePackDataInfo`, at the transition from `downloadingPacks` to `awaitingPacks`. The stack and source establish the race: each completed pack has a downloader goroutine that removes its entry from `awaitingPacks` under `Conn.packMu`, while the next `ResourcePackDataInfo` handler inserted into that same map without taking the mutex. The download window made pack completion and the next pack's setup overlap.

The fix uses the same connection mutex for pack-map lookup and the map transition. `TestResourcePackDataInfoTransitionsConcurrentPacksSafely` stresses simultaneous pack setup, and the focused nested-module test passes under Go's race detector. This addresses shared in-memory queue state only; it does not alter packet ordering, chunk scheduling, or packet contents. A fresh stamped The Hive session is still required to validate the runtime fix.
