# Capture analysis and export

## Analysis contract

`bedrock-debug-proxy analyze CAPTURE_DIRECTORY` verifies the capture and then emits one deterministic JSON object. The summary contains generator build identity, configured and negotiated protocol values, session milestones, capture status and completeness, manifest and observed counts, sorted event dimensions, packet combinations, content-addressed artifact totals, resource-pack metadata, grouped errors, and every verifier issue.

The session section reports whether the upstream connected, protocol negotiation completed, spawn completed, and the session closed. It also counts connection and GameData snapshots, raw packet and transport observations, decoded and unknown packets, decode errors, structured-view errors, and all error events. These fields support revision-matched live validation without requiring a report generator to copy sensitive event payloads.

Packet groups are keyed by direction, numeric ID, decoded name, and decode status. Artifact groups distinguish representations such as `packet_payload`, `bedrock_transport_payload`, and `minecraft_resource_pack_archive`. Unique byte totals count a content-addressed blob only once within each representation.

The resource-pack summary deliberately exposes only whether a content key was retained. It does not repeat the key. Raw events remain available through `inspect` when exact evidence is required.

The command exits with status 1 after writing the summary if verification found integrity issues. This allows an AI agent or CI workflow to retain the diagnostic output without treating a damaged capture as valid.

`tools/create-validation-report.ps1` consumes this JSON and emits a smaller sanitized live-session artifact. Validation remains outside the product CLI so the capture and analysis commands stay general. See [`validation.md`](validation.md).

## Human explanation

`bedrock-debug-proxy explain CAPTURE_DIRECTORY` derives Markdown from the same typed summary. Captured text is flattened and escaped from inline code delimiters before rendering. The explanation reports verifier results, manifest limitations, traffic dimensions, resource packs, grouped errors, and the boundary between local integrity and real client or server interoperability.

The explanation is evidence navigation, not a protocol verdict. A valid capture can still show implementation-specific ordering, omitted optional packets, or behavior that requires comparison with the named server software and protocol version.

## Event inspection

`inspect` streams original event objects as JSON Lines without loading the complete event stream into memory. Filters are exact matches and can be combined.

```powershell
bedrock-debug-proxy inspect --kind packet.decode_error --direction server_to_client --from-sequence 100 --limit 20 C:\path\to\capture
```

The command verifies referenced blobs before streaming and exits with status 1 if the verifier found an issue. Event output may include private payload metadata and resource-pack content keys. Treat it with the same care as the capture.

## Portable export

`export` accepts only a closed capture with no verifier issues. It creates a deterministic, uncompressed ZIP-compatible `.bdpcap` file outside the source directory without replacing any existing file. The CLI returns the absolute output path, archive SHA-256, byte length, entry count, and whether decrypted resource-pack artifacts are present as JSON. When they are present, it also writes a warning to standard error that the project license grants no ownership or redistribution rights for those assets.

Uncompressed entries preserve exact bytes and avoid wasting CPU on already-compressed or encrypted payloads. Fixed archive metadata makes the digest reproducible for the same capture. Consumers must still run `verify` after extraction rather than trusting the container alone. An export is a local copy, not a publication permission or a change to the ownership of captured content.
