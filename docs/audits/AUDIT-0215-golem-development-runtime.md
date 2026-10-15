# AUDIT-0215 - Add Golem CLI development automation

- Commit subject: `feat: add Golem CLI for plugin development`
- Change type: optional developer capability
- Runtime impact: new CLI and optional server companion. Existing network proxy behavior is unchanged.

## Intent and scope

The maintainer requested a CLI-oriented Golem facility for developers and AI agents, then explicitly expanded the first version to all plugin-testing action families. The implementation exposes 55 player actions and 10 session actions using a pinned Golem runtime, plus trusted PHP scripts for full server APIs and custom plugin logic. It supports multiple simulated players, typed JSON Lines requests, sequential scenarios, state assertions, tick waits, event subscriptions, cancellation and retained simulation evidence. It does not add a dashboard.

Files are limited to the Go CLI dispatch and mailbox client, their tests, `dev/golem-agent`, user documentation, provenance and this audit. No protocol, networking bridge, capture schema, viewer, resource-pack archive or BedrockPackArchive importer field changes are made. The private consumer checkout was inspected and remains unchanged. Its synchronization rule does not apply to an internal simulated completion reply that produces no real pack or capture data.

## Decision and provenance

[ADR-0013](../decisions/0013-golem-development-runtime.md) evaluates the requested scope against the existing capture tool and a network-client alternative. [The source review](../research/golem-development-runtime.md) records exact Golem, FakePlayer, Specter and Axolotl sources and licensing. The maintainer clarified that FakePlayer has no LICENSE and authorized read-only architectural study without source or asset reuse. The review maps its files, lifecycle, packet observers, tick behaviour and movement into independent design principles. Its current metadata declares API 5.35.0 despite the stale API 4 README.

The runtime dependency is Golem 0.7.0 at `18c86ea25b1d67fa9ceb6079c46ade2f8f99fc31`. The MIT patch and notice are preserved. A source hash lock checks the runtime before the companion enables. The first Axolotl spawn reproduced its removed STATUS_COMPLETED constant. The focused patch chooses the current enum where present and retains the legacy branch. No dependency tree or source from FakePlayer or Specter is imported. No upstream publication is authorized or performed.

## Validation

- Go mailbox tests cover complete atomic requests, response identity, cancellation without retry, retained requests, explicit size limits and missing deadlines. Focused race tests pass. Mailbox access uses `os.Root` to remain inside the selected data directory, including symbolic links.
- CLI tests cover help, malformed JSON, unknown fields, conflicting flags and invalid timeout. Optional PHP tests lint all companion and fixture sources and build a temporary PHAR. PHP is available for this local run.
- `tools/format.ps1` and `tools/quality.ps1 -LowMemory` pass, including tests, build, viewer regression checks, module/license checks and static analysis with zero issues. Earlier gates found capitalized Go errors and environment-derived path traversal warnings. These were resolved with lowercase messages and rooted filesystem access, without suppressions. The vulnerability scan finds zero reachable vulnerabilities and reports an additional imported/required advisory outside called code.
- A packaged GolemAgent with source-lock verification runs on an isolated Axolotl-PM 5.49.2+dev data directory using the installed server source based on `60a13906a193aa17e7a7582be26ffabde7a4007f` and BedrockProtocol `7ff605d5b071d46c694d2b71bc8af4514156dbc2`. The server checkout has pre-existing resource-pack changes. They are not modified or represented as a clean upstream server build.
- The source-generated `dev/golem-agent/tests/smoke.php` passes 74 checks. These include two-player lifecycle, chat and command behavior, operator and permission changes, movement, block cancellation and successful breaking/placement, inventory transactions, forms, actual combat damage, health/death/respawn, HUD contents, custom cancelled events, six standard event families, negative input and cleanup. The fixture uses a disposable flat-world area and synthetic player/item/message data only.
- A 100 ms request timeout is followed by a 40-tick request. The next request takes approximately 2.1 seconds and receives its own result, demonstrating that the expired request's deferred completion cannot finish the new request. The companion ends with no owned players, temporary operators, or saved fake-player data. The opt-in `TestDevelopmentServerIntegration` repeats the 74-check fixture and timeout isolation through the Go client and passes. The documented basic JSONL scenario returns eight successful responses.
- No real network capture is used as evidence for this feature. No public server, Xbox login, retail client, release, push or remote CI validation is claimed. PHP-less CI skips the optional PHP check explicitly, while Go tests still run.

## Limits and human validation

This implementation exposes the full reviewed runtime action families and arbitrary trusted server-API tests. It cannot promise every future plugin, event, item mechanic or Minecraft behavior. Simulated sessions skip pre-login and Xbox authentication. Physics is limited. Packet views are synthetic encodings and public-field projections, not observed network traffic. Event observers must be subscribed before the action and custom private fields require a focused test script. Trusted scripts are administrative PHP code and cannot be forcibly preempted inside the server thread.

Human validation can use a disposable copy of the maintainer's actual plugin server. Install the pinned two-PHAR pair, run the basic scenario, then test the plugin's main command/menu with a Golem and assert its expected state change. If it fails, retain the failed JSON response, boot journal, exact test source and server log. A real Minecraft session remains necessary only for visual, audio or client/network-specific claims.

This is a coherent capability checkpoint beyond v0.3.0. A later release should validate the target plugin set and legacy PocketMine branch in addition to the project's existing release gates. No release is requested by this task.
