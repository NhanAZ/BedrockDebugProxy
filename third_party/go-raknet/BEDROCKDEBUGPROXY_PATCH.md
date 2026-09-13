# BedrockDebugProxy patch

This directory is based on Sandertv go-raknet `v1.15.2-0.20260705184311-0d1fd09e2cf6` and retains its MIT license.

The project-specific change exposes the client RakNet GUID used during a dial and lets callers configure it. The proxy uses this narrow hook to preserve the vanilla client's RakNet identity across a Bedrock `Transfer`, together with the UDP source address. The upstream `MaxMTU` option is used for transfer handshakes so the route does not require a fragmented 1492-byte probe. No packet mutation or authentication bypass is added.

The patch was informed by the `NhanAZ-Tools/bedrocktool` fork, especially commits `cbeb266fbe9101f67d17840bb9b5e2749df4ace2`, `9b789ab61548b40daf618ffa56929626ceaff393`, and `afcaf783644c8a54a7de61123af3f28f4ca8c04`, reviewed on 2026-09-13. The source baseline and license remain those of go-raknet.
