# BedrockDebugProxy

BedrockDebugProxy is a high-fidelity Minecraft Bedrock traffic observation and research proxy.

The project is designed around one principle.

> Capture once, analyze many times.

Its first priority is to preserve enough structured evidence for AI agents to reconstruct a session, correlate events, inspect decoded packets, revisit raw payloads, and explain uncertainty. Human debugging and protocol reverse engineering are equally important consumers of the same capture.

## Status

The repository has an initial single-client terminating proxy for the current protocol shipped by gophertunnel `v1.61.0`. It records transport payloads, raw packet payloads, decoded packet views, library failures, and lifecycle events in a durable capture directory.

Automated tests, `go vet`, module verification, and local builds pass. Real Minecraft client and cross-server validation is still pending. The proxy is not ready for production use.

## Run

Go 1.25 or newer is required. The module pins a tested Go toolchain for Windows development.

```powershell
go build ./cmd/bedrock-debug-proxy
.\bedrock-debug-proxy.exe run --upstream example.org:19132
```

Upstream Xbox device authentication is enabled by default. Use `--auth none` only for a server that accepts unauthenticated connections. The downstream listener requires authenticated clients by default. Use `--allow-unauthenticated-client` only in a controlled test environment.

Each run creates a new directory under `captures/`. The exact path is printed before the listener starts. After a clean shutdown, the CLI verifies event ordering, manifest counts, blob paths, sizes, and SHA-256 digests.

An existing capture can be verified separately.

```powershell
.\bedrock-debug-proxy.exe verify C:\path\to\capture
```

## Initial scope

The proxy will observe both client-to-server and server-to-client traffic. It will preserve raw data where the networking layer exposes it, decode packets when possible, record unknown and malformed input, and make encryption, compression, batching, framing, timing, and protocol metadata visible.

Resource-pack collection, reconstruction, integrity validation, and decryption are part of the planned capture pipeline when a connection exposes the required data and keys.

Packet mutation, dropping, injection, replay, cheat behavior, and exploit tooling are not initial goals.

## Repository layout

- `cmd/bedrock-debug-proxy` contains the CLI.
- `internal/bedrock` adapts gophertunnel and go-raknet observation hooks.
- `internal/capture` owns the protocol-neutral capture schema, recorder, and verifier.
- `internal/packetview` produces JSON-safe decoded packet views.
- `internal/proxy` owns login, resource-pack negotiation, spawn, and forwarding.
- `docs/decisions` records material architecture choices.
- `docs/research` records source revisions, licenses, evidence, and open questions.

Project-wide working principles are in `AGENTS.md`. The capture layout and observation boundaries are documented in `docs/capture-format.md` and `docs/decisions/0001-capture-first-terminating-proxy.md`.

## Current limitations

- One downstream client and one fixed upstream hop are handled per process.
- Transfer packets are recorded, but the proxy does not follow them.
- Raw UDP datagrams and RakNet acknowledgement, fragmentation, retransmission, and loss details are not captured.
- Transport payloads are captured at the post-RakNet application boundary. Their encryption and compression state is not yet classified per event.
- The upstream resource-pack-required flag is not mirrored to the downstream listener.
- Downloaded resource-pack archives, metadata, checksums, and content keys are stored. Decryption and extraction are not implemented yet.

## Security

Captures can contain credentials, server addresses, identifiers, chat, and proprietary content. Treat every capture as sensitive. Never commit real captures or authentication state.

## License

No public source-code license has been selected yet. This private repository remains all rights reserved until a license file is added deliberately.
