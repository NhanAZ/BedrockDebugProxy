# Oomph proxy-detection research

## Scope and date

This note records the public Oomph source review performed on 2026-09-13. The goal is to understand whether the project contains evidence that Bedrock proxy connections can be classified separately from official clients. It is a protocol and compatibility research note, not a guide for evading an anti-cheat system.

The reviewed Oomph revision is commit `815f8038cd3c59c29457e1ce35f1cc610a5c2773` on the `stable` branch. The repository's root `LICENSE` is the Server Side Public License (SSPL). No Oomph source was copied into BedrockDebugProxy and it is not used as a dependency.

## Maintainer identity evidence

The Oomph README credits [ethaniccc](https://github.com/ethaniccc) with creating its movement and combat validation systems and links to that GitHub account. Ethan's public profile also lists Oomph as a project. This confirms the repository relationship, but it does not independently prove that a Discord account named `actual Ben` is the same person. That identity claim remains community-provided.

## Strongest clue: proxy detections are an explicit anti-cheat category

The default Oomph configuration contains the following named detections in [`anticheat/oconfig/config.go`](https://github.com/oomph-ac/oomph/blob/815f8038cd3c59c29457e1ce35f1cc610a5c2773/anticheat/oconfig/config.go):

- `Proxy_A` describes a player as likely using a game proxy and is configured to kick after a violation threshold.
- `Proxy_B` describes a player as connected with a game proxy and is configured to ban after one violation.
- `Cloud_Proxy` describes a player as using a game proxy and is configured as a cloud-managed ban detection.

This is direct evidence that Oomph treats proxy use as a first-class detection surface, not merely as a transport oddity. It does not prove that CubeCraft runs Oomph or that CubeCraft's Sentinel uses the same rules.

## Important limitation in the public tree

The public detection registration in [`anticheat/player/detection/register.go`](https://github.com/oomph-ac/oomph/blob/815f8038cd3c59c29457e1ce35f1cc610a5c2773/anticheat/player/detection/register.go) registers the edition, packet, movement, combat, reach, and hitbox detectors, but does not register `Proxy_A`, `Proxy_B`, or `Cloud_Proxy`. A source search at the same revision also found no public implementation files for those names.

Therefore the exact proxy fingerprint is unknown from this repository. The configuration entries may be reserved for a cloud or private detector, may be incomplete in the public branch, or may be stale configuration. We must not infer the algorithm from the names alone.

## Observable client metadata that an anti-cheat can validate

Oomph's public edition-faker checks show that a server-side intermediary can validate consistency between login metadata fields:

- [`edition_faker_a.go`](https://github.com/oomph-ac/oomph/blob/815f8038cd3c59c29457e1ce35f1cc610a5c2773/anticheat/player/detection/edition_faker_a.go) compares the identity title ID with the declared device operating system.
- [`edition_faker_b.go`](https://github.com/oomph-ac/oomph/blob/815f8038cd3c59c29457e1ce35f1cc610a5c2773/anticheat/player/detection/edition_faker_b.go) compares the default input mode with the device operating system.
- [`edition_faker_c.go`](https://github.com/oomph-ac/oomph/blob/815f8038cd3c59e1ce35f1cc610a5c2773/anticheat/player/detection/edition_faker_c.go) rejects invalid input-mode combinations.

These checks are not evidence of CubeCraft's implementation. They are evidence that device and identity metadata are practical anti-cheat inputs, even when gameplay packets are not intentionally modified.

## Why a gophertunnel relay can still be observable

The Oomph fork of gophertunnel at submodule commit `2974690ef0c637996a1c51fec48427e4998a72ff` retains the normal authenticated dial behavior. In its [`minecraft/dial.go`](https://github.com/oomph-ac/gophertunnel/blob/2974690ef0c637996a1c51fec48427e4998a72ff/minecraft/dial.go), the dialer creates a new ECDSA key for the upstream session, sets client data for the dialled server, and uses the authenticated login path's platform data. The same code documents that forwarding XUID and title ID without a valid authentication token is not technically valid and may be rejected by servers.

The upstream dial therefore creates a second authenticated Bedrock session. A read-only forwarding policy describes packet intent, but it does not make the two handshakes, keys, authentication chain, endpoint fields, batching, timing, and encryption state identical to one official client connection.

## Oomph's intended deployment model

Oomph's [`transferproxy` README](https://github.com/oomph-ac/oomph/blob/815f8038cd3c59c29457e1ce35f1cc610a5c2773/transferproxy/README.md) describes a proxy that preserves a public client connection while switching between backend servers. Its [setup guide](https://github.com/oomph-ac/oomph/blob/815f8038cd3c59c29457e1ce35f1cc610a5c2773/anticheat/docs/setup.md) documents a public Oomph listener authenticating the client and a private backend that accepts proxy connections, with backend Xbox authentication disabled where appropriate. This is an owner-controlled server topology, not an interoperability promise for connecting through a proxy to an unrelated public network such as CubeCraft.

The transfer proxy also keeps XBL identity data when it dials a backend in [`transferproxy/config.go`](https://github.com/oomph-ac/oomph/blob/815f8038cd3c59c29457e1ce35f1cc610a5c2773/transferproxy/config.go). That behavior is useful evidence about the identity-forwarding problem, but it is not a supported recipe for changing a public server's anti-cheat decision.

## Correlation with the CubeCraft capture

The local CubeCraft ban capture `session-20260912T184553Z` recorded the following high-level differences between the two terminated sessions. Sensitive identifiers are intentionally omitted here:

- The downstream login advertised the local proxy address in `ClientData.ServerAddress`; the upstream login advertised the CubeCraft address.
- Upstream identity data contained title and PlayFab metadata that was absent from the downstream identity record.
- The downstream and upstream identity UUIDs were different, while the display name, XUID, Android device classification, and touch input mode matched.

This rules out a simple Windows-versus-Android mismatch for that incident. It does show that the proxy was not an indistinguishable relay at the authenticated session boundary. The exact Sentinel trigger remains unproven. Candidate classes include authentication or key rebinding, identity metadata, endpoint fields, packet scheduling or batching, and a coarse proxy heuristic.

## Findings and project implications

### Confirmed

1. Oomph publicly models proxy use as a detection category, including a cloud-managed proxy detection.
2. Oomph publicly validates device and identity metadata consistency.
3. Oomph is designed for server operators who control both the public proxy and private backend.
4. BedrockDebugProxy's CubeCraft capture contains observable differences at the two-session authentication boundary.

### Unknown

1. Whether CubeCraft uses Oomph, shares code with it, or uses a separate Sentinel implementation.
2. Whether `actual Ben` is the same person as the GitHub account linked by Oomph.
3. Which exact signal caused the seven-day CubeCraft punishment.
4. Whether any code-only change can make a proxy acceptable under CubeCraft's rules.

### Next safe research step

Do not use additional public accounts to search for a Sentinel bypass. Preserve the existing capture and, if needed, add a local diagnostic report that compares downstream and upstream `ClientData`, identity metadata, protocol negotiation, packet ordering, batching, and timing without altering the live wire behavior. Questions about identity-preserving authenticated MITM support should be directed to gophertunnel or server maintainers using an owner-controlled test server.

## Sources and provenance

- [Oomph repository at reviewed revision](https://github.com/oomph-ac/oomph/tree/815f8038cd3c59c29457e1ce35f1cc610a5c2773), SSPL, accessed 2026-09-13.
- [Oomph README](https://github.com/oomph-ac/oomph/blob/815f8038cd3c59c29457e1ce35f1cc610a5c2773/README.md), accessed 2026-09-13.
- [Oomph anti-cheat configuration](https://github.com/oomph-ac/oomph/blob/815f8038cd3c59c29457e1ce35f1cc610a5c2773/anticheat/oconfig/config.go), accessed 2026-09-13.
- [Oomph detector registration](https://github.com/oomph-ac/oomph/blob/815f8038cd3c59c29457e1ce35f1cc610a5c2773/anticheat/player/detection/register.go), accessed 2026-09-13.
- [Oomph transfer proxy README](https://github.com/oomph-ac/oomph/blob/815f8038cd3c59c29457e1ce35f1cc610a5c2773/transferproxy/README.md), accessed 2026-09-13.
- [Oomph setup guide](https://github.com/oomph-ac/oomph/blob/815f8038cd3c59c29457e1ce35f1cc610a5c2773/anticheat/docs/setup.md), accessed 2026-09-13.
- [Oomph fork of gophertunnel dialer](https://github.com/oomph-ac/gophertunnel/blob/2974690ef0c637996a1c51fec48427e4998a72ff/minecraft/dial.go), MIT, accessed 2026-09-13.
