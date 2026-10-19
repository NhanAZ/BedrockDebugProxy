# Server targets and validation catalog

This catalog records the server selectors and endpoint references used by the project. An endpoint is an operational reference, not a guarantee that the service is available or that the service permits proxy connections. Experience resolution may discover a different address at runtime.

## Proxy-use disclaimer

BedrockDebugProxy is a read-only observation and research tool, but that does not make proxy use authorized by a server. Each server controls its own rules, terms, authentication, and anti-abuse systems. A server may reject a connection, disconnect or kick a player, or issue a temporary or permanent ban even when the proxy does not intentionally modify, inject, replay, or automate gameplay traffic. Check the server's current policy before connecting through a proxy. An explicit maintainer request to run an owner-controlled session is sufficient authorization for this internal workflow. Do not request, store, or publish private server messages. The project does not provide authorization to use any listed service.

### The Hive policy status

The Hive's [VPN and proxy policy](https://support.playhive.com/vpn/) was reviewed on 2026-09-20. It says that VPN and proxy use is not allowed, most proxy logins are automatically rejected, and blocklists are updated continually. It also says that proxy use alone should not result in a ban. This is a server policy boundary, not evidence that a proxy connection is safe for project validation.

The project must not use a punished account or another account to probe The Hive's enforcement behavior. Existing Hive captures are historical evidence. A future live baseline may proceed when the maintainer directly requests the internal validation session. Keep the report focused on observed technical evidence. Do not substitute another account or claim that the public policy changed.

## Active pre-release validation matrix

These six targets are the required release matrix. Use the exact selector with the stamped binary and record the observed endpoint in the local report rather than assuming that the reference remains current. CubeCraft is additionally subject to the maintainer risk gate below.

| Report label | Upstream selector | Endpoint reference | Category |
| --- | --- | --- | --- |
| The Hive | `experience:The Hive` | `geo.hivebedrock.network:19132` | Creator Experience |
| Galaxite | `experience:Galaxite` | `play.galaxite.net:19132` | Creator Experience |
| Lifeboat | `experience:Lifeboat` | `mco.lbsg.net:19132` | Creator Experience |
| Mineville Zeqa | `experience:Mineville Zeqa` | `play.inpvp.net:19132` | Creator Experience |
| Enchanted | `experience:Enchanted` | `play.enchanted.gg:19132` | Creator Experience |
| CubeCraft | `experience:CubeCraft` | `play.cubecraft.net:19132` | Featured server |

For release validation, follow [`validation.md`](validation.md) and [`releasing.md`](releasing.md). Use a concrete loopback address for a CLI client or a reachable LAN address for a retail client when testing `--follow-transfers`. Do not place raw addresses, credentials, or captures in sanitized reports.

Enchanted has a required release flow beyond hub connectivity. The server sends its form automatically when the player joins. Select a known pack-bearing child, such as OP SMP, in that form and observe the resulting `Transfer` and next-hop resource-pack exchange before recording a passing report. A hub-only session is not sufficient.

## Additional Experience references

These entries are useful for exploratory research but are not release gates.

### Featured Experiences

| Experience | Selector | Provider or endpoint reference |
| --- | --- | --- |
| Treasure Hunt | `experience:Treasure Hunt` | By Enchanted |
| Dimension Clash | `experience:Dimension Clash` | Provider not specified |
| OneBlockOnline | `experience:OneBlockOnline` | By InPvP |
| GenWars | `experience:GenWars` | By InPvP, `play.genwars.com:19132` |
| SoulSteel | `experience:SoulSteel` | Provider not specified |
| Mob Maze | `experience:Mob Maze` | Provider not specified |

### Creator Experiences

| Experience | Selector | Provider or endpoint reference |
| --- | --- | --- |
| MegaSMP | `experience:MegaSMP` | By InPvP, `play.megasmp.gg:19132` |

The `experience:` selector is the common Minecraft services lookup for both Featured Experiences and Creator Experiences. The category column describes the server listing, not which selector form it accepts. CubeCraft is a Featured server and its release selector is `experience:CubeCraft`; its endpoint is listed as a reference only. Provider and endpoint references in this document came from the maintainer's test catalog and must be rechecked if a service changes its routing.

## CubeCraft maintainer risk gate

CubeCraft is included as a release target only for an owner-controlled test where the maintainer explicitly accepts the risk of connecting through a proxy. Do not record, infer, or publish any private server discussion or claim that CubeCraft permits the connection. A risk acknowledgement is not a policy exception and does not change the server's published rules.

CubeCraft's published policies explicitly prohibit proxies in Minecraft Bedrock:

- [CubeCraft Rules](https://help.cubecraft.net/en/article/cubecraft-rules-1403lij/) - reviewed 2026-09-22.
- [Allowed Mods and Clients](https://www.cubecraft.net/threads/allowed-mods-and-clients.228596/) - reviewed 2026-09-22.

Before a CubeCraft session, the maintainer must explicitly accept the possibility of rejection, kick, or ban. Without that acknowledgement, do not connect and record the CubeCraft report as `incomplete`; the release gate must remain blocked. The validation uses the normal read-only proxy path and adds no bypass, mutation, replay, or anti-cheat evasion behavior. This is an agent workflow requirement, not a product runtime warning.
