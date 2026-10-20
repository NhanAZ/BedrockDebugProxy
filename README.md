# BedrockDebugProxy

BedrockDebugProxy is a high-fidelity Minecraft Bedrock traffic observation and research proxy. It records one client-to-server session as structured, machine-readable evidence that can be inspected again later.

> Capture once, analyze many times.

For plugin tests without a Minecraft client, [Golem CLI automation](docs/golem.md) provides persistent simulated players, interactive control, assertions, and bounded autonomy on a disposable PocketMine-MP or Axolotl-PM development server.

BedrockDebugProxy is independent software. It is not an official Minecraft product and is not approved by or associated with Mojang or Microsoft.

## Quick start

Run from the repository root. If the binary is missing, install the Go version declared in `go.mod`, keep the working tree clean, and build it with `.\tools\build.ps1 -Version dev`.

```powershell
# General form. Replace <UPSTREAM> before running.
# Use experience:<exact name shown in Minecraft> or HOST:PORT.
.\bin\bedrock-debug-proxy.exe `
    run `
    --listen 0.0.0.0:19132 `
    --upstream "<UPSTREAM>" `
    --auth device
```

For example, `--upstream "experience:The Hive"` selects The Hive. Replace it with the destination you intend to debug.

1. On first login, enter a non-empty local account label and complete Microsoft device authentication using the printed link and code. Later runs refresh the selected account's cached login until Microsoft requires authentication again.
2. Wait until the terminal prints `Listening on`. The proxy cannot accept Minecraft connections while authentication and Experience resolution are still in progress.
3. In Minecraft Bedrock, connect to port `19132` on the computer running the proxy. Use that computer's LAN address from another device. Do not enter `0.0.0.0` as the Minecraft server address.
4. Join the server and reproduce the behavior. The terminal prints connection transitions and three-second packet summaries while the capture retains the full evidence. Resource packs are buffered upstream before downstream delivery, and the terminal reports that wait.
5. Press `Ctrl+C` in the proxy terminal when finished.
6. Wait for the capture verification result. The exact output path is printed when the proxy starts and is normally `captures\session-<UTC timestamp>`.

The capture directory is the debug result. During play, `artifacts/` also fills with pack ZIPs, skin/cape PNGs, and categorized JSONL. No export command is needed to save them. These observations do not form a complete playable world. Keep them local because they may contain identifiers, chat, server data, resource packs, and other sensitive content. See the [folder guide](docs/session-artifacts.md).

`0.0.0.0` exposes the listener to reachable network interfaces. Use it only on a trusted network with an appropriate firewall. Use `127.0.0.1:19132` when only local software needs to connect.

By default, the downstream connection requires Xbox authentication. Add the proxy address in Minecraft's Servers tab. A LAN World entry uses a self-signed login. On a trusted LAN, `--allow-unauthenticated-client` accepts either path while keeping upstream authentication enabled. Any client that can reach this listener can use the proxy's authenticated upstream session.

When the destination sends a Bedrock `Transfer` packet, add `--follow-transfers` to keep the proxy in the path. The proxy rewrites the transfer destination to its local listener, waits for the client to reconnect, and then opens the next upstream hop in the same capture. Without this opt-in flag, the original transfer is captured and the process ends after the current hop. A concrete, reachable LAN address is required in `--listen` when following transfers; a wildcard such as `0.0.0.0:19132` cannot be sent to the client.

Add `--web-viewer` to inspect the capture in a local browser during the session. It follows complete events already committed to the capture. This example combines the viewer with trusted-LAN login and transfer following:

```powershell
.\bin\bedrock-debug-proxy.exe `
    run `
    --listen "<LAN_IP>:19132" `
    --upstream "<UPSTREAM>" `
    --auth device `
    --allow-unauthenticated-client `
    --follow-transfers `
    --web-viewer
