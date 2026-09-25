# Audit AUDIT-0114

## Intent

Record primary protocol and owner-controlled server-source evidence for the
client-cache startup ordering boundary without changing runtime behavior.

## Scope

- Updated `docs/research/client-cache-parity.md`.
- Added FIX-0082 to `docs/fix-history.md`.
- No runtime code, capture schema, dependency, or generated artifact changed.

## Evidence and provenance

- The versioned Mojang Bedrock protocol reference for game version 1.26.50
  and protocol 2193 describes packet 129 as a client-reported login capability:
  <https://mojang.github.io/bedrock-protocol-docs/1.26.50/>.
- The owner-controlled PocketMine source at
  `src/network/mcpe/NetworkSession.php` and
  `src/network/mcpe/handler/ResourcePacksPacketHandler.php` sends the resource
  pack offer before it observes the client response.
- The closed capture [captures/session-20260921T122228Z](https://github.com/BedrockDebugProxy-Captures/session-20260921T122228Z/releases/tag/capture-archive) records the same
  downstream ordering and the downstream `01` versus upstream `00` values.

## Validation

- Automated validation: documentation-only. `tools/quality.ps1 -LowMemory` and
  `git diff --check` are required before commit.
- Local validation: source paths and the closed capture were inspected directly.
- Live validation: not performed. No server or account was contacted.

## Limitations and follow-up

This record does not claim that a device OS implies cache support or that the
cache difference affects server enforcement. An owner-controlled server that
deliberately exercises cache-enabled chunks is still required before a startup
redesign or a runtime default change.
