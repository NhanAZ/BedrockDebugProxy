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

## Validation

- `go test ./internal/proxy -run '^TestRunnerForwardsBidirectionalPacketsAndCapturesSession$' -count=1 -v` passed after the regression assertion was added.
- A temporary loopback run against PocketMine with `xbox-auth=off` completed direct and proxied login, resource-pack exchange, spawn, and report capture. The temporary client and server setting were removed or restored after the run.
- The rebuilt-binary comparison used labeled `direct` and `bedrockdebugproxy` reports. The comparison showed zero field differences and zero packet-count differences.

## Limitations

The test used a synthetic offline gophertunnel client and an operator-controlled local server. It does not establish how a public server classifies a proxy, and it does not replace a real vanilla-client validation of the Xbox-authenticated path.
