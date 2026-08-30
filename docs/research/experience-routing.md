# Minecraft Experience routing research

## Scope

This note records the evidence behind `experience:<name-or-uuid>` routing. The feature resolves a destination offered to the authenticated Minecraft user. It does not scan server infrastructure, guess addresses, or maintain a hardcoded server list.

Research was last checked on 2026-08-30.

## Sources

| Source | Revision | Evidence used |
| --- | --- | --- |
| [`bedrock-tool/bedrocktool` selector parsing](https://github.com/bedrock-tool/bedrocktool/blob/d7788b57acbdd3eb93ac1efdd4f1107b78aea9b0/utils/connectinfo/connect_info.go) | `d7788b57acbdd3eb93ac1efdd4f1107b78aea9b0` | `experience:` accepts a displayed name or UUID, resolves names through the current discovery result, and joins destinations that have an Experience ID. |
| [`bedrock-tool/bedrocktool` Gatherings client](https://github.com/bedrock-tool/bedrocktool/blob/d7788b57acbdd3eb93ac1efdd4f1107b78aea9b0/utils/franchise/gatherings/gatherings.go) | `d7788b57acbdd3eb93ac1efdd4f1107b78aea9b0` | The discovery blob supplies titles, static addresses, and Experience IDs. The public implementation documents the IPv4 and port join response used by the RakNet path. |
| [`bedrock-tool/bedrocktool` legacy signaling connection](https://github.com/bedrock-tool/bedrocktool/blob/d7788b57acbdd3eb93ac1efdd4f1107b78aea9b0/utils/franchise/signaling/conn.go) and [dialer](https://github.com/bedrock-tool/bedrocktool/blob/d7788b57acbdd3eb93ac1efdd4f1107b78aea9b0/utils/franchise/signaling/dial.go) | `d7788b57acbdd3eb93ac1efdd4f1107b78aea9b0` | Legacy WebSocket endpoint, authentication header, credentials messages, signal forwarding, and keepalive behavior. |
| [`lactyy2/pia` JSON-RPC signaling connection](https://github.com/lactyy2/pia/blob/91a85ff9f353c01eb1571aae0238c233a2f8a015/signaling/messaging/conn.go), [dialer](https://github.com/lactyy2/pia/blob/91a85ff9f353c01eb1571aae0238c233a2f8a015/signaling/messaging/dial.go), and [network protocol constants](https://github.com/lactyy2/pia/blob/91a85ff9f353c01eb1571aae0238c233a2f8a015/realms/network_protocol.go) | `91a85ff9f353c01eb1571aae0238c233a2f8a015` | Independent MIT-licensed evidence for the messaging WebSocket endpoint, JSON-RPC methods, nested envelopes, single or batched receive parameters, delivery notifications, and `NETHERNET_JSONRPC`. No source was copied. |
| [`gophertunnel` service discovery and tokens](https://github.com/Sandertv/gophertunnel/tree/v1.61.0/minecraft/service) | `v1.61.0`, commit `283a5a97dfe65da94bcc0b401807f6aefa9e72ee` | Current public APIs for service discovery, PlayFab-backed service tokens, and key-bound multiplayer tokens. |
| [`gophertunnel` NetherNet adapter](https://github.com/Sandertv/gophertunnel/blob/v1.61.0/minecraft/nethernet.go) and [`go-nethernet`](https://github.com/df-mc/go-nethernet/tree/v1.0.20) | `v1.61.0` and `v1.0.20` | NetherNet requires signaling and transport-specific batch framing, packet reads, and encryption behavior that wrappers must preserve. |

The Minecraft service endpoints used here are discovered dynamically from the current retail service discovery response. They are not treated as a stable public API. A successful historical implementation is evidence, not a guarantee that the service contract will not change.

## Chosen behavior

1. Literal `HOST:PORT` values retain the existing RakNet path and do not invoke Experience resolution.
2. `experience:` requires device authentication. The same reusable Microsoft token source supports service discovery and the eventual upstream login.
3. A non-UUID selector is matched against the current discovery titles using exact case-insensitive equality. Partial matching is deliberately rejected to avoid silently choosing the wrong destination.
4. A result with only a static address uses RakNet directly. A result with an Experience ID calls the join endpoint for the current authorized session.
5. The join result selects RakNet, legacy NetherNet, or JSON-RPC NetherNet. The proxy does not replace a returned NetherNet route with a guessed IP address.
6. The capture manifest retains the original selector, resolved name, Experience ID, selected transport, and resolved destination. Validation reports continue to omit addresses and opaque network identifiers.
7. The observation wrapper preserves NetherNet's packet-reader, batch-header, encryption-disable, context, and latency capabilities. Losing those optional methods would change framing or encryption rather than merely reducing diagnostics.

No long-lived Microsoft, Xbox, PlayFab, Minecraft service, or signaling token is written by this feature. The current device token source remains in memory for the process lifetime. Captures still contain sensitive connection metadata and must be handled accordingly.

## Provenance and license

The HTTP response model, JSON-RPC signaling path, cancellation behavior, and early-signal buffering are project-owner code implemented for BedrockDebugProxy and cross-checked against the cited public sources. The legacy signaling portions of `internal/experience/signaling.go` are substantially adapted from the cited original bedrocktool repository and are recorded in `THIRD_PARTY_NOTICES.md`. The original bedrocktool repository is distributed under GNU GPL version 3. The direct transport and authentication libraries retain their ISC or MIT licenses as listed in `THIRD_PARTY_NOTICES.md`.

## Verification boundary

Unit tests cover selector recognition, exact name matching, discovery response parsing, RakNet and both NetherNet join results, invalid destinations, JSON-RPC message shapes, and preservation of transport capabilities. The project-wide quality gate covers formatting, tests, builds, static analysis, and known reachable vulnerabilities.

Live behavior remains unverified until a stamped build completes the following checks.

- `experience:The Hive` resolves through device authentication, connects, spawns, forwards both traffic directions, and closes cleanly.
- At least one Creator Experience whose join result uses NetherNet exercises the matching legacy or JSON-RPC signaling path.
- No service response, token, network ID, or signaling payload is copied into a sanitized validation report.
- A service change or unsupported response fails with an explicit error rather than falling back to a guessed destination.

Do not describe this feature as guaranteed support for every Experience. The current list is account, region, version, and service dependent. Do not claim the service endpoints are official public APIs. Do not claim automated tests prove live Featured or Creator Experience compatibility.
