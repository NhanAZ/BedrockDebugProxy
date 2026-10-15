# AUDIT-0216 - Add Golem personalities and interactive control

- Commit subject: `feat: add Golem personalities and interactive control`
- Change type: optional development runtime capability
- Runtime impact: GolemAgent 0.2.0 adds personality and bounded autonomy. Network proxy and capture behavior are unchanged.

## Intent and scope

The maintainer requested distinct personalities, natural basic actions and lightweight extensibility, then explicitly required CLI control of already online Golems without depending on predefined scripts. The existing JSON Lines mailbox already accepts individual requests and keeps players alive across CLI invocations. This extension adds nonblocking movement and observation so an agent can select its next action while the player moves.

Files cover the companion policy/controller, dispatch and state, packaging, focused PHP fixtures, Go integration tests, CLI help and documentation. The controller uses the existing Golem runtime. It adds no dependency, network protocol, capture field, viewer change, resource-pack handling or BedrockPackArchive contract. The previous audit and source review preserve the simulation dependency provenance. FakePlayer remains a read-only architectural reference with no code or asset reuse.

## Design and behavior

Each player owns a seeded personality with pace, curiosity, sociability and boldness. Five profiles and bounded trait overrides affect decision timing, speed, wandering, greetings and damage reactions. Optional autonomous execution has a tick deadline and home radius, probes local terrain and records choices. Combat reactions require an explicitly selected owned Golem. Block placement uses the Player interaction path with the held block item.

Direct commands take control from autonomy. Nonblocking `move` exposes pending, arrived, failed or cancelled state. `stop` cancels movement, and generation checks isolate old callbacks when a destination is replaced. Read-only observations leave autonomous execution running. A CLI EOF leaves players available to another invocation. The [decision](../decisions/0013-golem-development-runtime.md) and [guide](../golem.md) document the trust and simulation boundaries.

## Validation

- The pure PHP personality fixture passes 214 checks, including repeatable seeded choices, trait effects, reaction priority and rejected inputs.
- The isolated server behavior fixture passes 25 checks, including direct takeover, nonblocking moves, replacement callbacks, stopping, chat/sprint/jump during manual movement, social greetings, tick expiry, damage reaction, configured retaliation, default non-combat behavior, ground avoidance and input limits.
- The existing server action suite passes 74 checks with explicit placement now asserting the resulting block. Timeout isolation remains covered.
- The interactive Go-to-PHP test passes. It holds stdin open, creates a player, reads coordinates, chooses and starts movement, observes the online actor, stops, chats, closes input and reattaches through another invocation to observe the same player and clean up.
- The interactive test also passes with the race detector. `tools/format.ps1`, `tools/quality.ps1 -LowMemory` and `git diff --check` pass, including tests, build, module/license checks, viewer regressions, zero static-analysis issues and zero reachable vulnerabilities. The scanner separately reports an imported/required advisory outside called code. Earlier static checks identified unchecked test type assertions and missing explicit JSON field tags. Typed observations and field tags resolve them without lint suppressions.

The disposable server uses the same installed Axolotl source and pre-existing unrelated source changes disclosed in AUDIT-0215. This is synthetic local server evidence, not a clean upstream distribution, Xbox login, public server, network client or retail gameplay validation. Generated journals and PHARs remain outside Git. No push, remote CI or release is requested.

## Limits and human validation

This is a small inspectable behaviour policy. It is not full pathfinding, swimming, climbing, resource-pack processing or human-level AI. Autonomous decisions do not edit blocks or operate arbitrary plugin menus. Explicit CLI commands and trusted tests cover those operations. Seeds reproduce decisions for identical observations and order, not a changing plugin ecosystem. Blocking scenario requests still serialize the mailbox. Interactive callers should use `move` for movement they may need to interrupt.

On a disposable copy of the intended PocketMine or Axolotl plugin server, spawn two profiles and observe their movement and chat, then take control through CLI while one is moving. Exercise the plugin's actual form or block rule and assert the resulting state. Expected behavior is immediate control takeover, no continued old movement and the plugin's usual cancellation semantics. Retain the response, simulation journal, test source and server log on failure. Visual movement quality still needs a human Minecraft observation before making a natural-animation claim.

Together with the preceding Golem commit this is a substantial checkpoint beyond v0.3.0. Before an official release, validate the intended plugin set, the legacy PocketMine branch and the existing release gates. No release is published by this task.
