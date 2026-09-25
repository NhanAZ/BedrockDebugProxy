# NetworkSettings parity research

## Question

Does the terminating bridge expose the compression and client-throttle settings negotiated on each side of the session, and are those settings necessarily identical?

## Confirmed observation

The Bedrock `NetworkSettings` packet carries the compression threshold, compression algorithm, and client-throttle fields that control later packet batching and compression. The vendored gophertunnel listener creates this packet from its `ListenConfig`, while a dialled connection receives the packet from the upstream server during login.

The owner-controlled capture [captures/session-20260921T003441Z](https://github.com/BedrockDebugProxy-Captures/session-20260921T003441Z/releases/tag/capture-archive) recorded different settings on the two sides before the startup parity fixes. The upstream payload began with `01-00-00-00`, which decodes to a compression threshold of `1` and algorithm `0`. The downstream payload began with `00-01-00-00`, which decodes to a threshold of `256` and algorithm `0`. The remaining fields in both payloads were zero. This is a confirmed wire difference, not a decoder inference.

The local upstream value is also confirmed by the PocketMine source used for the controlled fixture. Its
`src/network/mcpe/handler/SessionStartPacketHandler.php` sends `NetworkSettingsPacket::COMPRESS_EVERYTHING`, which
encodes the threshold `1`. The proxy's vendored gophertunnel listener defaults `ListenConfig.CompressionThreshold` to
`256` in `minecraft/listener.go`. This explains the local pair without generalizing the value to every Bedrock server.

## Root-cause boundary

The listener sends its `NetworkSettings` response as soon as it handles the downstream `RequestNetworkSettings` packet. BedrockDebugProxy currently opens the upstream dial later, after downstream login metadata is available and resource-pack setup begins. Therefore the upstream settings are not available at the point where the listener chooses its response. Changing the downstream threshold dynamically would require a separate connection-start design decision and must not be guessed from one server capture.

## Narrow implementation

Each gophertunnel connection now retains the decoded `NetworkSettings` exchanged during login. BedrockDebugProxy adds bounded structured copies for both sides to the `session.negotiated` event, including an explicit `network_settings_match` boolean. The existing raw `packet.raw` events remain authoritative for exact bytes, and this change does not alter compression, batching, forwarding, or packet contents.

## Evidence classification

- Confirmed: the packet fields and the local capture values above.
- Confirmed: the current capture now records both structured settings when both login paths complete negotiation.
- Inferred: a threshold difference can affect later compression and transport payload boundaries because the packet encoder applies the threshold to batches.
- Unknown: whether any public server treats a particular threshold or batching profile as a policy signal, and whether matching it would improve compatibility.
- Unknown: a safe generic way to select the upstream threshold before the downstream client receives `NetworkSettings`; the current handshake requires that response before the client sends its `Login` metadata.

## Validation

- Nested gophertunnel tests cover retention of all `NetworkSettings` fields.
- The root loopback integration test checks that both sides expose equal settings in its deterministic fixture.
- The real Android loopback in [captures/session-20260922T044324Z](https://github.com/BedrockDebugProxy-Captures/session-20260922T044324Z/releases/tag/capture-archive) at revision `9994e4e11f595f7980fc6795f5cc8b05271d4de7` observed downstream threshold `256` and upstream threshold `1`, with protocol 2193 and game version 1.26.51 on both sides. The session reached spawn and closed cleanly. This confirms the mismatch on the owner-controlled local path without implying a public-server policy signal.
- No public server or external account was contacted for this change.

## Follow-up

Use a stamped local capture or an explicitly authorized owner-controlled server to compare settings across supported server categories. Do not change the listener handshake or attempt to normalize settings until a reproducible compatibility need and a safe per-connection design are established. The PocketMine `COMPRESS_EVERYTHING` result is evidence for that local server only, not a default to apply globally.
