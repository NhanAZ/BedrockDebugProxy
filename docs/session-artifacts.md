# Automatic session folders

Every `run` creates the folders below as relevant data arrives. No extraction command is needed. Keep the normal command and change only the upstream target. Add `--decrypt-resource-packs` if you also want supported plaintext pack ZIPs.

```text
captures/session-<UTC timestamp>/
  manifest.json                  Canonical capture identity and options
  events.jsonl                   Complete canonical event stream
  blobs/sha256/                  Canonical raw and derived evidence
  artifacts/
    status.json                  View version, revision, limits, and completion
    packs/
      index.jsonl                Names, source sequences, checksums, and paths
      original/<name>-<hash>.zip  Exact archive copies, including encrypted packs
      decrypted/<uuid>-<hash>.zip Only when session decryption succeeded
    skins/
      index.jsonl                Player associations, source sequences, and files
      images/<size>-<hash>.png    Skin, cape, and animation textures
      data/<hash>.json            Valid geometry or resource-patch JSON
      data/<hash>.bin             Other observed skin data
    observations/
      skins.jsonl
      entities.jsonl
      world.jsonl
      blocks.jsonl
      inventory.jsonl
    errors.jsonl                 Present only when a derivation failed
```

Folders without relevant observations are not created. An absent cape does not produce a blank placeholder PNG. File paths in indexes and copied events are relative to the capture root, not the index's directory.

## While playing

A single background reader follows the already-written event stream. It does not queue network packets in memory or perform PNG and ZIP writes on the forwarding goroutines. Small JSONL writes are buffered and flushed approximately every second or when the reader catches up. Completed image and archive files are published by rename. The folders can lag behind network traffic, especially on a slow disk.

Identical image pixels and dimensions reuse one PNG. A pack copy reuses its sanitized name and archive hash. ZIPs are streamed without decompressing or extracting pack entries. Copies are independent files, not hard links, so editing a convenience file cannot change its source blob. This costs extra disk space for ZIPs, images, and selected JSONL events. It also adds background CPU and disk work, so real-client performance still requires validation.

When finished, press `Ctrl+C` and keep the terminal open until the shutdown stages complete. The console reports when forwarding stops, network connections close, the capture recorder closes, automatic artifacts finish, and capture verification completes. The reader finishes after the recorder closes and checks its last source sequence against the closed manifest. `status.json` is written at startup and completion. An `open` file after an interrupted process is not proof that the folders finished.

## Meaning and boundaries

- Pack ZIPs are copies of the original or already-derived archive blobs. Folder generation does not use resource-pack keys, run decryption, retrieve remote assets, or enable decryption by itself. Pack keys are not duplicated into the convenience index.
- Skin images come from inbound `PlayerSkin` and `PlayerList` packet bodies using the pinned gophertunnel decoder. PlayerList removal entries have no new image. Connection snapshots also supply the login skin and cape. Source sequences distinguish the observations without assuming player-list or entity packet order.
- PNG conversion retains straight RGBA channels, including RGB under transparent pixels. Geometry and other binary skin fields remain separate, exact files. Packet skin metadata retains binary summaries. Login index metadata is deliberately minimal and points back to the connection snapshot for other properties.
- Observation files contain unchanged canonical raw and decoded event objects for named packet families. Raw events retain blob references, and decoded binary fields remain summaries. World observations include chunk/cache traffic and game data. They are not a playable world, a complete entity database, or an inventory snapshot. Other packet families remain available in `events.jsonl`.
- `complete` means the view reader reached the closed capture without a derivation error. It does not prove that every skin rendered, that a full world was received, or that the session was authorized for redistribution.

## Errors, limits, and compatibility

The view format is `bedrockdebugproxy.artifacts.v1`. Canonical capture v1 is unchanged. Existing captures are not rewritten. `verify` and `.bdpcap` export continue to cover only canonical events and referenced blobs. Convenience folders are not included in `.bdpcap` and are not checked by the canonical verifier.

For a session declaring `automatic_artifacts`, the live validation report additionally requires a matching, error-free, completed view status for the capture ID, revision, and final source sequence. This is a completion gate, not a checksum verification of convenience files edited after generation.

The reader retains at most one source event line, limited to 64 MiB, rather than an expanding packet queue. A skin packet or base64 field is limited to 16 MiB, PNGs to 1,048,576 pixels, and the pinned skin codec's collection limits remain enabled. These are view limits, not capture or forwarding limits. Original bytes remain in the capture when a limit prevents a convenience view. Limits are recorded in `status.json`.

A malformed skin, unsupported view, or pack-copy failure creates an `errors.jsonl` entry with its source sequence. Later source events can still be processed. A source-stream or output-stream failure stops the view reader. The first problem also prints an operator warning and adds a capture limitation when the recorder is still open. Final state is `complete_with_errors` or `failed`, and `run` returns an error without discarding the canonical capture. If the disk prevents writing status itself, the startup `open` status may remain. Inspect terminal output as well.

Treat generated folders as read-only during a run. They have the same sensitive-data boundary as the local capture. The source-code GPL does not relicense captured textures, packs, skins, models, or other third-party content. Do not commit or redistribute them without the applicable permission.

## Human validation

On The Hive, accept the offered packs, spawn, walk through a populated area for two to three minutes, and interact with an inventory or entity. While connected, confirm that the relevant artifact files appear. Stop with `Ctrl+C`. The agent should check final status, source-sequence coverage, ZIP byte equality, PNG dimensions, and observation files. Report visible lag, incomplete chunks, frozen entities, or errors rather than accepting file existence as proof of a healthy session. A missing optional cape is not an error.

Use `--decrypt-resource-packs` to exercise the existing supported decryption flow and confirm that both original and plaintext ZIPs are present. The five-server release gate remains required.
