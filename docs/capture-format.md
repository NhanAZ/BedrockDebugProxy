# Capture format

## Purpose

The canonical capture is designed for crash recovery, deterministic machine processing, human inspection, and later re-decoding. It is a directory in schema `bedrockdebugproxy.capture.v1`.

```text
capture-directory/
  manifest.json
  events.jsonl
  blobs/
    sha256/
      ab/
        abcdef...bin
```

## Manifest

`manifest.json` identifies the capture, generator, host platform, options, known limitations, completion status, and final counters. It is written when the capture opens and replaced from a synced temporary file when the recorder closes. Readers must tolerate an `open` manifest whose event stream contains newer data after an interrupted run.

A capture with `completeness.complete` set to `false` is still useful. The adjacent limitations and counters explain known missing, dropped, truncated, or failed data. An `open` status after the process is gone means the session did not close cleanly.

## Event stream

`events.jsonl` contains one JSON object per newline-terminated line. Sequence numbers start at one and define the canonical order. Each event also has UTC wall time, Unix nanoseconds, and elapsed monotonic nanoseconds from capture start. Nanosecond fields are decimal JSON strings so JavaScript and other limited-number readers do not lose precision.

Event kinds are namespaced strings such as `session.open`, `session.connection_metadata`, `session.game_data`, `session.spawned`, `transport.payload`, `packet.raw`, `packet.decoded`, `packet.decode_error`, `resource_pack.archive`, and `capture.limit`.

Connection and protocol context remains explicit through session ID, connection ID, hop, channel, logical direction, stage, source, and destination fields. A parent sequence can link a derived event to an earlier observation when the adapter can prove the relationship.

Decoded packet fields are stored in the optional `data` object. Each struct carries its Go type. Binary fields are summarized with their size, SHA-256 digest, and an optional short preview. The exact packet payload remains available through the raw packet event and blob reference.

Gophertunnel consumes some login and spawn packets before the normal forwarding loops can return them. The packet hook still retains their exact payloads as `packet.raw`. BedrockDebugProxy also records derived snapshots so later analysis does not have to decode every pre-play packet again.

- `session.connection_metadata` records the downstream or upstream role, authentication state, negotiated protocol, identity data, and decoded `ClientData` fields exposed by the active gophertunnel connection.
- `session.game_data` records a bounded decoded view of the `GameData` structure obtained from the upstream connection immediately before the same in-memory value is passed to downstream `StartGame`.
- `session.spawned` records the latency, client-cache state, and chunk radius exposed for both connections when the spawn handshake completes.

These snapshots use the same bounded structured encoder as decoded packet views. Byte slices become size, digest, and preview metadata. Login skin, cape, and geometry fields exposed as base64 strings remain full strings, which also supply the automatic skin files. The authoritative Login and StartGame payloads remain their `packet.raw` blobs. Connection snapshots contain player and device identifiers as well as those encoded assets and are sensitive.

A `resource_pack.archive` event references the exact archive retained by the adapter after download. Its data includes the pack UUID, version, manifest, byte length, computed checksum, delivery mechanism, feature flags, download URL when present, and content key when supplied by the server. The archive is stored before any decryption or extraction.

Skin, cape, entity, chunk, block-update, and inventory evidence follows the same canonical model. The `inspect --packet NAME` filter locates decoded packet views such as `PlayerSkin`, `LevelChunk`, `AddActor`, `UpdateBlock`, and `InventoryContent`. Exact encoded bytes remain in the corresponding `packet.raw` blob evidence. New CLI sessions also generate disposable [convenience folders](session-artifacts.md) under `artifacts/`. They supplement this evidence without changing capture v1, its counts, verifier, or portable export. `options.values.automatic_artifacts` identifies the separate view format. Its status and limits live in `artifacts/status.json`.

When opt-in decryption succeeds, `resource_pack.contents_manifest` and `resource_pack.decrypted_archive` events link to that raw archive through `parent_sequence`. A failed derivation produces `resource_pack.decrypt_error` with the same parent. Derived blobs never replace or mutate the original archive. Decrypted contents manifests contain per-file keys and are sensitive even though those keys are intentionally omitted from derived event metadata.

## Raw blobs

Raw data is addressed by lowercase SHA-256 digest. The canonical path uses the first digest byte as a directory and the full digest as the file name. Identical bytes are written once and may be referenced by many events.

Blob references include the digest, byte length, relative path, media type, and representation. Representation describes the observation boundary. For example, `packet_payload` and `bedrock_transport_payload` are different evidence and must not be conflated.

Blob paths are always relative and derived from the digest. Readers must reject absolute paths, parent traversal, non-canonical paths, hash mismatches, and size mismatches.

## Durability

The recorder writes each event under one ordering lock. A new raw blob is written to a temporary file, closed, and renamed to its content-addressed path before its referencing event is appended. Repeated in-memory payloads reuse an already published blob without another temporary file. Capture writes provide blocking backpressure and never silently drop evidence.

Normal CLI operation does not request a durable disk flush after every blob and event. The event stream is synced when the recorder closes, before the final manifest is published. This default avoids the severe forwarding latency observed when high-volume packet and chunk traffic waits for thousands of individual disk flushes. An operating-system or power failure may therefore lose the most recent unsynced filesystem writes even though an orderly `Ctrl+C` shutdown retains and verifies them.

`--sync-each-event` opts into syncing every new blob before rename and every event before forwarding continues. The manifest records this choice as `options.sync_each_event`. Use it only when maximum recovery from sudden power loss is more important than live session responsiveness.

Any future asynchronous or buffered mode must expose its queue bounds, memory use, backpressure, and loss policy in the manifest. It must never silently discard evidence.

## Compatibility

Readers must select behavior from the schema string instead of assuming the newest layout. Existing captures are immutable evidence. Schema migrations create a new capture or export and retain provenance to the source.

The capture verifier checks sequence continuity, capture identity, event counts, unique blob counts and bytes, every repeated blob reference, canonical paths, resolved path containment, regular-file type, byte lengths, and SHA-256 digests.

## Portable archive

The `export` command packages an already closed and verified capture as a ZIP-compatible `.bdpcap` file. It stores `manifest.json`, `events.jsonl`, and each referenced content-addressed blob exactly once. Entry order, timestamps, modes, compression method, and archive comment are fixed so exporting an unchanged capture twice produces identical bytes.

The exporter refuses an output inside the source capture, an existing output, paths that escape the capture directory, and source entries that resolve outside the capture through symbolic links. It reopens every completed entry to check ZIP integrity before returning the archive byte length and SHA-256 digest.

The archive is a portable copy, not a redacted report. It retains the same sensitive payloads, identifiers, endpoints, and resource-pack data as the source capture.
