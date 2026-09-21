# ADR 0010 - Define vanilla compatibility as protocol parity

## Status

Accepted on 2026-09-21.

## Context

BedrockDebugProxy is a terminating observation proxy. It has two transport sessions, records the traffic at the supported protocol boundary, and decodes owned copies for analysis. The project has investigated local identity and metadata mismatches because those are observable interoperability defects. Those investigations must not be described as implementing, defeating, or evading a server anti-cheat system.

## Decision

For this project, "vanilla compatibility" means matching the official client where the proxy controls the protocol boundary and where an owner-controlled or otherwise authorized server can be used for validation. The comparison covers login metadata, protocol version, packet ordering, resource-pack exchange, transport behavior, raw same-protocol packet bytes, and the resulting capture evidence.

The proxy remains observational for gameplay traffic. It must not add packet mutation, dropping, injection, replay, cheat behavior, or detector-specific workarounds. A compatibility fix is justified only by a reproducible protocol, transport, metadata, or capture-fidelity discrepancy and must preserve the original evidence.

Server-side policy, blocklists, and anti-abuse decisions are external outcomes. A server may accept or reject a proxy connection independently of whether the proxy intentionally changes gameplay traffic. Public-server testing requires the server's permission and is not evidence needed to define the proxy's vanilla protocol behavior.

## Consequences

- Local fingerprinting and comparison tools may identify parity defects, but their findings are treated as interoperability evidence rather than a recipe for changing server classification.
- Authenticated two-session differences remain an architectural boundary when the upstream authentication library creates a separately signed login chain. They are documented as compatibility limitations until an authorized server-side test or protocol-supported design provides stronger evidence.
- Runtime changes continue to require focused tests, capture evidence, and the validation gates in `AGENTS.md` and `docs/validation.md`.

## Evidence and provenance

- [`docs/decisions/0001-capture-first-terminating-proxy.md`](0001-capture-first-terminating-proxy.md) defines the default observational bridge.
- [`docs/decisions/0004-same-protocol-raw-forwarding.md`](0004-same-protocol-raw-forwarding.md) records raw forwarding without gameplay packet mutation.
- [`docs/research/local-identity-parity.md`](../research/local-identity-parity.md) records the local identity and ClientData parity fixes and their validation limits.
- [`docs/research/hive-unfair-advantage-2026-09.md`](../research/hive-unfair-advantage-2026-09.md) records the unresolved public-server incident without inferring its cause.
