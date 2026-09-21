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

After the fix was rebuilt into `bin/bedrock-debug-proxy.exe`, the real Windows Bedrock client signed in as `SrDuck84` through the proxy to the operator-controlled PocketMine server. The session ran for approximately 220.9 seconds, completed login, resource-pack negotiation, spawn, normal movement, and quit, and captured 4,350 packets. The proxy and upstream both negotiated protocol 2193 and game version 1.26.51 with Windows device metadata.

The closed ProxyFingerprint report for this run contained no `auth_client_playfab_mismatch` signal. Its only heuristic signals were `non_public_peer` and `non_full_authentication`, both expected for the loopback server configured with `xbox-auth=off` and `verify-xuid=false`. The upstream self-signed identity fields were empty while the forwarded ClientData remained structurally valid. This is evidence that the local offline identity-parity defect is fixed for a real vanilla client. It is not evidence about a public server's anti-cheat or acceptance policy.

## Authenticated upstream boundary

The authenticated path has a separate, source-backed boundary. Sandertv/gophertunnel at immutable commit [`7a556a07335b663744b50d38062636ad8283f314`](https://github.com/Sandertv/gophertunnel/blob/7a556a07335b663744b50d38062636ad8283f314/minecraft/dial.go) still calls `setAndroidData` before encoding an authenticated login because the XBL chain's title ID cannot be edited. That dialer creates a new signed upstream session from the token source. Preserving the downstream device metadata without also changing the signed chain would create an internal metadata contradiction, not a vanilla-equivalent login. This project therefore keeps the upstream behavior and treats authenticated two-session parity as an architectural limitation requiring an authorized server-side test, not a target for heuristic evasion.

## Validation

- `go test ./internal/proxy -run '^TestRunnerForwardsBidirectionalPacketsAndCapturesSession$' -count=1 -v` passed after the regression assertion was added.
- A temporary loopback run against PocketMine with `xbox-auth=off` completed direct and proxied login, resource-pack exchange, spawn, and report capture. The temporary client and server setting were removed or restored after the run.
- The rebuilt-binary comparison used labeled `direct` and `bedrockdebugproxy` reports. The comparison showed zero field differences and zero packet-count differences.
- `TestClearXBLIdentityDataRemovesAuthenticationBoundFields` passes in the nested gophertunnel module, and the proxy forwarding regression remains green.

## Limitations

The test used a synthetic offline gophertunnel client and an operator-controlled local server. It does not establish how a public server classifies a proxy, and it does not replace a real vanilla-client validation of the Xbox-authenticated path.
