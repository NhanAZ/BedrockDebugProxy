# BedrockDebugProxy patch

This directory is based on Sandertv go-raknet `v1.15.2-0.20260705184311-0d1fd09e2cf6` and retains its MIT license.

The project-specific change exposes the client RakNet GUID used during a dial and lets callers configure it. The proxy uses this narrow hook to preserve the vanilla client's RakNet identity across a Bedrock `Transfer`, together with the UDP source address. The upstream `MaxMTU` option is used for transfer handshakes so the route does not require a fragmented 1492-byte probe. No packet mutation or authentication bypass is added.

The patch was informed by owner-controlled transfer captures and project research reviewed on 2026-09-13. The source baseline and license remain those of go-raknet.
