# NetherNet server identity JWT validation

Research checked on 2026-09-25.

## Sources

| Source | Revision | Evidence |
| --- | --- | --- |
| [`go-nethernet` identity implementation](https://github.com/df-mc/go-nethernet/blob/v1.0.23/identity.go) and [identity tests](https://github.com/df-mc/go-nethernet/blob/v1.0.23/identity_test.go) | Release `v1.0.23`, tag commit `6b1260c879e94129e06a47e716966cfdce9049cb`, MIT | Server identity tokens skip temporal claim validation, still verify the self-signature with the `cpk` key, and have tests covering expired `exp`, future `iat`, and future `nbf`. Client identity tokens retain temporal validation. |
| [`go-nethernet` change that introduced the policy](https://github.com/df-mc/go-nethernet/commit/b5719958958a1d211334984ba520c1eb16fab640) | Commit `b5719958958a1d211334984ba520c1eb16fab640`, MIT | The upstream change explicitly describes skipping server identity temporal validation to match Bedrock behavior. |
| [Mojang NetherNet onboarding guide, Server Identity JWT](https://github.com/Mojang/bedrock-protocol-docs/blob/c5a73c3b275f3c32366f0c2048643f80e6a13ad8/additional_docs/NetherNetOnboardingGuide.md#server-identity-jwt) | Commit `c5a73c3b275f3c32366f0c2048643f80e6a13ad8`, checked 2026-09-25 | The guide says the client rejects clearly expired server tokens. This conflicts with the behavior described by the upstream implementation and its tests. |

## Confirmed implementation behavior

In `go-nethernet` v1.0.23, `claimPublicKey` validates standard JWT claims only when `selfSigned` is false. For server identity tokens, the library extracts `cpk`, verifies the JWT signature using that key, and does not apply temporal checks to `iat`, `nbf`, or `exp`. The upstream tests explicitly accept a server token with an expired `exp` and future `iat` and `nbf`; they continue to reject expired or future-issued client identity tokens.

BedrockDebugProxy uses the default `go-nethernet` server-token path. `internal/experience/target.go` configures the client identity but does not install a custom `VerifyServerToken` callback, so it does not add stricter temporal checks of its own.

## Decision and limits

This dependency update adopts the upstream compatibility behavior and preserves signature verification. It does not establish that every current retail client accepts every temporally invalid server token. The cited Mojang guide and the upstream implementation disagree, and this project has no live client capture that resolves the discrepancy. Do not describe the guide as corrected or the behavior as independently verified. Revisit the decision if Mojang clarifies the guide or a controlled client test demonstrates different behavior.

The change is limited to the dependency update. No token mutation, packet rewrite, local time-validation override, private capture, or private chat transcript is added.
