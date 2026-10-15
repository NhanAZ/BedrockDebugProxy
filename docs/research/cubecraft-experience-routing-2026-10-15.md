# CubeCraft Experience routing during the v0.4.0 release check

On 2026-10-15, the clean v0.4.0 candidate stamped from `0b597914c7b20eb988a2d7840bb0e91a4d908722` resolved `experience:CubeCraft` through authenticated Minecraft services but stopped before opening a listener or contacting CubeCraft. The join response identified `networkProtocol` as `NetherNet`, while the current resolver found no usable `netherNetId` or IPv4 destination. Two candidate attempts returned the same error.

A temporary local diagnostic inspected field names and value shapes without printing response values or tokens. The response contained `destinationInfo`, `ipV4Address`, `networkProtocol`, and `port`. Its `destinationInfo` contained creator and experience identifiers, but no transport identifier. The `ipV4Address` value was neither a parsed IPv4 address, a UUID, nor a decimal unsigned 64-bit identifier. Its exact value was not retained in this record. The diagnostic source was removed after inspection, and the candidate runtime was not changed.

This establishes a mismatch between the service response and `internal/experience/target.go` for this attempt. It does not establish whether the service intended a different NetherNet signaling route, supplied an unavailable destination, or returned an undocumented transient form. The published endpoint reference in `docs/server-targets.md` is not a substitute for the authenticated join result. Do not reinterpret a `NetherNet` response as RakNet or guess a destination to make the release check pass.

The [Mojang NetherNet onboarding guide](https://github.com/Mojang/bedrock-protocol-docs/blob/main/additional_docs/NetherNetOnboardingGuide.md), inspected on 2026-10-15, documents an HTTP signaling route for NetherNet servers. The installed [go-nethernet v1.0.24 dialer](https://github.com/df-mc/go-nethernet/blob/v1.0.24/dial.go) documents the transport dependency used by the current Experience path. Neither source confirms the meaning of this CubeCraft join response. No source was copied.

The first five retail release reports passed for the exact candidate runtime. CubeCraft has no capture or validation report because no connection was started. The six-server release gate therefore remains incomplete. Recheck the service response against a traceable current contract and a controlled NetherNet test before changing the resolver. A successful later routing change would require fresh runtime validation of all six targets.

## Follow-up on the authenticated route

The follow-up investigation on 2026-10-15 classified `ipV4Address` as a hostname with a valid port. A read-only capability request to the exact advertised host and port returned JSON status with HTTP `200` over HTTPS; plaintext HTTP returned `426`. This supports the official HTTPS NetherNet endpoint flow for that destination. No authentication token was sent to the status endpoint, and the hostname value was not retained in this note. The installed `go-nethernet` v1.0.24 endpoint client already implements the status and SDP exchange path.

The focused fix selects that HTTPS path only when the authenticated join response declares `NetherNet` and supplies a host and port without an ID. It preserves ID-based signaling and rejects invalid destinations or failed HTTPS status checks. A development build reached the listener after the status check. A separate `--packs-only` run completed authenticated upstream login and retained the offered packs, then closed without world spawn. These scoped checks do not establish retail-client spawn or release readiness; those require validation from a clean stamped revision.

## Retail follow-up and optional ping

A stamped retail session at revision `973f72dadad8b4931af7c25b5b219cc49c40a003` reached spawn, forwarded traffic in both directions, and closed with a verified capture. Its sanitized report could not pass because the optional pre-dial `transport.ping` failed with `NetherNet.PingContext: not supported`. The vendored gophertunnel dialer ignored that ping failure and continued to the successful HTTPS connection. This is not evidence of a gameplay failure. The HTTP endpoint status check had already succeeded during Experience resolution.

The follow-up fix uses the dialer's existing `SkipPing` option only when the selected target starts with `https://`. The skipped operation has no response bytes to preserve. RakNet routes keep their ping and capture behavior. Fresh exact-revision retail evidence is needed after this change.
