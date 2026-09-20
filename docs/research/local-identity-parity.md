# Local identity parity research

## Question

Does BedrockDebugProxy preserve the downstream identity when it dials an operator-controlled unauthenticated Bedrock server?

## Confirmed observation

On 2026-09-21, an offline gophertunnel client named `LocalProbe` connected directly to the local Axolotl-PM 5.49.2+dev server and then through BedrockDebugProxy. The local server used `xbox-auth=off` only for this loopback test and was restored to `xbox-auth=on` afterward.

Before the fix, the proxied connection reached the local server as `Steve`, which is gophertunnel's default display name when no identity is supplied to an unauthenticated `Dialer`. The direct connection reached the server as `LocalProbe`.

The fix passes the identity supplied to `FetchResourcePacks` into the upstream `minecraft.Dialer`. After the fix, both direct and proxied local connections reached the server as `LocalProbe`. The upstream fake-server regression test also confirms that the downstream identity UUID and display name are preserved.

## Source path

- `internal/proxy/proxy.go` receives `login.IdentityData` in the gophertunnel `FetchResourcePacks` callback and now passes it to `connectUpstream`.
- `internal/proxy/proxy.go` sets `minecraft.Dialer.IdentityData` for the upstream dial.
- `third_party/gophertunnel/minecraft/dial.go` replaces `Dialer.IdentityData` only when an Xbox-authenticated token source supplies a verified identity. The unauthenticated path keeps the supplied identity.

## Interpretation

This was a local protocol and identity-parity defect, not evidence about any public server's anti-cheat. Passing the identity is safe for the local unauthenticated path and does not override authenticated Xbox identity data. The change does not mutate packets or attempt to evade server policy.

## Validation

- `go test ./internal/proxy -run '^TestRunnerForwardsBidirectionalPacketsAndCapturesSession$' -count=1 -v` passed after the regression assertion was added.
- A temporary loopback run against PocketMine with `xbox-auth=off` completed direct and proxied login, resource-pack exchange, spawn, and report capture. The temporary client and server setting were removed or restored after the run.

## Limitations

The test used a synthetic offline gophertunnel client and an operator-controlled local server. It does not establish how a public server classifies a proxy, and it does not replace a real vanilla-client validation of the Xbox-authenticated path.
