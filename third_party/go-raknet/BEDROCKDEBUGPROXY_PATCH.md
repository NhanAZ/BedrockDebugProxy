# BedrockDebugProxy patch

This directory is based on Sandertv go-raknet `v1.15.2-0.20260705184311-0d1fd09e2cf6` and retains its MIT license.

The project-specific change exposes the client RakNet GUID used during a dial and lets callers configure it. The proxy uses this narrow hook to preserve the vanilla client's RakNet identity across a Bedrock `Transfer`, together with the UDP source address. The upstream `MaxMTU` option is used for transfer handshakes so the route does not require a fragmented 1492-byte probe. No packet mutation or authentication bypass is added.

The project-specific `Conn.CloseImmediately` method sends the RakNet disconnect notification and closes the dialled UDP socket without waiting for the peer's close acknowledgement. Transfer-following uses it only after forwarding `Transfer`, before probing or dialing the next hop with the same source address. Ordinary connection closes retain the upstream acknowledgement behavior.

The dialer also tolerates a repeated `CONNECTION_REQUEST_ACCEPTED` during the initial reliable connection setup by re-sending `NEW_INCOMING_CONNECTION`. A Galaxite capture recorded the repeated acceptance after the first response was sent, and treating the retransmission as fatal prevented the upstream session from completing. The local regression test covers this recovery path. This is a transport reliability correction and does not alter Bedrock packet payloads.

When the dial context is canceled while waiting for `CONNECTION_REQUEST_ACCEPTED`, the returned error now names that wait stage. This improves capture diagnostics only. It does not change retries, packet contents, or connection behavior.

The patch was informed by owner-controlled transfer captures and project research reviewed on 2026-09-13. The source baseline and license remain those of go-raknet.
