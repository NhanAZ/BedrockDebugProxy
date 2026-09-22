# Local identity parity research

## Question

Does BedrockDebugProxy preserve the downstream identity when it dials an operator-controlled unauthenticated Bedrock server?

## Confirmed observation

On 2026-09-21, an offline gophertunnel client named `LocalProbe` connected directly to the local Axolotl-PM 5.49.2+dev server and then through BedrockDebugProxy. The local server used `xbox-auth=off` only for this loopback test and was restored to `xbox-auth=on` afterward.

Before the fix, the proxied connection reached the local server as `Steve`, which is gophertunnel's default display name when no identity is supplied to an unauthenticated `Dialer`. The direct connection reached the server as `LocalProbe`.

The fix passes the identity supplied to `FetchResourcePacks` into the upstream `minecraft.Dialer`. After the fix, both direct and proxied local connections reached the server as `LocalProbe`. The upstream fake-server regression test also confirms that the downstream identity UUID and display name are preserved.

The binary used for the first post-fix loopback attempt was stale and predated this change. Rebuilding `bin/bedrock-debug-proxy.exe` from the current revision removed the apparent mismatch. The rebuilt comparison reported zero field differences, no packet-count differences, and a common packet prefix of eight for the direct and proxied sessions. Both sessions completed login, resource-pack exchange, spawn, and disconnect.

## Source path

- `internal/proxy/proxy.go` receives `login.IdentityData` in the gophertunnel `FetchResourcePacks` callback and now passes it to `connectUpstream`.
- `internal/proxy/proxy.go` sets `minecraft.Dialer.IdentityData` for the upstream dial.
- `third_party/gophertunnel/minecraft/dial.go` replaces `Dialer.IdentityData` only when an Xbox-authenticated token source supplies a verified identity. The unauthenticated path keeps the supplied identity.

## Interpretation

This was a local protocol and identity-parity defect, not evidence about any public server's anti-cheat. Passing the identity is safe for the local unauthenticated path and does not override authenticated Xbox identity data. The change does not mutate packets or attempt to evade server policy.

The local PocketMine console also emitted `Unexpected JSON property` warnings for `PlayFabId` in ClientData and `ipt`/`tid` in the self-signed token. PocketMine's local JSON models do not declare those fields in this revision, while gophertunnel includes them. The warnings were identical on direct and proxied connections, and both connections completed successfully. They are therefore parser-compatibility observations, not evidence that the proxy caused a ban or that the local server rejected the login.

## Follow-up identity finding

The 1.26.51 vanilla-client loopback run on 2026-09-21 exposed a second local parity issue. The proxied offline login contained a self-signed token `mid` derived from the forwarded `IdentityData.PlayFabID`, while the forwarded `ClientData.PlayFabId` remained the client value. ProxyFingerprint therefore reported `auth_client_playfab_mismatch`. The direct vanilla shape has no authenticated `mid` claim in this offline path. The unauthenticated dial helper cleared XUID and legacy title ID but not the two PlayFab identity fields. The focused fix clears all four authentication-bound fields before `EncodeOffline` and leaves the operator's display name and UUID intact.

## Post-fix vanilla-client loopback

After the fix was rebuilt into `bin/bedrock-debug-proxy.exe`, the real Windows Bedrock client signed in as `SampleAccount` through the proxy to the operator-controlled PocketMine server. The session ran for approximately 220.9 seconds, completed login, resource-pack negotiation, spawn, normal movement, and quit, and captured 4,350 packets. The proxy and upstream both negotiated protocol 2193 and game version 1.26.51 with Windows device metadata.

The closed ProxyFingerprint report for this run contained no `auth_client_playfab_mismatch` signal. Its only heuristic signals were `non_public_peer` and `non_full_authentication`, both expected for the loopback server configured with `xbox-auth=off` and `verify-xuid=false`. The upstream self-signed identity fields were empty while the forwarded ClientData remained structurally valid. This is evidence that the local offline identity-parity defect is fixed for a real vanilla client. It is not evidence about a public server's anti-cheat or acceptance policy.

## Authenticated upstream boundary

