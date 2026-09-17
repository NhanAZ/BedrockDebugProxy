# ADR 0009 - Follow committed capture events during a live session

## Status

Accepted on 2026-09-17. This decision supersedes only the closed-capture restriction in ADR 0008.

## Context

The closed capture viewer is useful after a session, but packet research often needs feedback while reproducing behavior in Minecraft. Sending packets directly to a browser or adding the browser as a recorder sink would couple user-interface speed and failure to network forwarding. Repeatedly verifying or rescanning an open capture would also waste work and report expected open-manifest counts as integrity failures.

The recorder already establishes a safe derived-view boundary. It stores each content-addressed blob before appending the newline-terminated event that references it. The manifest remains open until the asynchronous ordered writer drains and the event file closes.

## Decision

Allow the viewer to tail an open canonical capture through `view --live`. Add `run --web-viewer` as a convenience that starts the same live viewer for the new session. Both forms serve only on an automatically selected `127.0.0.1` port and retain the per-run access path and browser security headers from ADR 0008.

Index only complete newline-terminated JSON Lines records. Retain the last committed byte offset and read only appended bytes during refresh. Ignore an incomplete trailing record without advancing the offset. Continue reading full selected events and blob previews from the canonical files on demand.

Use one-second browser polling and server-side tail pagination. Do not add a packet bus, WebSocket, recorder callback, capture schema field, second database, or dependency from forwarding to the viewer. Browser closure, slow rendering, and HTTP failure must not affect packet forwarding or capture writing.

While the manifest is open, display observed event and unique blob counts from the incremental index and mark verification as pending. When the manifest closes, index any remaining records and run the normal verifier. Display the resulting verified or failed state and stop polling.

## Consequences

- Operators can inspect current packet evidence in a browser while reproducing a server flow.
- The browser can lag behind disk writes by the polling interval without delaying the proxy.
- A process using `view --live` can attach to a separately started proxy when it knows the capture directory.
- The integrated viewer ends with the `run` process. The canonical closed capture remains available for a later normal `view` invocation.
- The incremental row index still grows in proportion to event count. This decision does not introduce a bounded history or discard evidence.
- Live status is not an integrity verdict. Full verification begins only after the recorder closes the manifest.

## Evidence and validation

- Recorder ordering in `internal/capture/recorder.go` commits blob content before appending the referencing event line and drains the ordered asynchronous writer before closing the manifest.
- Synthetic tests cover events appended after viewer startup, tail pagination, blob preview during capture, transition to final verification, incomplete trailing records, loopback serving, and concurrent HTTP access under the Go race detector.
- The forwarding path and capture schema are unchanged.
