# BedrockDebugProxy

BedrockDebugProxy is a high-fidelity Minecraft Bedrock traffic observation and research proxy.

The project is designed around one principle.

> Capture once, analyze many times.

Its first priority is to preserve enough structured evidence for AI agents to reconstruct a session, correlate events, inspect decoded packets, revisit raw payloads, and explain uncertainty. Human debugging and protocol reverse engineering are equally important consumers of the same capture.

## Status

The project is in its initial research and architecture phase. Networking and capture behavior are not ready for production use.

## Initial scope

The proxy will observe both client-to-server and server-to-client traffic. It will preserve raw data where the networking layer exposes it, decode packets when possible, record unknown and malformed input, and make encryption, compression, batching, framing, timing, and protocol metadata visible.

Resource-pack collection, reconstruction, integrity validation, and decryption are part of the planned capture pipeline when a connection exposes the required data and keys.

Packet mutation, dropping, injection, replay, cheat behavior, and exploit tooling are not initial goals.

## Repository layout

The detailed layout will be added after the architecture research is recorded. Project-wide working principles are in `AGENTS.md`.

## Security

Captures can contain credentials, server addresses, identifiers, chat, and proprietary content. Treat every capture as sensitive. Never commit real captures or authentication state.

## License

No public source-code license has been selected yet. This private repository remains all rights reserved until a license file is added deliberately.