The authenticated path has a separate, source-backed boundary. Sandertv/gophertunnel at immutable commit [`7a556a07335b663744b50d38062636ad8283f314`](https://github.com/Sandertv/gophertunnel/blob/7a556a07335b663744b50d38062636ad8283f314/minecraft/dial.go) still calls `setAndroidData` before encoding an authenticated login because the XBL chain's title ID cannot be edited. That dialer creates a new signed upstream session from the token source. Preserving the downstream device metadata without also changing the signed chain would create an internal metadata contradiction, not a vanilla-equivalent login. This project therefore keeps the upstream behavior and treats authenticated two-session parity as an architectural limitation requiring an authorized server-side test, not a target for heuristic evasion.

### Authenticated ClientData PlayFab field

The 2026-09-22 local authenticated run at revision `936acb064ac7c5a216a60ba919aaddb6e1eaa80a` exposed a narrower metadata defect inside that boundary. The direct vanilla report omitted `ClientData.PlayFabId`, while the proxied upstream report contained a generated value whose HMAC differed from the authenticated token's `mid` claim. ProxyFingerprint recorded `auth_client_playfab_mismatch`; both sessions still completed login, spawn, and quit.

The source path is confirmed in the in-tree gophertunnel fork. `defaultClientData` generated a random `PlayFabId` before the authenticated branch called `setAndroidData`, even when the downstream client had supplied an empty field. The dialer now enables that fallback only for offline logins and preserves an empty authenticated field. This removes a proxy-created identifier without copying, forging, or re-signing the authenticated identity. The Android device metadata and the separate signed upstream session remain known authenticated-parity boundaries.

The mismatch removal is confirmed by source and deterministic unit coverage. A fresh real-client authenticated loopback report is still required to verify the emitted login shape after the fix. No public-server behavior is inferred from this local evidence.

### Post-fix authenticated vanilla loopback

On 2026-09-22, a real Windows Bedrock client signed in as `SampleAccount` through the stamped `970531e7d7132a4e0f889e1ab9ad735a8d4002c1` binary to the operator-controlled local PocketMine server. The session reached spawn, remained active for roughly one minute, and ended with `player_quit`.

The closed proxy capture `captures/session-20260922T041204Z` contains 12,750 events and 6,118 blobs with zero write errors, dropped events, truncation, or decode errors. Its manifest records the exact tested revision. The matching ProxyFingerprint report completed with 1,391 packets, `ClientData.PlayFabId` absent, and only the expected `non_public_peer` signal. The earlier `auth_client_playfab_mismatch` signal is absent after the change. The PocketMine log still reported an `Unexpected JSON property for ClientData: PlayFabId` warning during this run because the empty field was omitted from the selected report but was still serialized on the wire. The follow-up JSON-tag change omits the key itself when empty.

After that follow-up was built into revision `9994e4e11f595f7980fc6795f5cc8b05271d4de7`, the real Windows client repeated the local loopback on 2026-09-22. The client reached spawn, sent chat and gameplay traffic, and quit normally. Capture `captures/session-20260922T043146Z` closed with 8,250 events and 3,837 blobs, with zero write errors, dropped events, truncation, or decode errors. The matching ProxyFingerprint report `20260922T043303.076778Z_47906eadaa099058_unlabeled.json` completed after `player_quit`, recorded `ClientData.PlayFabId` as absent, and emitted only `non_public_peer`. The PocketMine log contains no `Unexpected JSON property for ClientData: PlayFabId` warning for this session and records `XUID match`. This confirms the empty-key wire-shape correction on the owner-controlled local authenticated loopback.

This confirms the synthetic authenticated PlayFab field is removed on the real-client local path. The upstream report still advertises Android metadata because the authenticated gophertunnel session remains a separate signed connection. This local result does not establish public-server compatibility or enforcement behavior.

### Local XUID rejection with an unauthenticated upstream

The closed local capture `captures/session-20260921T121717Z` provides a second confirmation of this boundary. The downstream connection was an authenticated Bedrock client, while the proxy was started with `auth_mode=none` and the local upstream required its normal identity checks. The upstream dial reached the login exchange but failed with `XUID does not match (possible impersonation attempt)` after 1 minute 29.872 seconds. The capture has no `session.negotiated` event because the upstream never completed login.

This result is confirmed evidence that an authenticated downstream identity cannot be presented as a matching authenticated upstream identity by the offline dial path. The upstream Login packet is a newly encoded self-signed session, so its raw size differs from the downstream Login packet even though later same-protocol gameplay packets in successful local captures are forwarded byte-for-byte. It is not evidence of packet mutation, public-server enforcement, or an anti-cheat cause. The correct local validation for this mode is an upstream with authentication disabled. A server that verifies XUID requires an explicitly authorized authenticated upstream flow.

An independent upstream report, [gophertunnel issue #407](https://github.com/Sandertv/gophertunnel/issues/407), describes a related key-binding boundary in a later BDS OIDC flow. The report says the multiplayer token's `cpk` was bound to the client key, while a proxy needed its own key to decrypt the second connection. It records that raw login forwarding reached world generation but then failed because the encryption keys did not match. The report covers BDS protocol 944 and is not evidence about The Hive, protocol 2193, or any server-side enforcement decision. It is useful corroborating evidence that copying or re-signing authenticated login metadata is not a safe parity fix without a protocol-supported design.

## Validation

- `go test ./internal/proxy -run '^TestRunnerForwardsBidirectionalPacketsAndCapturesSession$' -count=1 -v` passed after the regression assertion was added.
- A temporary loopback run against PocketMine with `xbox-auth=off` completed direct and proxied login, resource-pack exchange, spawn, and report capture. The temporary client and server setting were removed or restored after the run.
- The rebuilt-binary comparison used labeled `direct` and `bedrockdebugproxy` reports. The comparison showed zero field differences and zero packet-count differences.
- ProxyFingerprint commit [`323aa03de`](https://github.com/NhanAZ/PocketMine-MP/commit/323aa03de) now includes lifecycle milestone timing in comparison output. The timing view is an observation aid for repeated local runs and is not treated as an identity signal or evidence of a third-party enforcement rule.
- `TestClearXBLIdentityDataRemovesAuthenticationBoundFields` passes in the nested gophertunnel module, and the proxy forwarding regression remains green.
- The upstream issue was reviewed on 2026-09-21 as an external interoperability report. No source code was copied and no public server was contacted.

## Limitations

The test used a synthetic offline gophertunnel client and an operator-controlled local server. It does not establish how a public server classifies a proxy, and it does not replace a real vanilla-client validation of the Xbox-authenticated path.
