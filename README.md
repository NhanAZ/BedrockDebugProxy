# BedrockDebugProxy

BedrockDebugProxy is a high-fidelity Minecraft Bedrock traffic observation and research proxy. It records one client-to-server session as structured, machine-readable evidence that can be inspected again later.

> Capture once, analyze many times.

BedrockDebugProxy is independent software. It is not an official Minecraft product and is not approved by or associated with Mojang or Microsoft.

## Quick start

Already have `bin\bedrock-debug-proxy.exe` and only want a useful debug capture? Choose an upstream and run the general form below from the repository root.

```powershell
# General form. Replace <UPSTREAM> before running.
# Use experience:<exact name shown in Minecraft> or HOST:PORT.
.\bin\bedrock-debug-proxy.exe `
    run `
    --listen 0.0.0.0:19132 `
    --upstream "<UPSTREAM>" `
    --auth device
```

For example, the following command selects The Hive by its current Featured Experience name. The Hive is only an example target and is not required by the proxy.

```powershell
# Example only. Replace The Hive with the exact Experience name you want to debug.
.\bin\bedrock-debug-proxy.exe `
    run `
    --listen 0.0.0.0:19132 `
    --upstream "experience:The Hive" `
    --auth device
```

Then follow these steps.

1. Complete the Microsoft device login the first time. Later runs reuse the cached login until Microsoft requires authentication again.
2. Wait until the terminal prints `Listening on`. The proxy cannot accept Minecraft connections while authentication and Experience resolution are still in progress.
3. In Minecraft Bedrock, connect to port `19132` on the computer running the proxy. Use that computer's LAN address from another device. Do not enter `0.0.0.0` as the Minecraft server address.
4. Join the server and reproduce the behavior you want to debug. The terminal prints connection transitions and compact one-second packet summaries while the full evidence is written to the capture.
5. Press `Ctrl+C` in the proxy terminal when finished.
6. Wait for the capture verification result. The exact output path is printed when the proxy starts and is normally `captures\session-<UTC timestamp>`.

That capture directory is the debug result. Keep it local because it may contain identifiers, chat, server data, resource packs, and other sensitive content.

`0.0.0.0` exposes the listener to reachable network interfaces. Use it only on a trusted network with an appropriate firewall. Use `127.0.0.1:19132` when only local software needs to connect.

By default, the downstream Minecraft connection must use Xbox authentication. Add the proxy address in Minecraft's Servers tab for this secure path. A LAN World entry uses a self-signed client login instead. To accept that login on a trusted LAN, add the existing opt-in flag below.

```powershell
# Trusted LAN only. Replace <UPSTREAM> before running.
.\bin\bedrock-debug-proxy.exe `
    run `
    --listen 0.0.0.0:19132 `
    --upstream "<UPSTREAM>" `
    --auth device `
    --allow-unauthenticated-client
