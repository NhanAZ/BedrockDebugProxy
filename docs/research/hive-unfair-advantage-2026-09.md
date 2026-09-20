# The Hive Unfair Advantage incident and proxy identity investigation

## Scope and safety

This note records the investigation performed on 2026-09-20 after a BedrockDebugProxy session to The Hive received an `Unfair Advantage` ban response. It separates capture evidence, official server policy, public release information, and hypotheses. It is not a guide for bypassing a punishment or an anti-cheat system.

The source captures remain local and sensitive. This note intentionally omits XUIDs, identity UUIDs, authentication data, packet payloads, and resource-pack keys. No additional live Hive test should be performed with the punished account or another account while proxy use is not authorized.

## Incident evidence

The affected capture is `captures/session-20260920T135412Z`.

- The capture is closed and verified with 93,461 events and 37,596 blobs.
- The stamped generator commit is `1b7ff3a0b55ce5e8f5eea60b6e793cd29aab74bf`.
- The upstream selector is `experience:The Hive`, resolved through RakNet at `geo.hivebedrock.network:19132`.
- The proxy recorded the upstream identity display name `SrDuck84` and the downstream identity display name `NhanAZ` in separate login metadata events.
- The server sent `You are banned from The Hive` with reason `Unfair Advantage` and an expiry of approximately seven days. The capture records the server response; it does not identify the server-side detector.
- Same-protocol traffic was forwarded through the raw path. The capture does not show an intentional gameplay mutation, injection, replay, or transfer-address rewrite before the ban.

## Comparative capture evidence

A repository-wide search found 23 Hive captures containing the upstream display name `SrDuck84`. Twenty-two ended without a Hive ban and one is the affected capture above.

The successful capture `captures/session-20260917T034859Z` is the closest comparison. It also contains the `NhanAZ` downstream and `SrDuck84` upstream identity pattern, the same Hive selector, protocol 2193, and the same upstream game baseline. Its options are not identical: it used a different listen/authentication setup, did not follow transfers, and enabled pack decryption. Those differences prevent a claim that the two sessions were identical.

The clearest client-side difference is in the login metadata. All successful Hive captures inspected for this comparison advertise client game version `1.26.50`; the affected capture advertises `1.26.51`. The Hive proxy baseline remains protocol 2193 and game version `1.26.50` in both manifests.

The successful viewer capture was generated at commit `76dd00fd0ae105a8bc44d3c8182787c4f54f6e51`. A source comparison from that commit to the affected commit found no changes under `internal/proxy`, `internal/bedrock`, `cmd`, or `third_party`. Recent viewer-only commits therefore do not explain the network outcome.

## Official policy and release evidence

The Hive's [VPN and proxy policy](https://support.playhive.com/vpn/) says that VPN and proxy use is not allowed, most proxy logins are automatically rejected, and proxy use alone should not result in a ban. The same page says that blocklists are updated continually and can cause a connection that previously worked to be blocked later. This distinguishes a proxy blocklist change from a gameplay anti-cheat decision.

The Hive's [client modification policy](https://support.playhive.com/allowed-and-denied-mods/) says that unlisted modifications should be assumed disallowed. Its categorical proxy entry specifically covers proxies that let Java Edition clients join Bedrock servers, so that item should not be overextended to prove that every Bedrock-to-Bedrock relay is covered. The broader VPN and proxy policy still applies to this connection.

The Hive [supports client version 26.51](https://support.playhive.com/what-versions-does-the-hive-support/) and last updated that list on 2026-09-16. Mojang's [Bedrock protocol changelog](https://mojang.github.io/bedrock-protocol-docs/changelog/) lists stable 1.26.51 as protocol 2193 with no packet changes. This makes an unsupported-version explanation unlikely, but it does not rule out client behavior or timing differences outside the packet schema.

The public [Hive changelog](https://updates.playhive.com/changelog) shows a cosmetics update on 2026-09-17 and the BedWars Season 5 update on 2026-09-04. No public anti-cheat change was found in the entries reviewed around the incident. A private detector or blocklist rollout remains possible because public release notes are not a complete record of server-side enforcement changes.

## Findings

### Confirmed

1. The Hive sent a temporary `Unfair Advantage` ban response during the affected proxy session.
2. Hive policy does not permit VPN or proxy use. The project has no authorization from Hive to use this proxy path.
3. The same `NhanAZ` to `SrDuck84` identity pattern appeared in many earlier successful Hive captures.
4. The affected session is the only inspected session with client metadata advertising `1.26.51`.
5. The recent viewer commits did not change the runtime proxy, protocol, or bridge packages.

### Plausible but unproven

1. Hive may have changed a server-side proxy blocklist or anti-cheat detector after the earlier successful sessions.
2. The `SrDuck84` upstream identity and the two-session authenticated proxy boundary may be an interaction factor that became detectable after such a change.
3. Client 1.26.51 may have changed timing or behavior that interacts with the proxy boundary, even though the protocol ID and packet schemas did not change.

### Not established

1. The exact Hive detector, rule, or packet that produced the ban.
2. Whether `SrDuck84` was the direct trigger, a contributing signal, or unrelated to the ban.
3. Whether the cause was a server-side rollout, client behavior, proxy identity presentation, or a combination.
4. Whether any code-only change could make this proxy path acceptable under Hive policy.

## Maintainer decision and follow-up

The previous statement that `SrDuck84` could be excluded was too strong. The correct conclusion is that `SrDuck84` remains a leading interaction hypothesis, but the existing captures disprove the simpler claim that this identity alone always causes a ban.

No runtime change is justified by this evidence. Preserve the affected capture, do not retest a punished account or substitute another account, and treat any future Hive validation as incomplete unless current policy review and explicit authorization permit the proxy connection. If authorized testing becomes available, compare a stamped binary, client version, downstream and upstream login metadata, timing, and the server response without attempting to evade enforcement.

## Sources and provenance

- The Hive VPN and proxy policy, last updated 2026-08-18, accessed 2026-09-20.
- The Hive allowed and denied client modifications policy, last updated 2026-07-06, accessed 2026-09-20.
- The Hive supported-version page, last updated 2026-09-16, accessed 2026-09-20.
- The Hive public changelog, entries reviewed through 2026-09-20.
- Mojang Bedrock protocol changelog, stable 1.26.51 entry, accessed 2026-09-20.
- Local closed captures listed above and the Git history comparison between the two capture revisions. Raw payloads and authentication material remain outside this note.
