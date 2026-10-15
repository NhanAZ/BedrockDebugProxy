# Golem CLI for plugin development

Golem lets a developer or AI agent drive simulated players on a running PocketMine-MP 5 or compatible development fork without opening Minecraft. The CLI accepts JSON Lines, executes scenarios sequentially, and returns structured results. The optional GolemAgent server plugin uses the reviewed Golem 0.7.0 runtime. No dashboard is required.

This is an administrative development tool. Use a disposable server and world. The companion can change player permissions, inventories, blocks and plugin state. Anyone who can write its data directory can control these simulated players and run trusted PHP tests with plugin-level authority. Keep that directory private. Normal proxy commands do not enable this facility.

## Installation

Use your server's PHP binary with the required PocketMine extensions. Build the two optional plugins from a clean pinned upstream checkout. Run these commands from the BedrockDebugProxy repository in PowerShell. The output paths must not already exist.

```powershell
New-Item -ItemType Directory -Force .local/golem-build | Out-Null
git clone https://github.com/achedon12/golem.git .local/golem-build/runtime
git -C .local/golem-build/runtime checkout --detach 18c86ea25b1d67fa9ceb6079c46ade2f8f99fc31
$patch = (Resolve-Path dev/golem-agent/golem-0.7.0.patch).Path
git -C .local/golem-build/runtime apply --check $patch
git -C .local/golem-build/runtime apply $patch
php -d phar.readonly=0 .local/golem-build/runtime/.github/scripts/build-phar.php .local/golem-build/Golem.phar
php -d phar.readonly=0 dev/golem-agent/build.php .local/golem-build/GolemAgent.phar
```

Copy both PHAR files into the disposable server's plugins directory and start it. GolemAgent verifies the runtime source against `runtime-lock.json`. A mismatch disables the companion with an error. The patch supports the current Axolotl resource-pack response enum and the legacy PocketMine constant. It affects only simulation, not network pack negotiation.

Find GolemAgent's data directory. With current server defaults it is `SERVER_DATA/plugin_data/GolemAgent`. Servers configured for legacy plugin data may use `PLUGINS/GolemAgent`. Successful startup creates `ready.json`, `requests`, `responses`, and `scripts` there. Use that directory for `--data`.

```powershell
go build -o bin/bedrock-debug-proxy.exe ./cmd/bedrock-debug-proxy
$data = 'C:/dev-server/plugin_data/GolemAgent'
bin/bedrock-debug-proxy.exe golem --data $data --request '{"op":"capabilities"}'
bin/bedrock-debug-proxy.exe golem --data $data --request '{"op":"spawn","args":{"name":"AgentOne"}}'
bin/bedrock-debug-proxy.exe golem --data $data --request '{"op":"chat","player":"AgentOne","args":{"message":"/help"}}'
bin/bedrock-debug-proxy.exe golem --data $data --request '{"op":"messages","player":"AgentOne"}'
```

Player names are case insensitive. Names belonging to online players, saved players or operators are rejected. Players persist across CLI invocations. `reset` disconnects all companion-owned players and removes operator status granted through the factory. A rejected or interrupted spawn also resets the companion-owned players to prevent a half-created session from leaking.

## CLI and JSON contract

```text
bedrock-debug-proxy golem --data DIRECTORY [--request JSON | --input FILE] [--timeout 30s]
```

Input defaults to stdin. `--request` supplies one JSON object. `--input` supplies a JSON Lines scenario. They are mutually exclusive. Blank lines are ignored. A scenario stops on the first failed response. Exit 0 means all submitted requests succeeded, exit 1 means a runtime, transport or assertion failure, and exit 2 means invalid CLI or JSON input. Successful action delivery alone does not establish plugin correctness. Use assertions on effects.

Each request has `op`, optional `id`, optional `player`, and an `args` object. Unknown top-level fields and action arguments fail. `capabilities` returns the exact action and argument catalog from the running server. Numeric booleans and string numbers are not coerced into typed arguments.

Each response contains schema `bedrockdebugproxy.golem.v1`, boot identity, transport request UUID, caller ID, `ok`, server `tick`, and either `result` or `error`. Stdout contains only these JSON responses. CLI diagnostics go to stderr. The optional ID is for caller correlation. Each transport request gets a separate generated UUID even when caller IDs repeat.

The per-request wall-clock timeout defaults to 30 seconds and must be greater than zero and at most 10 minutes. Tick-based waits advance with the server scheduler and do not block the main thread. Requests are claimed once and never retried automatically. Expired, cancelled or previous-boot requests do not execute. An active timeout resets simulated players and cancels its coroutine. Effects already performed in the world or other plugins are not rolled back. A later completion cannot complete another request.

