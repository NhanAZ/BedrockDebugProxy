# Golem source review

Reviewed on 2026-10-15. This is a new development workflow. No packet definitions or wire behavior are added.

## Sources

- [achedon12/golem](https://github.com/achedon12/golem/tree/18c86ea25b1d67fa9ceb6079c46ade2f8f99fc31), revision `18c86ea25b1d67fa9ceb6079c46ade2f8f99fc31`, version 0.7.0. The [MIT license](https://github.com/achedon12/golem/blob/18c86ea25b1d67fa9ceb6079c46ade2f8f99fc31/LICENSE), [Golem API](https://github.com/achedon12/golem/blob/18c86ea25b1d67fa9ceb6079c46ade2f8f99fc31/src/Golem.php), [factory](https://github.com/achedon12/golem/blob/18c86ea25b1d67fa9ceb6079c46ade2f8f99fc31/src/Runtime/GolemFactory.php), [session](https://github.com/achedon12/golem/blob/18c86ea25b1d67fa9ceb6079c46ade2f8f99fc31/src/Runtime/Network/GolemSession.php), and [limitations](https://github.com/achedon12/golem/blob/18c86ea25b1d67fa9ceb6079c46ade2f8f99fc31/docs/how-it-works.md) were read at that revision. Runtime use is a dependency, not independently implemented simulation. Its license stays in its separately built PHAR.
- [Muqsit/FakePlayer](https://github.com/Muqsit/FakePlayer/tree/5f7b84626bb61b17ef49281f34c2808162dd77b1), revision `5f7b84626bb61b17ef49281f34c2808162dd77b1`. Its README still describes API 4.0.0, but the reviewed [plugin metadata](https://github.com/Muqsit/FakePlayer/blob/5f7b84626bb61b17ef49281f34c2808162dd77b1/plugin.yml) declares API 5.35.0 and version 0.3.0. The reviewed tree has no license file. The maintainer clarified that this is a read-only technical reference and withdrew the earlier MIT assertion. No permission to reuse its source or assets is assumed. No source or assets are imported, adapted or redistributed, and the original repository is unchanged.
- [falkirks/Specter](https://github.com/falkirks/Specter/tree/7903e92c5028f93f56ea42aedb2c42998aa86a6a), revision `7903e92c5028f93f56ea42aedb2c42998aa86a6a`. Its [MIT license](https://github.com/falkirks/Specter/blob/7903e92c5028f93f56ea42aedb2c42998aa86a6a/LICENSE), DummyPlayer API, SpecterInterface and plugin metadata were read. It exposes a player through a fake interface, feeds simulated replies, and supports chat, teleport, attack and respawn. Metadata targets PocketMine API 3 and protocols 407-428. It supports the server-side architecture choice but is not a current-protocol runtime. No code is copied.

## FakePlayer structure and execution model

The recursive Git tree and the following source files were read at the pinned revision. This section is an explanation of technical principles, not a source translation or a reimplementation of its protected expression. No FakePlayer runtime was executed during this review.

| Files | Responsibility |
| --- | --- |
| `plugin.yml`, `resources/config.yml` | Plugin entry point, API requirement, operator command permission and delayed configured-player creation |
| `info/FakePlayerInfo.php`, `info/FakePlayerInfoBuilder.php` | Identity and fixture configuration, including UUID, username, skin, extra client metadata and behaviour configuration |
| `Loader.php` | Session construction, asynchronous Player creation, owned-player registry, lifecycle listeners and per-tick scheduling |
| `network/FakePlayerNetworkSession.php`, `FakePacketSender.php` | In-process session, outbound packet observation and a transport sink that performs no network sending |
| `DefaultFakePlayerListener.php` | Simulated initialization, respawn and dimension acknowledgement after server output |
| `FakePlayer.php`, `behaviour/FakePlayerBehaviourFactory.php`, `util/SortedMap.php` | Player wrapper, registered behaviour construction, priority ordering and tick dispatch |
| `behaviour/internal/*Movement*.php` | Shared motion state and movement processing around custom behaviours |
| `behaviour/PvPFakePlayerBehaviour.php` | Periodic target selection, motion, facing and attacks |
| `FakePlayerCommandExecutor.php` | Operator control for chat, forms, interaction and teleporting fake players |

Paths in this table are relative to `src/muqsit/fakeplayer` except the root metadata and resources. The resources directory also contains configured identities and a skin asset. Those assets are not reused.

### Creation and lifecycle

[Loader.php](https://github.com/Muqsit/FakePlayer/blob/5f7b84626bb61b17ef49281f34c2808162dd77b1/src/muqsit/fakeplayer/Loader.php) builds a specialized session with the server's packet, entity and conversion services, registers it with the session manager, installs synthetic identity information through reflection, and invokes the internal successful-login transition. A simulated resource-pack completion response advances Player creation. This is an internal test shortcut, not Xbox authentication or evidence of resource-pack downloading.

[FakePlayerNetworkSession.php](https://github.com/Muqsit/FakePlayer/blob/5f7b84626bb61b17ef49281f34c2808162dd77b1/src/muqsit/fakeplayer/network/FakePlayerNetworkSession.php) delegates actual Player construction to the server. The loader registers the wrapper and behaviours before resolving the externally returned promise. This ordering ensures completion callbacks can find the owned fake player. Promise completion at Player creation should not be confused with full client spawn readiness. Quit handling removes the wrapper and behaviours without attempting another disconnect.

Configured players are normally added after a delay. The [configuration](https://github.com/Muqsit/FakePlayer/blob/5f7b84626bb61b17ef49281f34c2808162dd77b1/resources/config.yml) explains why creating players while plugins are still enabling can cause other plugins to miss login events. The transferable principle is to make runtime readiness and plugin initialization order explicit.

### Output, forms and simulated client replies

The session observes outbound send-buffer entries by decoding them and dispatching general or packet-class listeners. [FakePacketSender.php](https://github.com/Muqsit/FakePlayer/blob/5f7b84626bb61b17ef49281f34c2808162dd77b1/src/muqsit/fakeplayer/FakePacketSender.php) provides an inert transport. These are server-produced packet objects, not packets observed on a real socket.

[DefaultFakePlayerListener.php](https://github.com/Muqsit/FakePlayer/blob/5f7b84626bb61b17ef49281f34c2808162dd77b1/src/muqsit/fakeplayer/DefaultFakePlayerListener.php) reacts to spawn status, respawn and dimension changes with delayed internal replies or Player calls. Its fixed 40-tick delays are implementation choices at this revision, not general protocol requirements.

[FakePlayerCommandExecutor.php](https://github.com/Muqsit/FakePlayer/blob/5f7b84626bb61b17ef49281f34c2808162dd77b1/src/muqsit/fakeplayer/FakePlayerCommandExecutor.php) invokes Player chat while temporarily forwarding text output to the operator. Form submission obtains the latest form identifier from private Player state and feeds a form response into the session handler. Interaction invokes block interaction or held-item use. A temporary chat observer can miss replies emitted after the command returns. The independent bridge retains observations across requests instead.

### Behaviour and movement

[FakePlayer.php](https://github.com/Muqsit/FakePlayer/blob/5f7b84626bb61b17ef49281f34c2808162dd77b1/src/muqsit/fakeplayer/FakePlayer.php) executes behaviours in ascending priority. The loader places motion preparation before ordinary behaviours and movement application after them. The shared motion record keeps server-side velocity between ticks. The [preparation stage](https://github.com/Muqsit/FakePlayer/blob/5f7b84626bb61b17ef49281f34c2808162dd77b1/src/muqsit/fakeplayer/behaviour/internal/TryChangeMovementInternalFakePlayerBehaviour.php) observes server motion output and invokes internal movement processing. The [application stage](https://github.com/Muqsit/FakePlayer/blob/5f7b84626bb61b17ef49281f34c2808162dd77b1/src/muqsit/fakeplayer/behaviour/internal/UpdateMovementInternalFakePlayerBehaviour.php) computes a candidate location, then passes it through Player movement handling. This allows server rules and plugin movement handling to participate. It is not a full retail-client physics engine.

The [PvP behaviour](https://github.com/Muqsit/FakePlayer/blob/5f7b84626bb61b17ef49281f34c2808162dd77b1/src/muqsit/fakeplayer/behaviour/PvPFakePlayerBehaviour.php) periodically examines nearby eligible entities, compares squared distances, caches a target, adjusts motion and facing, and invokes the Player attack API within reach. Random early selection and movement offsets mean this is a heuristic, not guaranteed nearest-target selection or pathfinding. Deterministic plugin tests should supply explicit targets and controlled movement instead.

### Independent application to BedrockDebugProxy

The useful principles are a real server Player behind a simulation boundary, asynchronous readiness, tick-based actions, observable outputs, lifecycle cleanup and assertions on effects. The bridge implements its own CLI, mailbox, action dispatch, event projection, cancellation identities and evidence journal from those requirements. The simulation engine remains the explicitly declared MIT Golem dependency. Independent of FakePlayer does not mean the entire engine is written from scratch.

The chosen flow is CLI JSON request to a private local mailbox, then main-thread dispatch, then the Golem/Player API, plugin handlers and world state, followed by a structured response and journal. A scenario can spawn two players, subscribe to the event under test, invoke a command or interaction, wait for its effect, assert the resulting state and clean up. Custom PHP tests cover plugin-specific APIs and events. No copied identity fixtures, skin, movement implementation, private-state access code or command parser from FakePlayer is needed.

Reflection-heavy session construction is version sensitive. Keep it inside the pinned simulation dependency and verify the installed server with lifecycle, cancellation, forms, inventory, combat and movement tests. Do not expose simulated login as a validated network login, reuse fixed delays as protocol rules, or claim that arbitrary server API calls exercise every client event.

## Findings

All three references simulate players inside PocketMine rather than connecting a real network client. Golem's factory and Player methods provide the requested action families. The fake session skips pre-login and Xbox authentication. Movement implements a limited client simulation. Its public `waitForWindow` currently depends on the test runner singleton, so the companion uses its own Clock condition around the supported `window` and session tick APIs. No upstream contribution is submitted.

## Demonstrated Axolotl integration fix

The first local spawn failed with `Undefined constant ResourcePackClientResponsePacket::STATUS_COMPLETED`. Axolotl-PM source at `60a13906a193aa17e7a7582be26ffabde7a4007f` uses BedrockProtocol `7ff605d5b071d46c694d2b71bc8af4514156dbc2`. Its [ResourcePackClientResponseStatus](https://github.com/axolotl-pm/BedrockProtocol/blob/7ff605d5b071d46c694d2b71bc8af4514156dbc2/src/types/ResourcePackClientResponseStatus.php) defines COMPLETED as an enum. The exact source was read and supports this claim. The dependency patch selects that enum when present and retains the legacy constant otherwise. The subsequent local spawn and action suite pass. This is a local dependency integration incompatibility, not evidence of a wire protocol defect.

The patch changes only the fake session's internal completion reply. It does not handle real pack offers, archives, capture fields, or the producer/consumer contract. BedrockPackArchive is unaffected. The runtime lock verifies the reviewed dependency source hashes, including the patched file, before the companion enables. The patch context and MIT notice are retained in dev/golem-agent. No library tree is copied into the Go source.

The bridge runs actions on the server main thread. Golem's Deferred and Coroutine types allow tick-based waits without blocking that thread. Player data saves for simulated players are cancelled, names colliding with real/offline players or operators are rejected, and the factory restores temporary operator grants on reset.

The source review does not prove every third-party plugin works. Local integration tests must exercise the actual installed fork and report unsupported behavior explicitly. Custom events and API assertions are available through trusted PHP scripts instead of claiming a finite generic command list covers every future plugin.

## Personality and interactive control follow-up

The controller uses the pinned Golem public `movement()`, `walkTo()`, `jump()`, Player actions and Deferred callbacks. The runtime's `Movement::abort()` cancels an active movement. A controller generation number rejects stale completion callbacks after an agent changes destination or takes control. No simulation dependency patch was added for personalities.

PHP's built-in [Random Engine Xoshiro256StarStar](https://www.php.net/manual/en/class.random-engine-xoshiro256starstar.php) accepts a seed and is available in the installed PHP 8.4 runtime. Its official manual and local availability were checked on 2026-10-15. It supplies independent per-player choices without a new library. The decision policy and terrain probes are independently authored project code. They do not reuse FakePlayer's movement implementation, randomized PvP heuristic or assets.

Environment sensing uses owned-player positions, loaded local blocks and the server's damage events. Decisions prefer damage reaction over blocked-route handling, social behaviour and wandering. This fixed priority and explicit inputs are intentionally inspectable. Terrain probing avoids unsupported ground and obstruction but does not provide a pathfinding guarantee. Damage is observed at MONITOR priority, so another later MONITOR listener could still alter the final event. Full world determinism and retail-client parity remain unproven.
