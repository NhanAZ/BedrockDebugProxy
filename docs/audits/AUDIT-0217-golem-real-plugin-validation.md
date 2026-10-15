# AUDIT-0217 - Validate Golem CLI against development plugins

- Commit subject: `docs: record Golem real-plugin validation`
- Change type: documentation
- Runtime impact: none. This commit records tests of the preceding Golem runtime commits.

## Intent and scope

The maintainer requested end-to-end Golem CLI testing with actual plugins on the local development server, fixes for Golem defects found, and a push to the existing remote. The isolated server run exercised commands, forms, custom events, block ownership, synthetic packets, movement and control across CLI processes. No Golem code defect was demonstrated, so the runtime was not changed. The source plugin repositories and the server's pre-existing source modifications were not edited.

## Environment and provenance

- BedrockDebugProxy runtime revision: `1244f160a27e2e4a09a3d8ffc9afd92378ca413e`. The stamped CLI and companion PHAR hashes remain those in AUDIT-0216.
- Server: Axolotl-PM 5.49.2+dev at `60a13906a193aa17e7a7582be26ffabde7a4007f`, PHP 8.4 on Windows, installed `axolotl-pm/bedrock-protocol` at `7ff605d5b071d46c694d2b71bc8af4514156dbc2`. The server checkout already had unrelated resource-pack source changes and an untracked test directory. Data and plugin PHARs were copied into a disposable, Git-ignored directory. The test server ran on local port 19442 with authentication disabled for this local simulation.
- DevTools 1.0.15: source revision `3dedda2a1f6e42885aa680aaadeebacbc0f62733`, PHAR SHA-256 `a2f4722caa1f61a194dc45a121c44c5bf59baaaf3254876e07f331d3d4dc309d`.
- SimpleEconomy 1.0.0: source revision `d2fcb55142add8f4ba006c7fd9b322ca0e7c2e31`, existing DevTools-built PHAR SHA-256 `cc637dbaabf29079ec09c3fac5e2fcc60ce2673435fd6a746078559374c7bb8e`.
- BlockDataExample 1.0.0: source revision `123edcf68e0a06d3dea15115d4982689348d8539`, existing DevTools-built PHAR SHA-256 `3c3bbaa9d99bd4fddd1f64d14b010b3f3d083c8bda296c00b72d2fa57e1bf030`.
- CustomToastExample 1.0.0: source revision `8be18f4b89bfc37497803c58c3ae2304fa4f2cfa`, existing DevTools-built PHAR SHA-256 `9b3cba1993b49c6c99d2e9c41ed0a9dfbd576622e0f4d69692a85224d3b7386a`.
- FormApiDemo 1.0.0: existing PHAR SHA-256 `509dd278100f9db0872a8095586375eef649200cb7bd6b01799509decc34ea10`. The PHAR source was inspected read-only for its form callbacks and private audit format.
- ClientboundUpdateSoundDataPacket 0.1.2: source revision `4edd47c4991cd0501451394fe106855faf0a34a8`, PHAR SHA-256 `865abf0537e573571e799d7b2cca2e8e144570baa752efdfbbb9dc553d5db964`.

These are local source and artifact identities, not a claim that each PHAR was rebuilt from the listed source revision during this task. Runtime evidence comprises CLI JSON responses, companion journals, the isolated server log and the FormApiDemo audit. All remain outside Git. They are synthetic server-side evidence and are not proxy captures.

## Observed flows

1. The server loaded Golem 0.7.0, GolemAgent 0.2.0 and the plugin set. The CLI `capabilities` request returned the action catalog. Two separate `spawn` requests joined `PluginAgent` and `PluginPeer`. `/devtools status` returned DevTools 1.0.15 and its environment summary to the first player's inbox.
2. Both players acquired SimpleEconomy's default 1,000 balance. `PluginAgent` issued `/pay PluginPeer 125`. The sender received `You sent $125 to PluginPeer.` and the recipient received `You received $125 from PluginAgent.`. A subscription to `NhanAZ\SimpleEconomy\event\TransactionSuccessEvent` observed four plugin-fired events. The later `/money` response for the recipient showed 1,125. The event count is reported as observed, not normalized to a claimed number of transfers.
3. `PluginAgent` held stone and `placeBlock` accepted a placement at `(261, 4, 258)`. A `BlockPlaceEvent` was observed and the inventory count fell from four to three. `PluginPeer` attempted to break that block and received `This block belongs to PluginAgent!`; the observed `BlockBreakEvent` was cancelled. After `/inspect`, interaction returned an owner and timestamp. The owner's break succeeded, produced an uncancelled event and sent `Block data removed.`. An initial attempt to place a block inside the actor's own position returned false without an event, consistent with collision rejection.
4. `/formapi` exposed the simple, modal and custom forms through `formData`. `clickButton` and `submitForm` produced `Button 0 selected.`, `You selected Yes.`, and a custom result containing `[null,"agent input",false,7,2,0]`. FormApiDemo's audit independently recorded the same responses.
5. `move` immediately returned pending state. A later `/money` command ran while the Golem continued moving, and `stop` changed movement state to cancelled. The social profile then ran bounded autonomy and greeted the other Golem. A held-open stdin/stdout CLI process listed the already online player, sent another `/money` command, read its inbox, exited cleanly, and a new process reattached to the same online player.
6. With `play-sound: false` set only in the disposable CustomToastExample config, `/toast ToastPeer plain round green Agent hello` returned `Toast sent to ToastPeer.`. The recipient's `packets` view included a synthetic TextPacket whose message began `%toast%` and contained `Agent hello`. Packet bytes were produced by the server simulation. No client rendering was observed.

## Compatibility failures and classification

- The ClientboundUpdateSoundDataPacket PHAR's delayed `/soundtest repro 2` task called `ClientboundUpdateSoundDataPacket::create()` with `null` as its second argument. The installed BedrockProtocol API requires `SoundDataEvent`. The uncaught plugin task exception crashed the disposable server. The CLI then surfaced companion unavailability; Golem did not generate this packet or choose its arguments. The plugin was excluded from later test boots.
- The CustomToastExample PHAR's default sound path called `PlaySoundPacket::create()` with `null` for its seventh argument. The installed API requires an integer loop count. The CLI surfaced the `TypeError` and the server remained running. Disabling the plugin's optional sound in isolated config allowed the toast TextPacket path to pass. This does not validate the default sound behavior or a Minecraft client's visual toast.
- SimpleEconomy's bundled libasynql emitted PHP 8.4 implicit-nullability deprecation messages during startup, but enabled and completed the tested transfer. No compatibility source fix was made in unrelated plugin repositories.

## Validation and limits

- Direct local server plugin flows: pass for DevTools status, SimpleEconomy transfer and messages, BlockDataExample ownership, FormApiDemo form callbacks, and CustomToastExample's sound-disabled packet path. The sound plugin reproduction and CustomToastExample default sound path fail as classified above.
- Interactive CLI and reattachment against actual plugins: pass. Companion reset and server shutdown completed. The disposable world, plugin data, request files and journals were not committed.
- Canonical `tools/format.ps1`, `tools/quality.ps1 -LowMemory` and `git diff --check`: pass. The quality gate passed tests and build, found zero static-analysis issues and zero reachable vulnerabilities. Its scanner also reported an imported and required advisory outside called code.
- Retail client, Xbox login, resource-pack rendering, public server interoperability and complete third-party plugin compatibility: not tested. The server source had pre-existing unrelated changes, so this is not a clean upstream Axolotl distribution validation. The two sound API failures need compatible plugin builds and separate plugin-owner review. No runtime Golem fix is justified by these findings.
