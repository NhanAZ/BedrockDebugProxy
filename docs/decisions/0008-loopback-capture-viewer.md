# ADR 0008 - Keep capture viewing separate from packet forwarding

## Status

Accepted on 2026-09-17. The closed-capture restriction was superseded by ADR 0009. The separation and loopback security decisions remain active.

## Context

Terminal summaries and JSON Lines inspection preserve evidence but are slow for a person exploring a large packet stream. EndstoneMC Spyglass demonstrates useful interaction patterns through its ordered packet list, filters, decoded detail tree, hex pane, keyboard navigation, and visible counters. Its Windows client overlay and injection boundary do not match BedrockDebugProxy's terminating-proxy architecture or capture-first mission.

The viewer can expose addresses, identifiers, chat, content keys, decoded payloads, and resource-pack bytes. It must not turn a local research capture into an accidental network service. It must also remain a derived view rather than a second capture format or a component on the forwarding path.

## Decision

Add `bedrock-debug-proxy view CAPTURE_DIRECTORY` as a read-only consumer of closed canonical captures. It verifies all events and referenced blobs before serving, builds a lightweight in-memory row index, and reads full selected events from `events.jsonl` on demand. The browser receives paginated row summaries, complete selected event JSON, and at most 64 KiB of the selected blob for a hex preview.

Serve embedded HTML, CSS, and JavaScript only on an automatically selected `127.0.0.1` port. Protect every route with an unpredictable per-run URL path, send restrictive browser security headers, and provide no LAN binding option. The command opens the default browser unless `--no-open` is selected and runs until interrupted.

Keep the viewer out of the network proxy, recorder, and artifact writer. Do not tail an open capture, mutate packets, rewrite capture files, add a second event database, contact a CDN, or copy Spyglass source. Display pagination, blob-preview length, and decoded-tree node limits are explicitly non-canonical UI limits.

## Consequences

- A human can navigate packet history without losing the structured and raw evidence used by agents and command-line tools.
- Viewer failure or browser closure cannot interrupt packet forwarding because the viewer runs only against a closed capture.
- Startup performs full verification and one indexing pass. Its lightweight index still consumes memory proportional to the number of events.
- Full selected event JSON can be large. The browser stops building the decoded tree after 4,000 nodes while leaving the event and capture unchanged.
- The random URL reduces accidental cross-origin discovery but is not a substitute for protecting the local machine or capture directory.

## Evidence and validation

- EndstoneMC Spyglass revision [`8fb7e2cdd888b591382ad4d002e08390f3685d83`](https://github.com/EndstoneMC/spyglass/tree/8fb7e2cdd888b591382ad4d002e08390f3685d83), licensed MIT, was reviewed for interaction and storage behavior. No source was copied.
- Synthetic closed-capture tests cover route protection, session metadata, filtering, problem selection, on-demand event loading, and bounded blob previews.
- CLI parsing tests cover the new command's required capture argument.
