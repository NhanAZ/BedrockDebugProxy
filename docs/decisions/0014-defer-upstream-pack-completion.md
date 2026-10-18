# Defer upstream pack completion until downstream delivery

## Context

The terminating proxy downloads upstream archives before offering them to its downstream client. Previously its upstream dial also completed world startup before the downstream pack download began. A slow downstream delivery could therefore leave an already active upstream world waiting for client traffic that the bridge could not yet forward. The [capture failure review](../research/capture-failure-review-2026-10-18.md) and [the paired validation audit](../audits/AUDIT-0224-paired-stream-error-validation.md) record the observed OP SMP failure.

## Decision

Normal proxy dials return after validating and retaining ResourcePackStack, with PackResponseCompleted still unsent. The existing listener serves the retained archives, offer, and stack. After Listener.Accept confirms the downstream pack exchange, runHop resumes the upstream state machine through DoSpawnContext. It waits for upstream world metadata before publishing GameData and starting the downstream world.

This is an opt-in DeferResourcePackCompletion setting in the retained gophertunnel dialer. It has one completion signal and one resume signal per connection. The decoder owns the resumed state transition and flushes the ordinary completion response once. Closing the connection releases a paused decoder. A canceled DoSpawnContext does not release the barrier. Missing archives still reject the stack before the dial boundary is published. Exchange-only and deferred-completion modes cannot be combined.

The ordinary library dial and run --packs-only boundaries remain unchanged. Packs-only still flushes completion and closes before processing world startup. No packet definition, pack content, client cache policy, download pacing, or capture schema changes. The proxy does not manufacture gameplay acknowledgements or latency responses while the downstream client is loading.

## Evidence and alternatives

The official [protocol 2193 ResourcePackClientResponse definition](https://mojang.github.io/bedrock-protocol-docs/1.26.51/packets/resource-pack-client-response-packet/) distinguishes download completion from resource-pack stack completion. The existing [gophertunnel connection implementation at the pinned base revision](https://github.com/Sandertv/gophertunnel/blob/283a5a97dfe65da94bcc0b401807f6aefa9e72ee/minecraft/conn.go) owns separate dialled and listener login exchanges. These sources establish the handshake boundary, not a universal server timeout. Current local packet definitions are retained unchanged. The upstream source remains under MIT. No source was copied for this change.

Increasing a timeout cannot prevent the observed remote disconnect. Faster downstream chunk delivery would change an existing client compatibility safeguard. Responding to queued latency packets at the proxy would acknowledge receipt before the downstream client observed the surrounding traffic. Delaying only SetLocalPlayerAsInitialised was tested locally and did not prevent the disconnect, so that experimental change was removed. Delaying the resource-pack completion response keeps world startup behind the actual pack-loading boundary instead.

The synthetic mixed URL/chunk fixture verifies that downstream completion precedes upstream completion, while retaining offer order and metadata. The old configuration fails that ordering assertion. The paired BedrockPackArchive fixture verifies the unchanged archive and provenance contract with the additional ordered control packets. Exact runtime and live validation results belong in the commit audits.

## Limits

This change addresses startup overlapping a downstream pack exchange. It does not make a failed CDN response, unavailable backend, authentication outage, or rejected pack fallback succeed. A server can still enforce its own negotiation timeout. Headless authenticated sessions establish only the exercised network flow and do not establish retail rendering or gameplay compatibility.
