# Server targets and validation catalog

This catalog records the server selectors and endpoint references used by the project. An endpoint is an operational reference, not a guarantee that the service is available or that the service permits proxy connections. Experience resolution may discover a different address at runtime.

## Proxy-use disclaimer

BedrockDebugProxy is a read-only observation and research tool, but that does not make proxy use authorized by a server. Each server controls its own rules, terms, authentication, and anti-abuse systems. A server may reject a connection, disconnect or kick a player, or issue a temporary or permanent ban even when the proxy does not intentionally modify, inject, replay, or automate gameplay traffic. Check the server's current policy before connecting through a proxy. An explicit maintainer request to run an owner-controlled session is sufficient authorization for this internal workflow. Do not request, store, or publish private server messages. The project does not provide authorization to use any listed service.

### The Hive policy status

The Hive's [VPN and proxy policy](https://support.playhive.com/vpn/) was reviewed on 2026-09-20. It says that VPN and proxy use is not allowed, most proxy logins are automatically rejected, and blocklists are updated continually. It also says that proxy use alone should not result in a ban. This is a server policy boundary, not evidence that a proxy connection is safe for project validation.

The project must not use a punished account or another account to probe The Hive's enforcement behavior. Existing Hive captures are historical evidence. A future live baseline may proceed when the maintainer directly requests the internal validation session. Keep the report focused on observed technical evidence. Do not substitute another account or claim that the public policy changed.

## Active pre-release validation matrix

These six targets are the required release matrix. Use the exact selector with the stamped binary and record the observed endpoint in the local report rather than assuming that the reference remains current. CubeCraft is included without a separate repeated policy confirmation.

| Report label | Upstream selector | Endpoint reference | Category |
| --- | --- | --- | --- |
| The Hive | `experience:The Hive` | `geo.hivebedrock.network:19132` | Creator Experience |
| Galaxite | `experience:Galaxite` | `play.galaxite.net:19132` | Creator Experience |
| Lifeboat | `experience:Lifeboat` | `mco.lbsg.net:19132` | Creator Experience |
| Mineville Zeqa | `experience:Mineville Zeqa` | `play.inpvp.net:19132` | Creator Experience |
| Enchanted | `experience:Enchanted` | `play.enchanted.gg:19132` | Creator Experience |
| CubeCraft | `experience:CubeCraft` | `play.cubecraft.net:19132` | Featured server |

For release validation, follow [`validation.md`](validation.md) and [`releasing.md`](releasing.md). Use a concrete loopback address for the agent-controlled protocol client when testing `--follow-transfers`; a retail client is not required. Use Golem CLI only for disposable local server/plugin scenarios, not as a public-server network client. Do not place raw addresses, credentials, or captures in sanitized reports.

Enchanted has a context-aware release flow beyond hub connectivity. The server sends a form automatically when the player joins, and that form is the authority for the current child list and status. Inventory the entries without assuming a fixed order or fixed names. Classify each entry as an available child, a maintenance or unavailable item, a navigation item, or an informational/status item.

For every actionable top-level child, start from a fresh Hub session, select the current form entry, and classify the result. If the server sends `Transfer`, follow the transfer through the next hop, complete every offered pack exchange, reach spawn, and retain at least 30 seconds of bidirectional traffic. If the server sends another form, treat it as a nested menu and select one currently available child from that form; repeat the same classification if another form appears. If the server sends no `Transfer`, inspect whether the selection produced a direct spawn, a status or maintenance response, a navigation action, or an unchanged Hub session. Do not mark the route failed merely because it has no hop 2.

Return to the Hub before testing the next top-level child. The Enchanted pass decision requires a reviewed form inventory, an outcome for every actionable entry, at least one playable route that reaches spawn and the stability interval, and no unexplained failure. Transfer and child-pack checks apply only when the observed route actually uses those flows. A route explicitly classified as direct or unavailable is recorded as `not_applicable` for those conditional checks, not as an artificial transfer failure.

The Enchanted child menu is dynamic. `Steal a Mob` is no longer an Enchanted child; its successor, `Rob a Mob`, is a separate Featured Experience. Do not include either name as an Enchanted child-selection step. The Enchanted release flow requires at least one currently playable route, not one fixed historical child.

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
| SurvivalOnline Beta | `experience:SurvivalOnline Beta` | Provider and endpoint not recorded |

The `experience:` selector is the common Minecraft services lookup for both Featured Experiences and Creator Experiences. The category column describes the server listing, not which selector form it accepts. CubeCraft is a Featured server and its release selector is `experience:CubeCraft`; its endpoint is listed as a reference only. Provider and endpoint references in this document came from the maintainer's test catalog and must be rechecked if a service changes its routing.

## CubeCraft policy boundary

CubeCraft is included in the owner-controlled release test without a separate repeated risk warning. Do not record, infer, or publish any private server discussion or claim that CubeCraft permits the connection. This workflow choice is not a policy exception and does not change the server's published rules.

CubeCraft's published policies explicitly prohibit proxies in Minecraft Bedrock:

- [CubeCraft Rules](https://help.cubecraft.net/en/article/cubecraft-rules-1403lij/) - reviewed 2026-09-22.
- [Allowed Mods and Clients](https://www.cubecraft.net/threads/allowed-mods-and-clients.228596/) - reviewed 2026-09-22.

The release command uses the normal read-only proxy path and must not add bypass, mutation, replay, or anti-cheat evasion behavior. The connection may still be rejected, kicked, or banned under the published policy. This is an agent workflow boundary, not a product runtime warning.
