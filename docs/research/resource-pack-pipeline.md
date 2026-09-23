# Resource-pack download pipeline research

## Question

Why did a featured-experience resource-pack exchange take several minutes, and can the proxy reduce that wait without changing packet contents or delivery policy?

## Confirmed local behavior

The pinned gophertunnel base handled an upstream `ResourcePackDataInfo` by requesting one chunk and waiting for its response before sending the next request. The downstream listener also sends one chunk at a time. Those two paths are separate. The downstream pacing is a deliberate transport-safety boundary, while the upstream one-at-a-time request loop unnecessarily leaves the server's response window idle.

The Galaxite capture `captures/session-20260922T151050Z` at revision `a992d6c37f07ae04afb651ce2130712f9e4b209e` recorded 197 upstream chunk requests and 186 responses during the five-minute deadline. The exchange was still active when the deadline fired. The capture contains no decode or resource-pack integrity error that explains the delay.

## Primary source

The [gophertunnel resource-pack pipeline PR](https://github.com/Sandertv/gophertunnel/pull/509) describes the vanilla client behavior as a bounded request window of up to 100 chunks, replenished as responses arrive, with a reorder buffer for responses that arrive out of order. The PR also keeps a downstream send delay because removing it without send-side congestion control can increase RakNet loss. The PR was open on 2026-09-22 and is recorded as research evidence, not as an adopted dependency revision.

## Implementation decision

BedrockDebugProxy now applies only the upstream part of that behavior. Each pack uses a maximum of 100 outstanding chunk requests, honors the server-advertised chunk size, and replenishes the window after each accepted response. Chunk indices, request membership, bounds, and exact fragment lengths are validated. Response bytes are copied before entering the bounded reorder buffer. The pack is appended in the original pack sequence only after all fragments are contiguous and the archive parses successfully.

The downstream listener remains sequential and retains its existing pacing. No packet is rewritten, dropped, injected, replayed, or reordered on the wire. Captures continue to record every request and response, including malformed or unexpected fragments that terminate the connection.

## Confidence and limits

- Confirmed: the old upstream loop was serial, the local Galaxite capture was active but slow, and the new code bounds the request window and reorder memory.
- Inferred: the vanilla-sized request window should reduce idle time for servers that tolerate concurrent chunk requests.
- Unknown: the optimal window and server-side congestion behavior for each featured experience. The 100-request maximum is a bounded compatibility choice, not a claim that every server uses that exact value.

Focused tests cover the chunk-count bound, the initial request window, window replenishment, requested-index validation, and copying of decoder-owned response bytes. A fresh stamped The Hive baseline and a fresh Galaxite session at the exact revision remain required before release approval.

## Concurrency incident

On 2026-09-23, a live Hive session terminated with a Go `fatal error: concurrent map writes` in `Conn.handleResourcePackDataInfo`, at the transition from `downloadingPacks` to `awaitingPacks`. The stack and source establish the race: each completed pack has a downloader goroutine that removes its entry from `awaitingPacks` under `Conn.packMu`, while the next `ResourcePackDataInfo` handler inserted into that same map without taking the mutex. The download window made pack completion and the next pack's setup overlap.

The fix uses the same connection mutex for pack-map lookup and the map transition. `TestResourcePackDataInfoTransitionsConcurrentPacksSafely` stresses simultaneous pack setup, and the focused nested-module test passes under Go's race detector. This addresses shared in-memory queue state only; it does not alter packet ordering, chunk scheduling, or packet contents. A fresh stamped The Hive session is still required to validate the runtime fix.
