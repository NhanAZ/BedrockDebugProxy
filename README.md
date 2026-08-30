# BedrockDebugProxy

BedrockDebugProxy is a high-fidelity Minecraft Bedrock traffic observation and research proxy.

BedrockDebugProxy is an independent project. It is not an official Minecraft product and is not approved by or associated with Mojang or Microsoft.

The project is designed around one principle.

> Capture once, analyze many times.

Its first priority is to preserve enough structured evidence for AI agents to reconstruct a session, correlate events, inspect decoded packets, revisit raw payloads, and explain uncertainty. Human debugging and protocol reverse engineering are equally important consumers of the same capture.

## Status

The repository has an initial single-client terminating proxy for the current protocol shipped by gophertunnel `v1.61.0`. It records transport payloads, raw packet payloads, decoded packet views, decoded login and game-state snapshots, library failures, and lifecycle events in a durable capture directory.

Automated tests, `go vet`, module verification, and local builds pass. Real Minecraft client and cross-server validation is still pending. The proxy is not ready for production use.

## Run

Go 1.25 or newer is required. The module pins a tested Go toolchain for Windows development.

```powershell
go build ./cmd/bedrock-debug-proxy
.\bedrock-debug-proxy.exe run --upstream example.org:19132
```

Upstream Xbox device authentication is enabled by default. Use `--auth none` only for a server that accepts unauthenticated connections. The downstream listener requires authenticated clients by default. Use `--allow-unauthenticated-client` only in a controlled test environment.

Each run creates a new directory under `captures/`. The exact path is printed before the listener starts. After a clean shutdown, the CLI verifies event ordering, manifest counts, blob paths, sizes, and SHA-256 digests.

Encrypted resource-pack decryption is opt-in because it creates additional sensitive plaintext artifacts. Use it only for packs you are authorized to inspect.

```powershell
.\bedrock-debug-proxy.exe run --upstream example.org:19132 --decrypt-resource-packs
```

The supported AES-256-CFB8 variant is evidence-backed and bounded. Unsupported variants retain their original archive and produce a structured warning instead of triggering a guessed compatibility fallback.

An existing capture can be verified separately.

```powershell
.\bedrock-debug-proxy.exe verify C:\path\to\capture
```

## Inspect and export

The CLI can stream selected canonical events as JSON Lines, produce a deterministic JSON summary, render a concise Markdown explanation, or package a closed verified capture into a portable `.bdpcap` archive.

```powershell
.\bedrock-debug-proxy.exe inspect --direction server_to_client --kind packet.decoded C:\path\to\capture
.\bedrock-debug-proxy.exe analyze C:\path\to\capture > analysis.json
.\bedrock-debug-proxy.exe explain C:\path\to\capture > explanation.md
.\bedrock-debug-proxy.exe export C:\path\to\capture C:\path\to\session.bdpcap
```

`inspect` preserves the original event objects and supports exact `--kind`, `--direction`, and `--channel` filters plus `--from-sequence` and `--limit`. `analyze` groups packet, error, artifact, direction, channel, and resource-pack evidence without embedding raw payloads or content keys. `explain` is derived from the same summary and distinguishes capture integrity from real server compatibility.

Export refuses open or invalid captures, never replaces an existing output, and writes entries in deterministic order with fixed ZIP metadata. A `.bdpcap` file still contains the complete sensitive capture, including raw blobs, retained resource-pack keys, decrypted contents manifests, and decrypted pack assets when present.

## Quality checks

Run the shared quality gate before committing or declaring a code change complete.

```powershell
.\tools\quality.ps1
```

The gate checks deterministic `gofmt` formatting, module tidiness and checksums, tests, builds, selected high-signal static analyzers, reachable known vulnerabilities, and Git whitespace errors. Analyzer and vulnerability-scanner versions are pinned in the script and do not modify `go.mod`.

Use `.\tools\format.ps1` to format every Go source file. CI runs the same quality gate with the race detector on Linux and repeats tests and builds on Windows. A local Windows environment with `CGO_ENABLED=0` should use the normal command because Go's race detector requires cgo.

Runtime behavior also requires a stamped real-client session and a structured validation report. The Hive is the minimum development and pull-request baseline. Releases require all six project baselines. See [`docs/validation.md`](docs/validation.md) for the exact workflow.

## Initial scope

The proxy will observe both client-to-server and server-to-client traffic. It will preserve raw data where the networking layer exposes it, decode packets when possible, record unknown and malformed input, and make encryption, compression, batching, framing, timing, and protocol metadata visible.

Resource-pack collection, reconstruction, integrity validation, and opt-in decryption are implemented when a connection exposes the required archive and keys.

Packet mutation, dropping, injection, replay, cheat behavior, exploit tooling, and a growing set of one-purpose download commands are not initial goals. The preferred workflow is one high-fidelity session that can support many later analyses.

## Repository layout

- `cmd/bedrock-debug-proxy` contains the CLI.
- `internal/bedrock` adapts gophertunnel and go-raknet observation hooks.
- `internal/analysis` creates deterministic machine and human summaries.
- `internal/capture` owns the protocol-neutral capture schema, recorder, and verifier.
- `internal/capturearchive` creates verified portable capture archives.
- `internal/packetview` produces JSON-safe decoded packet views.
- `internal/proxy` owns login, resource-pack negotiation, spawn, and forwarding.
- `internal/resourcepack` owns bounded derivation of supported encrypted resource-pack archives.
- `docs/decisions` records material architecture choices.
- `docs/research` records source revisions, licenses, evidence, and open questions.

Project-wide working principles are in `AGENTS.md`. The capture layout, analysis contract, validation workflow, protocol update workflow, and observation boundaries are documented in `docs/capture-format.md`, `docs/analysis-and-export.md`, `docs/validation.md`, `docs/protocol-updates.md`, and `docs/decisions/0001-capture-first-terminating-proxy.md`.

## Current limitations

- One downstream client and one fixed upstream hop are handled per process.
- Transfer packets are recorded, but the proxy does not follow them.
- Raw UDP datagrams and RakNet acknowledgement, fragmentation, retransmission, and loss details are not captured.
- Transport payloads are captured at the post-RakNet application boundary. Their encryption and compression state is not yet classified per event.
- The upstream resource-pack-required flag is not mirrored to the downstream listener.
- Downloaded resource-pack archives, metadata, checksums, and content keys are stored. Opt-in decryption supports only the documented AES-256-CFB8 format and does not extract files to the filesystem.

## Security

Captures can contain credentials, server addresses, identifiers, chat, and proprietary content. Treat every capture as sensitive. Never commit real captures or authentication state.

Resource packs and other captured content remain owned and licensed by their respective rights holders. They do not become `GPL-3.0-or-later` merely because BedrockDebugProxy captured or decrypted them. Operators are responsible for having authority to inspect a session and for any retention, disclosure, or redistribution of its artifacts. See [`docs/legal-and-responsible-use.md`](docs/legal-and-responsible-use.md) for the implementation trace, risk boundaries, and current source review.

## License

BedrockDebugProxy is licensed under the GNU General Public License version 3 or any later version. Its SPDX identifier is `GPL-3.0-or-later`. See [LICENSE](LICENSE) for the complete terms.

Copyright (C) 2026 NhanAZ.

The repository is private while the project is under development, with the intent to make it public when it is stable and ready. Private development does not change the license. "Information wants to be free" expresses the project's preference for debugging knowledge and distributed improvements to remain available to the community.

Third-party dependencies and provenance are recorded in [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md) and `docs/research/`.
