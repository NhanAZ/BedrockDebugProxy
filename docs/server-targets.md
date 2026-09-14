# Server targets and validation catalog

This catalog records the server selectors and endpoint references used by the project. An endpoint is an operational reference, not a guarantee that the service is available or that the service permits proxy connections. Experience resolution may discover a different address at runtime.

## Proxy-use disclaimer

BedrockDebugProxy is a read-only observation and research tool, but that does not make proxy use authorized by a server. Each server controls its own rules, terms, authentication, and anti-abuse systems. A server may reject a connection, disconnect or kick a player, or issue a temporary or permanent ban even when the proxy does not intentionally modify, inject, replay, or automate gameplay traffic. Check the server's current policy and obtain permission before connecting through a proxy. The project does not provide authorization to use any listed service.

## Active pre-release validation matrix

These five targets are the required release matrix. Use the exact selector with the stamped binary and record the observed endpoint in the local report rather than assuming that the reference remains current.

| Report label | Upstream selector | Endpoint reference | Category |
| --- | --- | --- | --- |
| The Hive | `experience:The Hive` | `geo.hivebedrock.network:19132` | Creator Experience |
| Galaxite | `experience:Galaxite` | `play.galaxite.net:19132` | Creator Experience |
| Lifeboat | `experience:Lifeboat` | `mco.lbsg.net:19132` | Creator Experience |
| Mineville Zeqa | `experience:Mineville Zeqa` | `play.inpvp.net:19132` | Creator Experience |
| Enchanted | `experience:Enchanted` | `play.enchanted.gg:19132` | Creator Experience |

For release validation, follow [`validation.md`](validation.md) and [`releasing.md`](releasing.md). Use a concrete reachable LAN address for `--listen` when testing `--follow-transfers`. Do not place raw addresses, credentials, or captures in sanitized reports.

Enchanted has a required release flow beyond hub connectivity. The server sends its form automatically when the player joins. Select any available minigame in that form and observe the resulting `Transfer` and next-hop resource-pack exchange before recording a passing report. A hub-only session is not sufficient.

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

The five active targets above also use Creator Experience selectors. The provider and endpoint references in this document came from the maintainer's current test catalog and must be rechecked if a service changes its routing.

## CubeCraft exclusion

CubeCraft is intentionally excluded from the active validation matrix, pull request recommendations, release gates, and proposed test targets. Existing CubeCraft notes in `docs/research/` are historical evidence only and do not authorize a new connection.

CubeCraft's published policies explicitly prohibit proxies in Minecraft Bedrock:

- [CubeCraft Rules](https://help.cubecraft.net/en/article/cubecraft-rules-1403lij/) - reviewed 2026-09-14.
- [Allowed Mods and Clients](https://www.cubecraft.net/threads/allowed-mods-and-clients.228596/) - reviewed 2026-09-14.

If the maintainer explicitly asks to test CubeCraft, the agent must surface this policy and the possible connection, kick, or ban risk before proceeding. This is an agent workflow warning, not a runtime warning and not a bypass recommendation.