```

This flag does not disable upstream authentication. It does allow any client that can reach the listener to use the proxy's authenticated upstream session, so it is intentionally not the default.

On Windows, the Microsoft token cache is `%AppData%\BedrockDebugProxy\auth-token.json`. It contains authentication secrets, remains outside the repository, and must not be shared or committed. To change accounts, stop the proxy, delete this file, and start the proxy again.

If the binary does not exist yet, install the Go version declared in `go.mod`, keep the working tree clean, and build it once.

```powershell
.\tools\build.ps1 -Version dev
```

## Choosing an upstream

The `experience:` form resolves the current destination through Minecraft services, so a Featured Experience or Creator Experience does not need a manually discovered IP and port.

```powershell
--upstream "experience:<exact name shown by Minecraft>"
--upstream "experience:<Experience UUID>"
```

Experience names are matched exactly without case sensitivity. An Experience UUID is also accepted. Device authentication is required because the resolver uses the same authorized Minecraft services session as the upstream connection. The resolver preserves the transport returned by the service, including RakNet and supported NetherNet variants.

For a normal Bedrock server, use its address directly.

```powershell
--upstream example.org:19132
```

Use `--auth none` only when that direct server accepts unauthenticated connections. The downstream listener still requires authenticated Minecraft clients unless `--allow-unauthenticated-client` is deliberately enabled in a controlled test environment.

## What a session preserves

The canonical capture retains ordered evidence from both directions, including transport payloads exposed after the active transport, exact packet payloads, decoded packet views, unknown and malformed packet evidence, login and game-state snapshots, errors, protocol metadata, and connection lifecycle events.

Representative packet coverage includes skin and cape data, entities, chunks and subchunk requests, block updates, inventory content and transactions, movement, emotes, and other decoded packets that pass through the proxy. Binary fields are summarized in decoded views while their exact packet payload remains in content-addressed raw blobs.

Resource-pack archives, metadata, checksums, download information, and content keys supplied by the current upstream session are retained. Decryption is opt-in because it creates additional sensitive plaintext artifacts.

```powershell
# Replace <UPSTREAM> with an Experience selector or direct HOST:PORT.
.\bin\bedrock-debug-proxy.exe run --upstream "<UPSTREAM>" --decrypt-resource-packs
```

Unsupported encryption variants retain the original archive and produce structured error evidence instead of a guessed fallback.

## Inspecting a capture

The quick-start capture can be verified, filtered, summarized, explained, or packaged without reconnecting to the server.

```powershell
.\bin\bedrock-debug-proxy.exe verify C:\path\to\capture
.\bin\bedrock-debug-proxy.exe inspect --direction server_to_client --kind packet.decoded C:\path\to\capture
.\bin\bedrock-debug-proxy.exe analyze C:\path\to\capture > analysis.json
.\bin\bedrock-debug-proxy.exe explain C:\path\to\capture > explanation.md
.\bin\bedrock-debug-proxy.exe export C:\path\to\capture C:\path\to\session.bdpcap
```

An exported `.bdpcap` is a portable copy, not a redacted copy. It can contain all sensitive evidence present in the source capture.

## Status and limitations

The project currently uses gophertunnel `v1.61.0` and supports one downstream client with one upstream hop per process. Automated tests include a full local RakNet session and focused tests for capture integrity, representative packet preservation, Experience response parsing, and transport capability preservation.

Automated success does not prove live Minecraft compatibility. Runtime changes require a stamped real-client session on The Hive, and releases require the six-server validation matrix. Until the current candidate completes that process, treat it as development software rather than a production-ready proxy.

Current known boundaries include the following.

- Transfer packets are recorded but not followed automatically.
- The default capture path applies blocking backpressure and closes each content-addressed blob before forwarding continues. It avoids a durable disk flush per event for normal responsiveness. `--sync-each-event` provides stronger crash and power-loss durability at a substantial latency cost.
- Raw UDP datagrams and RakNet acknowledgement, fragmentation, retransmission, and loss details are outside the current capture boundary.
- Transport payloads are captured at the post-transport application boundary.
- The upstream resource-pack-required flag is not mirrored to the downstream listener.
- Opt-in resource-pack decryption supports only the documented AES-256-CFB8 contents format and does not extract pack files to the filesystem.

## Developer path

Run the shared quality gate before committing or declaring a code change complete.

```powershell
.\tools\quality.ps1
```

Use `tools\format.ps1` for deterministic Go formatting. CI runs the quality gate with the race detector on Linux and repeats tests and builds on Windows.

Use the document that matches your task.

- New contributors should follow [`CONTRIBUTING.md`](CONTRIBUTING.md).
- Pull requests use the repository template and the workflow in [`CONTRIBUTING.md`](CONTRIBUTING.md).
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
- `internal/analysis` and `internal/capturearchive` provide later analysis and portable export.
- `docs/research` records source revisions, licenses, evidence, and open questions.

## Security and content ownership

Never commit real captures, authentication state, resource-pack keys, decrypted packs, or third-party assets. Resource packs and other captured content remain owned and licensed by their respective rights holders. They do not become `GPL-3.0-or-later` because BedrockDebugProxy captured or decrypted them. Operators are responsible for their authority to inspect, retain, disclose, or redistribute session artifacts.

See [`docs/legal-and-responsible-use.md`](docs/legal-and-responsible-use.md) for the implementation trace, boundaries, and legal research notes.

## License

BedrockDebugProxy is licensed under `GPL-3.0-or-later`. See [LICENSE](LICENSE). Third-party dependency licenses and source provenance are recorded in [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md) and `docs/research/`.

Copyright (C) 2026 NhanAZ.