Only one request runs at a time. Multiple CLI processes may enqueue requests, but ordering between those processes is unspecified. Use one scenario for deterministic action ordering. Arbitrary trusted PHP code can block the server thread or perform irreversible actions, so its wall-clock deadline is cooperative. The CLI still exits at its deadline, but it cannot preempt PHP code running inside the server.

## Actions

All player actions require `player`. `capabilities` lists optional argument names. Core families are:

| Family | Actions and arguments |
| --- | --- |
| Lifecycle | `spawn` with `args.name`, `list`, `isOnline`, `isAlive`, `quit` with optional `reason`, `disconnectReason`, `respawn`, `reset` |
| Chat and commands | `chat` with `message`, `command` with `command`, `messages`, `rawMessages`, `lastMessage` |
| Permissions and setup | `op`, `deop`, `grant` with `permission` and optional boolean `value`, `deny` with `permission`, `gamemode` with `mode` |
| Movement | `teleport` and `walkTo` with `x`, `y`, `z`, `walk` with relative `x`, `z`, optional `speed` for walking, `jump`, `sneak`, `sprint`, `look` with `yaw`, `pitch`, `position` |
| Blocks and entities | `breakBlock` and `interactBlock` with `x`, `y`, `z` and optional `face` for interaction, `useItem`, `attack`, `interactEntity` with a golem name or entity ID in `target` |
| Inventories | `give`, `hold` with parser `item` name and optional `count`, `setSlot` with `slot`, `item`, optional `count`, `inventory`, `window`, `waitForWindow` with optional `ticks`, `clickSlot` with `slot`, `closeWindow` |
| Forms | `formData`, `submitForm` with JSON `data`, `clickButton` with button index or label, `closeForm` |
| Player state | `state`, `health`, `hunger` with optional numeric `value`, `kill`, `respawn` |
| Output | `titles`, `subtitles`, `actionBars`, `popups`, `tips`, `toasts`, `scoreboard`, `bossBar`, `sounds`, `packets` with optional packet `class`, `clearInbox` |
| Tests | `assert`, `waitFor`, `wait`, `subscribe`, `events`, `script` |

`chat` uses the player chat path, including slash commands and chat preprocess behavior. `command` directly dispatches a command as the player. Choose `chat` to test typed command events. `give`, `setSlot`, `op`, `teleport`, health and hunger setters are test setup through server APIs. They do not simulate survival acquisition or a network client request. `clickSlot` executes a real server inventory transaction. Block, entity, movement and form operations use the runtime's Player methods and respect applicable plugin cancellation.

`walkTo` uses the runtime's simple collision and gravity model. Its target height is ignored, as documented upstream. It does not implement swimming, ladder climbing, or complete client physics. Entity interaction returns whether the entity handled it. A generic player target may legitimately return false.

## Scenarios and assertions

Save a scenario as `scenario.jsonl` and use `--input scenario.jsonl`. This example runs a plugin's menu command and confirms its observable reply. Adapt the command and expected text to your plugin.

The repository also includes `dev/golem-agent/examples/basic.jsonl`, which uses only built-in server behavior and cleans up its two simulated players.

```jsonl
{"id":"join","op":"spawn","args":{"name":"AgentOne"}}
{"id":"menu","op":"chat","player":"AgentOne","args":{"message":"/menu"}}
{"id":"form","op":"formData","player":"AgentOne"}
{"id":"choose","op":"clickButton","player":"AgentOne","args":{"button":"Confirm"}}
{"id":"check","op":"waitFor","player":"AgentOne","args":{"path":"messages","contains":"Confirmed","ticks":100}}
{"id":"cleanup","op":"reset"}
```

`assert` reads `state` and follows a dot-separated path, then checks `equals` or `contains`. `contains` means an exact array element or a substring of a string. `waitFor` repeats the same assertion each tick, defaulting to 100 ticks. `wait` pauses for a given number of ticks, defaulting to one. Use `clearInbox` deliberately between phases to avoid matching stale messages. If a scenario fails, its remaining steps, including reset, are not executed. Send a separate reset after inspecting evidence or use a trusted script with `finally` cleanup.

## Events and plugin-specific tests

Subscribe before the action under test. Custom autoloadable plugin event classes work too.

```jsonl
{"op":"subscribe","args":{"class":"pocketmine\\event\\block\\BlockBreakEvent"}}
{"op":"events","args":{"after":0}}
```

Subscriptions run at MONITOR priority and include cancelled events. The event response carries a monotonically increasing event sequence, concrete class, tick, cancellation state, and common getters when available. `after` is an exclusive cursor and `next` is the next cursor to use. This is a documented projection, not a dump of every private field. Overlapping parent and child subscriptions may observe the same event more than once. Use exact classes when counting events. A later MONITOR handler can still modify cancellation state, so the snapshot means state at this observer's position in the handler order.

