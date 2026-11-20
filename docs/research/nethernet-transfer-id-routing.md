# NetherNet Transfer ID routing

Reviewed on 2026-11-20 against BedrockDebugProxy revision `6a371a7e92bb70b5696b00c210db32627be5c32e` and the current `go-nethernet` dependency. This is transport-routing research, not a protocol baseline update.

## Sources

| Source | Revision or version | Evidence |
| --- | --- | --- |
| [Microsoft `TransferPlayerNetherNetOptions`](https://learn.microsoft.com/en-us/minecraft/creator/scriptapi/minecraft/server-admin/transferplayernethernetoptions?view=minecraft-bedrock-experimental) | Page last updated 2025-05-20 | `netherNetId` is the destination server's NetherNet ID. |
| [Mojang NetherNet onboarding guide](https://github.com/Mojang/bedrock-protocol-docs/blob/19e25de129227373e97dfa340af93ccf12c0feb9/additional_docs/NetherNetOnboardingGuide.md#server-identity-jwt) | Commit `19e25de129227373e97dfa340af93ccf12c0feb9`, checked 2026-11-20 | The guide describes using `TransferPacket` to move a client between servers that share the same operator key. It does not define the meaning of a zero port in the packet. |
| [go-nethernet `Dialer.DialContext`](https://github.com/df-mc/go-nethernet/blob/4c06b550759d8b488ee209eb09ae87cefd6fe849/dial.go) | Release `v1.0.27`, tag commit `4c06b550759d8b488ee209eb09ae87cefd6fe849`, MIT | The supplied network ID is used as the destination when creating the connection and in the outgoing signal. |
| [gophertunnel `NetherNet.DialContext`](https://github.com/Sandertv/gophertunnel/blob/283a5a97dfe65da94bcc0b401807f6aefa9e72ee/minecraft/nethernet.go) | Release `v1.61.0`, commit `283a5a97dfe65da94bcc0b401807f6aefa9e72ee`, MIT | The network passes its address directly to `go-nethernet.Dialer.DialContext` with its existing signaling instance. |
| `internal/experience/target.go` and `third_party/gophertunnel/minecraft/nethernet.go` | BedrockDebugProxy revision `6a371a7e92bb70b5696b00c210db32627be5c32e` | The service resolver parses `netherNetId` as a UUID and creates an authenticated NetherNet network that retains the signaling connection. |

## Confirmed behavior and inference

The Microsoft API identifies a NetherNet ID as a destination. The pinned Go implementation can dial a destination ID through an existing `minecraft.Network` and signaling connection. Mojang documents `TransferPacket` as one way partners move clients between servers, but neither source specifies that `address=<UUID>, port=0` always represents a NetherNet destination.

The UUID plus zero-port transfer shape was encountered during the current exploratory server acquisition. Its raw session is not archived because it includes derived third-party pack content, so this research note does not treat that session as independently verifiable evidence. The packet shape remains an observed trigger for investigation, not a general protocol guarantee.

## Routing decision

BedrockDebugProxy now reuses the resolved service signaling network only when the initial upstream is a NetherNet ID and the Transfer target is a UUID with port zero. Go Golem applies the same narrow rule only for resolved `nethernet` or `nethernet-jsonrpc` targets. It passes the parsed UUID alone to the existing network. Other ports, malformed IDs, RakNet targets, and HTTPS endpoint targets retain their prior routing paths.

The change does not alter packet decoding, packet fields, protocol baselines, capture schemas, or stored Transfer bytes. It adds a `transfer.route_selection` lifecycle event that records the selected transport class without copying the destination ID into a second event.

## Limits

The official sources support dialing a destination by NetherNet ID through a signaling network. They do not establish that every UUID address with a zero port has that meaning. The classifier is constrained by the already resolved transport and transfer shape, and its behavior is covered by synthetic tests. Fresh exact-revision public-server validation remains necessary before treating the route as live-compatible. No external code was copied or modified for this change.
