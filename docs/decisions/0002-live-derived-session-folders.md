# ADR 0002 - Live derived session folders

## Status

Accepted on 2026-08-31 after maintainer approval.

## Decision

Keep the canonical recorder unchanged and add one concrete `internal/artifacts` consumer of its live event file. The CLI starts it after opening a new capture and drains it after closing the recorder. A new package is justified by this separate file-format and lifecycle boundary, not by a generic plugin or export framework.

The event file acts as the existing disk-backed backlog. This avoids a second in-memory packet queue, mutable decoded-packet ownership across goroutines, extra callbacks inside the recorder lock, and output work on the forwarding path. Re-reading selected skin packet bodies uses the existing pinned decoder, not a parallel protocol implementation. Categorized JSONL copies preserve source order, timing, direction, and raw references without inventing relationships between raw and decoded events.

Resource-pack ZIPs are copied only after the corresponding canonical archive event exists. This consumer has no decryption or remote-fetch capability. Independent copies protect original blobs from edits to convenience files. PNG deduplication uses dimensions and raw-pixel hashes without a session-growing in-memory cache. Output names never use untrusted paths.

## Consequences

Convenience output can lag and consumes extra CPU and disk space. Explicit view limits and a separate versioned status file expose incomplete derivations without weakening capture fidelity. No canonical schema bump or historical capture migration is needed. Existing verification and portable export intentionally exclude these disposable views. See [session artifacts](../session-artifacts.md) for the exact layout, limits, and validation steps.