```

Use `login NAME`, `accounts`, `use NAME`, `rename OLD_ACCOUNT NEW_ACCOUNT`, and `logout` to manage local accounts. The [authentication guide](docs/authentication.md) covers prompts, cache age, legacy caches, and storage. Authentication caches contain secrets and must not be shared or committed.

Use `help` for the command list and `help COMMAND` for usage:

```powershell
.\bin\bedrock-debug-proxy.exe help
.\bin\bedrock-debug-proxy.exe help run
.\bin\bedrock-debug-proxy.exe help login
```

## Choosing an upstream

The `experience:` form resolves the current destination through Minecraft services, so a Featured Experience or Creator Experience does not need a manually discovered IP and port.

```powershell
--upstream "experience:<exact name shown by Minecraft>"
--upstream "experience:<Experience UUID>"
```

Experience names are matched exactly without case sensitivity. An Experience UUID is also accepted. Device authentication is required because the resolver uses the same authorized Minecraft services session as the upstream connection. The resolver preserves the transport returned by the service, including RakNet, ID-based NetherNet signaling, and HTTPS NetherNet endpoints. HTTPS endpoints are checked through `/v1/join` before the proxy begins listening. The resolver does not fall back to a different transport when that check fails.

For a normal Bedrock server, use its address directly.

```powershell
--upstream example.org:19132
```

Use `--auth none` only when that direct server accepts unauthenticated connections. The downstream listener still requires authenticated Minecraft clients unless `--allow-unauthenticated-client` is deliberately enabled in a controlled test environment.

## What a session preserves

The canonical capture retains ordered evidence from both directions, including transport payloads exposed after the active transport, exact packet payloads, decoded packet views, unknown and malformed packet evidence, login and game-state snapshots, errors, protocol metadata, and connection lifecycle events.

Representative packet coverage includes skin and cape data, entities, chunks and subchunk requests, block updates, inventory content and transactions, movement, emotes, and other decoded packets that pass through the proxy. Binary fields are summarized in decoded views while their exact packet payload remains in content-addressed raw blobs.

Resource-pack archives, metadata, checksums, download information, and content keys supplied by the current upstream session are retained. The downstream offer preserves the upstream required/optional policy and other offer metadata when available. URL entries are replaced with archives served by the proxy. See [resource-pack behavior](docs/resource-packs.md) for fallback and delivery details. Decryption is opt-in because it creates additional sensitive plaintext artifacts.

```powershell
# Replace <UPSTREAM> with an Experience selector or direct HOST:PORT.
.\bin\bedrock-debug-proxy.exe run --upstream "<UPSTREAM>" --decrypt-resource-packs
```

Unsupported encryption variants retain the original archive and produce structured error evidence instead of a guessed fallback.

To capture resource packs without opening a local listener or connecting a Minecraft client, use packs-only mode.

```powershell
# Replace <UPSTREAM> with an Experience selector or direct HOST:PORT.
.\bin\bedrock-debug-proxy.exe run --upstream "<UPSTREAM>" --packs-only --web-viewer
```

This performs a real upstream login and resource-pack exchange, then closes before requesting spawn. The server may count, log, or reject the connection. Default client metadata can produce a different pack offer from a particular PC or mobile client. The capture retains login and pack evidence plus exact archives. `--decrypt-resource-packs` remains opt-in, and captures must stay local.

## Inspecting a capture

These commands analyze or package an existing capture:

```powershell
.\bin\bedrock-debug-proxy.exe verify C:\path\to\capture
.\bin\bedrock-debug-proxy.exe inspect --direction server_to_client --kind packet.decoded C:\path\to\capture
.\bin\bedrock-debug-proxy.exe view C:\path\to\capture
.\bin\bedrock-debug-proxy.exe view --live C:\path\to\open-capture
.\bin\bedrock-debug-proxy.exe analyze C:\path\to\capture > analysis.json
.\bin\bedrock-debug-proxy.exe explain C:\path\to\capture > explanation.md
.\bin\bedrock-debug-proxy.exe export C:\path\to\capture C:\path\to\session.bdpcap
```

`view` verifies a closed capture and opens a read-only local interface with filters, decoded details, and bounded blob previews. `view --live` follows an open capture. `inspect` locates original events and blobs, and `export` packages them without the convenience folders. Packet counts count events, not unique players, assets, or chunks. See [capture analysis and viewer behavior](docs/analysis-and-export.md).

```powershell
# Resource-pack archives and their blob paths.
.\bin\bedrock-debug-proxy.exe inspect --kind resource_pack.archive --limit 5 C:\path\to\capture

