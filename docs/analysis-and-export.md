# Capture analysis and export

## Live output and saved data

One `run` invocation writes `events.jsonl` and content-addressed blobs throughout the session. It does not hold the complete capture in memory until shutdown. `Ctrl+C` closes and synchronizes the files, finalizes the manifest, and runs integrity verification. Analysis and export commands are optional consumers of that evidence, not required save steps.

Alongside canonical blobs, every new `run` automatically creates named pack ZIPs, skin/cape PNGs, and categorized JSONL under `artifacts/`. No extra extraction command is needed. The [folder guide](session-artifacts.md) explains source references, completion status, privacy, limits, and failure behavior. A complete playable world cannot be assumed from the chunks and updates that happened to reach one client.

Counts such as `PlayerSkin`, `PlayerList`, `AddActor`, `LevelChunk`, and `InventoryContent` count packet events. One packet can describe multiple objects, and an object can appear in many packets. These counts are not unique asset or object totals. `resource_pack.archive` counts recorded archive events. The event's blob reference locates the retained archive.

Live packet summaries are only operator navigation. They are emitted in three-second buckets so a busy session or a resource-pack exchange does not flood the terminal. The complete packet event stream remains in the capture. During upstream login, the console explicitly reports that the resource-pack exchange is intentionally buffered, then reports its elapsed duration before downstream delivery can begin. Pre-spawn raw packet summaries are tracked per hop, so a later transfer hop still shows its login and resource-pack exchange even after an earlier hop reached gameplay. Labels, channels, directions, totals, and selected packet names remain readable without color. Colored output follows PocketMine-MP's semantic logger roles and xterm 256-color terminal mapping. The timestamp uses aqua, informational messages use white, warnings use yellow, and errors use dark red. Packet directions and transfers retain their own navigation colors. Message text keeps the terminal's default foreground.

Upstream login and resource-pack exchange are bounded by a five-minute per-hop dial deadline. Featured experiences may advertise many packs and pace chunk responses across them, so a shorter deadline can abort an active exchange. A timeout is retained as an `upstream.dial_error` event and is not hidden behind an indefinite console wait.

| Output | Minecraft color | Hex |
| --- | --- | --- |
| Timestamp | Aqua | `#5FFFFF` (xterm 87) |
| INFO | White | `#FFFFFF` (xterm 231) |
| Client to server | Aqua | `#5FFFFF` (xterm 87) |
| Server to client | Green | `#5FFF5F` (xterm 83) |
| TRANSFER | Light purple | `#FF5FFF` (xterm 207) |
| WARN | Yellow | `#FFFF5F` (xterm 227) |
| ERROR | Dark red | `#AF0000` (xterm 124) |

The logger roles and terminal values are adapted from [PocketMine-MP MainLogger.php](https://github.com/pmmp/PocketMine-MP/blob/6a7cc02e9dff59b69241aa0bcffdb9903ce86beb/src/utils/MainLogger.php) and [Terminal.php](https://github.com/pmmp/PocketMine-MP/blob/6a7cc02e9dff59b69241aa0bcffdb9903ce86beb/src/utils/Terminal.php), inspected on 2026-09-14. This is a behavioral and palette reference, not a source-code import. Output redirection disables generated color escapes, as does a present `NO_COLOR` environment variable. Color does not carry information missing from the text labels.

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

Automatic `artifacts/` folders are disposable convenience views and are not included in `.bdpcap`. Their source events and blobs remain part of the canonical export.

Uncompressed entries preserve exact bytes and avoid wasting CPU on already-compressed or encrypted payloads. Fixed archive metadata makes the digest reproducible for the same capture. Consumers must still run `verify` after extraction rather than trusting the container alone. An export is a local copy, not a publication permission or a change to the ownership of captured content.
