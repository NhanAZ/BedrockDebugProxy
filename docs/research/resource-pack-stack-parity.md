# ResourcePackStack parity research

## Question

Does the terminating bridge send the same resource-pack stack that the upstream server sent, or does the listener reconstruct a different stack?

## Confirmed observation

The upstream Hive capture `captures/session-20260920T135412Z` recorded `ResourcePackStack` packet ID 7 with a 1,449-byte payload. The downstream copy was 1,446 bytes. Decoding the payload prefix shows that the upstream stack advertised 24 texture-pack entries while the old listener-generated stack advertised 32. The additional entries came from gophertunnel's built-in exempted-pack list, not from the upstream offer.

This was a resource-pack protocol-boundary discrepancy. It did not change gameplay packet bytes, but it meant the client did not receive the same stack metadata and ordering as a direct client.

## Narrow fix

Dialled connections now retain a copy of the complete `ResourcePackStack` in `GameData`. The listener exposes a small `FetchResourcePackStack` callback for bridges that have an exact upstream stack available. BedrockDebugProxy uses that callback to send the retained stack downstream. The normal listener-generated stack remains the fallback for callers that do not provide one.

The retained value is cloned at both connection boundaries. The fix preserves texture-pack order, base game version, experiments, and editor and toggle flags without changing resource-pack archive bytes or adding a second transfer path.

## Evidence and provenance

- Affected capture: local Hive capture `captures/session-20260920T135412Z`, sequences 979 and 1,183, packet ID 7. The capture remains local sensitive evidence and is not committed.
- Packet definition: vendored gophertunnel `minecraft/protocol/packet/resource_pack_stack.go`, which serializes the required flag, stack entries, base game version, experiments, and remaining flags in that order.
- Upstream reference: [gophertunnel ResourcePackStack](https://github.com/Sandertv/gophertunnel/blob/7a556a07335b663744b50d38062636ad8283f314/minecraft/protocol/packet/resource_pack_stack.go), reviewed 2026-09-21.

## Validation

- The nested gophertunnel unit test verifies complete-stack retention and slice ownership.
- The proxy loopback integration test configures a non-default stack upstream and verifies the same decoded values downstream.
- No public server was contacted for the fix. The capture evidence is a protocol observation, not evidence about any server-side enforcement decision.

## Limitations

`ResourcePacksInfo` remains subject to the explicit downstream delivery policy because the public listener API does not expose every upstream offer flag. URL offers still require archive-backed downstream delivery so the client receives bytes through the proxy connection. These are separate resource-pack boundaries from `ResourcePackStack` preservation.