# Filter decoded packets by name, such as PlayerSkin, LevelChunk,
# AddActor, UpdateBlock, or InventoryContent.
.\bin\bedrock-debug-proxy.exe inspect --kind packet.decoded --packet PlayerSkin --limit 5 C:\path\to\capture
.\bin\bedrock-debug-proxy.exe inspect --kind bridge.forward_timing --direction server_to_client --limit 50 C:\path\to\capture
```

Each event's `blob.path` points to the exact retained bytes under the capture directory. Encrypted resource-pack archives are retained during a normal run, but plaintext derived archives exist only when `--decrypt-resource-packs` was selected before the session.

An exported `.bdpcap` is a portable copy, not a redacted copy. It can contain all sensitive evidence present in the source capture.

## Status and limitations

The project currently selects Bedrock 1.26.52 client metadata and protocol 2193, through the reviewed 1.26.50 gophertunnel packet definitions, the official 1.26.51 stable schema, and local featured-experience compatibility patches. Mojang has not published a stable 1.26.52 schema, so this selection does not establish exhaustive packet compatibility. See [the current maintenance review](docs/research/maintenance-2026-10-15.md) for evidence and validation limits. It supports one downstream client. With `--follow-transfers`, subsequent upstream hops remain in the same process and capture session. Automated tests include a full local RakNet session and focused tests for capture integrity, representative packet preservation, Experience response parsing, transport capability preservation, protocol metadata, and new packet bodies.

Automated success does not prove live Minecraft compatibility. Routine runtime pull requests use the scoped validation workflow in [`docs/validation.md`](docs/validation.md); a packs-only smoke check does not prove retail-client spawn or gameplay. Official releases require the six-server full-session validation matrix, driven by authenticated CLI clients or retail Minecraft clients. CLI sessions establish observed protocol flow without claiming retail rendering or gameplay. CubeCraft is included only when the maintainer accepts the connection risk for the exact diagnostic session. Until the current candidate completes that release process, treat it as development software rather than a production-ready proxy.

The Enchanted featured-experience transfer investigation, including the pre-fix failure signature and revision-matched live result, is recorded in the [research note](docs/research/enchanted-transfer-handshake.md).

Current known boundaries include the following.

- Transfer following is opt-in. Without `--follow-transfers`, Transfer packets are recorded and the current process ends after that hop. With it, the transfer is rewritten to the local listener and the next hop is captured in the same session. A session follows at most 64 hops to prevent transfer loops from creating unbounded output.
- The default capture path applies blocking backpressure and closes each content-addressed blob before forwarding continues. It avoids a durable disk flush per event for normal responsiveness. `--sync-each-event` provides stronger crash and power-loss durability at a substantial latency cost.
- Raw UDP datagrams and RakNet acknowledgement, fragmentation, retransmission, and loss details are outside the current capture boundary.
- Transport payloads are captured at the post-transport application boundary.
- Upstream resource-pack offer metadata is retained. URL-advertised entries are replaced with proxy-served archives for the downstream connection, and failed URL retrieval remains observable in the capture and operator warnings.
- Opt-in resource-pack decryption supports only the documented AES-256-CFB8 contents format. Original and supported plaintext ZIPs appear under `artifacts/packs/`, without unpacking their entries.
- Automatic artifact folders add background CPU and disk work. They can lag during play and have explicit view limits. Check `artifacts/status.json` after shutdown.

## Developer path

Run the shared quality gate before committing or declaring a code change complete.

```powershell
.\tools\quality.ps1
```

If local Go builds need a bounded memory profile, use `.\tools\quality.ps1 -LowMemory`. This temporarily limits Go to two runtime workers and one package at a time, then restores the caller's environment.

Use `tools\format.ps1` for deterministic Go formatting. CI runs the quality gate with the race detector on Linux and repeats tests and builds on Windows.

Use the document that matches your task.

- Contribution and pull request workflow is in [`CONTRIBUTING.md`](CONTRIBUTING.md).
- Live capture and report details are in [`docs/validation.md`](docs/validation.md).
- Maintainers preparing a release should follow [`docs/releasing.md`](docs/releasing.md).
- Protocol updates must follow [`docs/protocol-updates.md`](docs/protocol-updates.md).
- Capture and analysis contracts are in [`docs/capture-format.md`](docs/capture-format.md) and [`docs/analysis-and-export.md`](docs/analysis-and-export.md).
- Resource-pack behavior and legal boundaries are in [`docs/resource-packs.md`](docs/resource-packs.md) and [`docs/legal-and-responsible-use.md`](docs/legal-and-responsible-use.md).
- AI-agent and long-term maintenance principles are in [`AGENTS.md`](AGENTS.md).

## Repository layout

- `cmd/bedrock-debug-proxy` contains the CLI.
- `internal/experience` resolves Experience names and selects the returned upstream transport.
- `internal/bedrock` adapts transport and packet observation hooks.
- `internal/proxy` owns login, resource-pack negotiation, spawn, and forwarding.
- `internal/capture` owns the protocol-neutral capture schema, recorder, and verifier.
- `internal/analysis`, `internal/viewer`, and `internal/capturearchive` provide later analysis, local visual inspection, and portable export.
- `docs/research` records source revisions, licenses, evidence, and open questions.

## Security and content ownership

Never commit real captures, authentication state, resource-pack keys, decrypted packs, or third-party assets. Resource packs and other captured content remain owned and licensed by their respective rights holders. They do not become `GPL-3.0-or-later` because BedrockDebugProxy captured or decrypted them. Operators are responsible for their authority to inspect, retain, disclose, or redistribute session artifacts.

See [`docs/legal-and-responsible-use.md`](docs/legal-and-responsible-use.md) for the implementation trace, boundaries, and legal research notes.

## Server policy disclaimer

Using a proxy can violate a server's rules or terms even when the session is read-only and intended only for packet debugging. A server may reject the connection, disconnect or kick the player, or issue a temporary or permanent ban. Check the destination's current policy and obtain permission before connecting. The project does not authorize proxy use on any server. See the [server target and validation catalog](docs/server-targets.md) for the active matrix and known policy exclusions.

## AI-assisted development

NhanAZ maintains BedrockDebugProxy with substantial AI assistance and remains responsible for accepted changes and releases. Contributions require provenance review and validation under the documented workflow. Technical feedback should focus on reproducible behavior, evidence, and concrete improvements.

## License

BedrockDebugProxy is licensed under `GPL-3.0-or-later`. See [LICENSE](LICENSE). Third-party dependency licenses and source provenance are recorded in [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md) and `docs/research/`.

Copyright (C) 2026 NhanAZ.
