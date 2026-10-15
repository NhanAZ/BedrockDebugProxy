# Golem development runtime

Accepted on 2026-10-15 for the maintainer's request to automate the full plugin test workflow through the CLI.

## Need and classification

This is a new developer capability, not a regression. The fix-history review found only synthetic network smoke tests and packs-only capture. Neither provides controllable inventories, physics, plugin events, permissions, or server assertions.

## Decision

Use a companion PocketMine-MP development plugin and a JSON Lines CLI. Reuse the pinned MIT Golem simulation runtime instead of implementing a second physics and inventory client. A local filesystem mailbox connects the CLI to the server main thread. No dashboard, public control socket, remote service, authentication bypass in the proxy, or protocol changes are introduced.

The operator explicitly installs both Golem and GolemAgent on a disposable development server. Access to the plugin data directory is administrative authority. Trusted PHP test files have the same authority as installed plugins. Never expose the mailbox through a web server or shared untrusted directory.

The CLI supports independent requests and sequential scenarios. Golems persist between invocations until reset or server shutdown. Responses report success, failure, request identity, server tick, and runtime identity. Requests and responses are retained. Timed-out requests must not execute later. Long-running actions are serialized and time bounded.

The complete Golem action families are exposed, including multi-player lifecycle, chat, commands, permissions, movement, block and entity interactions, inventories, forms, HUD observations, and state assertions. Explicit event subscriptions include custom plugin event classes. Trusted PHP scripts cover plugin-specific APIs through the real Player and server objects without adding speculative JSON methods.

## Evidence boundary

These are server-side simulation results. They do not establish Xbox authentication, PlayerPreLoginEvent, wire transport, resource-pack delivery, rendering, sound playback, or retail-client compatibility. Physics retains upstream limitations such as swimming and climbing. A successful command dispatch is not proof that the plugin achieved its intended effect. Assert the relevant state or event.

Simulation records have their own versioned format and are never labelled canonical network captures. Existing capture storage, resource-pack contracts, viewer fields, and BedrockPackArchive importing remain unchanged. No consumer modification is necessary for this isolated developer facility.

## Alternatives

A network client would exercise authentication and wire behavior, but a complete plugin-testing client also requires chunk state, physics, inventory prediction, and version-specific input. Adding all of that would duplicate a game client and exceed the demonstrated need. A raw packet injection API would not solve those semantics. The ordinary observation proxy remains available when testing the real network path.

See [the source review](../research/golem-development-runtime.md) for immutable evidence and licensing.