For assertions on arbitrary custom fields, register a listener in a trusted PHP test and inspect the original event object. The same route exposes the complete Golem `player()` API, worlds, entities, inventories, effects, plugin services, scheduling and other server APIs without requiring a new JSON action for each plugin.

Place trusted PHP files in GolemAgent's `scripts` directory. Each returns a closure accepting `Agent $agent` and a params value. It may return JSON-compatible data, a Golem Deferred, or a generator yielding Deferred or PocketMine Promise values. `clock()->wait()` and `clock()->until()` provide cancellable-by-coroutine tick waits. The helper `golem(NAME)` returns the simulation object. Return a focused projection rather than a whole world or plugin object.

```php
<?php
use BedrockDebugProxy\GolemAgent\Agent;
return static function(Agent $agent, array $params): Generator {
    $golem = yield $agent->dispatch(['op' => 'spawn', 'args' => ['name' => 'Probe']]);
    try {
        $golem->chat('/your-plugin-command');
        yield $agent->clock()->wait(5);
        if (!$golem->player()->hasPermission($params['permission'])) {
            throw new RuntimeException('Expected plugin permission');
        }
        return ['passed' => true, 'messages' => $golem->messages()];
    } finally {
        $agent->dispatch(['op' => 'reset']);
    }
};
```

Invoke with `{"op":"script","args":{"file":"test.php","params":{"permission":"example.use"}}}`. Scripts are trusted plugin code, not sandboxed code. Their exact source and SHA-256 are retained in the simulation journal. Do not put credentials into tests. Path traversal and scripts outside the configured scripts directory are rejected.

## Evidence and limits

The companion retains claimed request files, response files and an append-only `journal-BOOT.jsonl`. Journal rows contain schema, boot, ordered sequence, UTC time, monotonic elapsed nanoseconds and server tick. They record runtime identity, explicit limits, requests, results, subscribed event projections and executed test source. They are simulation evidence and cannot be passed to the proxy's capture verifier or pack importer.

The limits are explicit. Requests, responses, and script sources are at most 1 MiB each. Each boot journal is at most 256 MiB. The companion allows 32 owned players, 256 event subscriptions, 16384 packets per player inbox and 4096 event projections in memory. An expired event cursor fails instead of silently returning a partial history. Complete subscribed event projections remain in the journal. Object views reject depth beyond 12 and collections beyond 16384 entries. Limits produce errors rather than successful truncated results. Reaching a runtime journal or inbox bound disables the companion and closes its players.

Inboxes and packet objects remain in runtime memory until `clearInbox`, `reset`, quit cleanup or shutdown. They are not automatically copied into the journal. Read the outputs you need before clearing them. `packets` results include synthetic server-encoded packet bytes in base64 alongside public fields. These bytes were encoded from simulation objects and were not received over a real network. Binary strings in other views are tagged base64. Large packet queries must be filtered or projected in a script.

There is no real Xbox handshake, client rendering, pack transfer, transport capture, or retail gameplay proof. PlayerPreLoginEvent is skipped by the runtime. The existing observation proxy is still the appropriate tool for real network behavior. This implementation does not claim every possible third-party plugin or Minecraft action works. It exposes all reviewed Golem action families and the full trusted server API for additional tests, with observed limits reported explicitly.

## Validation

The normal `tools/quality.ps1` gate tests the Go mailbox and CLI. When PHP is installed it also checks companion syntax and packaging. A PHP-less environment skips that optional companion check explicitly. The full server suite requires an isolated server and is separate from the network proxy gate.

Copy `dev/golem-agent/tests/smoke.php` into the companion's scripts directory, then run:

```powershell
bin/bedrock-debug-proxy.exe golem --data $data --timeout 60s --request '{"op":"script","args":{"file":"smoke.php"}}'
```

This fixture changes a small flat-world area and exercises player actions, event delivery, negative inputs, forms, inventory transactions, HUD observations and cleanup. Use it only in a disposable world. See the commit audit for the exact server revision and verified results. Real network login, public-server checks and retail gameplay are not part of this local simulation feature's evidence.

For a reproducible Go-to-PHP integration check, set `BDP_GOLEM_TEST_DATA` to the absolute data directory of that disposable companion and run `go test -count=1 -run TestDevelopmentServerIntegration -v ./internal/golem`. This copies the synthetic PHP fixture into scripts, runs it, and checks that an expired request cannot complete a later request. Without the environment variable this server-dependent test is skipped explicitly. Unset the variable after testing.
