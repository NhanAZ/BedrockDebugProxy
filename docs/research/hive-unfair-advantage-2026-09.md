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
- The client error panel also shows `InitialConnection-55`. No public Hive or Mojang mapping for `InitialConnection-55` was found. In this screenshot it is a client-side diagnostic label shown alongside the ban response, not an independent explanation of the punishment.
- The same panel reports `RakNet:2193`, which is the transport and Bedrock protocol metadata. It is not itself a ban code. `NetworkType`, `Connection`, `WorldSession`, and `Client` are diagnostic fields whose exact meanings are not publicly documented; their values are intentionally omitted from this note.
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

## Public community reports

The public discussions below show that other players have received similar classifications or connection flags. They do not prove that any one report has the same cause as this incident.

- In [CPS Limits & False Bans](https://forum.playhive.com/t/cps-limits-false-bans-a-console-player-s-suggestion/94635), a PS5 player reported a one-off `Unfair Advantage` ban that they attributed to unintended controller clicks. A Hive forum staff reply said that specific anti-cheat information is not disclosed and recommended lowering click speed if repeated kicks occur.
- In [Banned For No Reason And My Appeal Got DENIED](https://forum.playhive.com/t/banned-for-no-reason-and-my-appeal-got-denied/99621), a PC player reported a one-week `Unfair Advantage` ban and suspected a false anti-cheat positive. The thread response did not identify the detector and advised waiting after a denied appeal.
- In [Replace Lag Bann with Kick](https://forum.playhive.com/t/replace-lag-bann-with-kick/94161), a player described week-long bans after bad connections and claimed that support sometimes lifted the first punishment. This is a user report, not a Hive technical statement.
- In [Vpn or proxy detected](https://forum.playhive.com/t/vpn-or-proxy-detected/91515), a PC player reported being incorrectly identified as using a VPN. The staff response redirected them to the official VPN policy. This supports the existence of false proxy flags, but it is a connection block report, not an `Unfair Advantage` ban report.
- A 2026 Reddit thread, [I can't join Hive, this keeps happening](https://www.reddit.com/r/hivemc/comments/1toyqph/i_cat_join_hive_this_keeps_happening/), describes repeated connection failures, occasional VPN/proxy messages, and a later 62-day punishment claim. The thread is anecdotal and does not establish the cause.

No indexed public result was found that names BedrockDebugProxy, `SrDuck84`, or this capture's exact hybrid identity pattern. The forum evidence therefore supports the possibility of false positives or connection-sensitive enforcement, but it does not establish a Hive anti-cheat rollout or identify the cause of this ban.

## Findings

### Confirmed

1. The Hive sent a temporary `Unfair Advantage` ban response during the affected proxy session.
2. Hive policy does not permit VPN or proxy use. The project has no authorization from Hive to use this proxy path.
3. The same `NhanAZ` to `SrDuck84` identity pattern appeared in many earlier successful Hive captures.
4. The affected session is the only inspected session with client metadata advertising `1.26.51`.
5. The recent viewer commits did not change the runtime proxy, protocol, or bridge packages.
6. Public Hive discussions contain other reports of `Unfair Advantage` false positives and VPN/proxy misclassification, but none identifies this proxy or identity pattern.

### Plausible but unproven

1. Hive may have changed a server-side proxy blocklist or anti-cheat detector after the earlier successful sessions.
2. The `SrDuck84` upstream identity and the two-session authenticated proxy boundary may be an interaction factor that became detectable after such a change.
3. Client 1.26.51 may have changed timing or behavior that interacts with the proxy boundary, even though the protocol ID and packet schemas did not change.

### Not established

1. The exact Hive detector, rule, or packet that produced the ban.
2. Whether `SrDuck84` was the direct trigger, a contributing signal, or unrelated to the ban.
3. Whether the cause was a server-side rollout, client behavior, proxy identity presentation, or a combination.
4. Whether any code-only change could make this proxy path acceptable under Hive policy.
5. Whether the community reports share a detector, connection condition, or enforcement rollout with this incident.
6. The internal meaning of the client's `InitialConnection-55` diagnostic code.

## Maintainer decision and follow-up

The previous statement that `SrDuck84` could be excluded was too strong. The correct conclusion is that `SrDuck84` remains a leading interaction hypothesis, but the existing captures disprove the simpler claim that this identity alone always causes a ban.

No runtime change is justified by this evidence. Preserve the affected capture, do not retest a punished account or substitute another account, and treat any future Hive validation as incomplete unless current policy review and explicit authorization permit the proxy connection. If authorized testing becomes available, compare a stamped binary, client version, downstream and upstream login metadata, timing, and the server response without attempting to evade enforcement.

## Official appeal and support routes

The Hive's [Punishments and Appeals](https://support.playhive.com/ive-been-banned-muted/) page says to use the official [Player Support form](https://contact.playhive.com/) and select `Appeal In-Game Punishment` for an in-game ban. The form requires the Minecraft username, the punishment ID, the punishment reason, and an explanation of why the punishment is disputed. The page says that only the punished player or their parent or guardian should submit the appeal, that decisions are final, and that additional appeals for the same punishment are not answered.

The [Punishment ID guide](https://support.playhive.com/punishment-id/) says the Ban ID shown on the join error screen is the required punishment ID. If the ID is missing, a 2026 Hive forum staff reply says to select `Game Ban` in the appeal form because that route does not require a punishment ID. The screenshot for this incident contains additional connection and client identifiers. Those values should be submitted only through the official private form if requested, not posted publicly or added to the repository.

The separate [VPN exemption route](https://support.playhive.com/vpn/) is under `Technical Support` and `VPN Exemption Request`. It is a request for future access and is not a substitute for appealing the current ban. An appeal should be factual and should disclose the proxy use rather than claiming that the proxy was permitted.

## Sources and provenance

- The Hive VPN and proxy policy, last updated 2026-08-18, accessed 2026-09-20.
- The Hive allowed and denied client modifications policy, last updated 2026-07-06, accessed 2026-09-20.
- The Hive supported-version page, last updated 2026-09-16, accessed 2026-09-20.
- The Hive public changelog, entries reviewed through 2026-09-20.
- Mojang Bedrock protocol changelog, stable 1.26.51 entry, accessed 2026-09-20.
- Local closed captures listed above and the Git history comparison between the two capture revisions. Raw payloads and authentication material remain outside this note.
