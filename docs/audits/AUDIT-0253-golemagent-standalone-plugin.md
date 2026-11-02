# AUDIT-0253 - Move GolemAgent to a standalone plugin repository

- Commit subject: `Move GolemAgent to a standalone plugin repository`
- Change type: `tooling`
- Runtime impact: `no`

## Intent

Separate the optional PHP PocketMine/Axolotl GolemAgent companion from
BedrockDebugProxy before replacing the BedrockDebugProxy Golem command with a
Go network-client workflow. The plugin remains available for local server-side
simulation tests, while the proxy repository no longer owns its PHP sources.

## Scope

- Published `NhanAZ-Plugins/GolemAgent` from the `dev/golem-agent` subtree,
  preserving its two historical Golem commits and adding independent README,
  GPL license, ignore rules, and packaging metadata.
- Fixed the plugin build script to package its local `LICENSE` file and updated
  its website and description.
- Removed the PHP companion source from BedrockDebugProxy and updated the
  local Golem guide and third-party notice to use the standalone repository.
- Kept the existing Go mailbox bridge and `bedrockdebugproxy.golem.v1`
  contract unchanged as a transitional compatibility layer. The Go network
  Golem replacement is a separate follow-up.

## Evidence and provenance

- Standalone repository: https://github.com/NhanAZ-Plugins/GolemAgent
- Published commit: `74c30f49ea34169320016cb604ba2745a8aa515e`.
- The split history contains the original Golem CLI companion commits
  `5bced92` and `d08aef0` before the standalone packaging commit.
- The plugin source retains the reviewed Golem runtime lock, patch, and MIT
  notice. No credentials, captures, or server data were moved.

## Validation

- Automated checks: PHP lint passed for all plugin PHP files; the personality
  fixture passed 214 checks; the behavior fixture passed; the standalone PHAR
  build passed; and the PHAR contained `GOLEM-LICENSE.txt`, `LICENSE`,
  `plugin.yml`, `runtime-lock.json`, and `src`.
- Live validation: `not_observed`. No disposable server was started by this
  repository split.
- Capture or report: `none`.

## Limitations and follow-up

The current BedrockDebugProxy Go `golem` command still controls this plugin's
mailbox for compatibility. It is not yet the desired Go Bedrock network bot.
Implement and validate that network-client Golem before removing the bridge or
changing the release workflow to use it.
