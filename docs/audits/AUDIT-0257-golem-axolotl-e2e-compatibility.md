# AUDIT-0257 - Run Go Golem against an Axolotl-PM disposable server

- Commit subject: `Make Go Golem local E2E compatible with Axolotl-PM startup tables`
- Change type: local E2E compatibility and diagnostic reliability
- Parent decision: [0017 Go network-client Golem](../decisions/0017-go-network-client-golem.md)

## Intent and scope

The maintainer requested a disposable PocketMine-MP server and sample JSONL
scenarios driven by the Go Golem CLI, without opening Minecraft for Windows.
The first fixture used official PocketMine-MP 5.44.3, which rejected the
current protocol 2225 client before login. The fixture was then moved to
Axolotl-PM 5.51.1, whose Bedrock v26.60 startup flow matches protocol 2225 and
loads both Golem 0.7.0 and GolemAgent 0.2.0.

This audit records the local interoperability corrections needed to complete
that flow. It does not claim public-server compatibility or full gameplay.

## Reproduced failures

- The official PocketMine-MP 5.44.3 fixture rejected the Go client with
  protocol 2225 as an incompatible version.
- Axolotl-PM accepted a valid offline username but sent an `ItemRegistry` with
  2,081 entries, a `CreativeContent` table with 1,722 items, and a
  `CraftingData` vector with 1,425 recipes. The generic gophertunnel
  1,024-entry safety bound reported decode errors.
- The Go Golem recorded every generic packet as a scenario notification. The
  startup burst filled the 1,024-entry notification queue before a scenario
  could consume the next actionable event.
- The validation wrapper passed one `-Check` parameter per check. PowerShell
  rejected the repeated named parameter even after the session itself had
  completed.

## Implementation

- Added bounded 16,384-entry readers for `ItemRegistry`, both
  `CreativeContent` tables, and all `CraftingData` recipe vectors. The generic
  protocol slice bound remains 1,024.
- Kept generic packet telemetry in the bounded diagnostic history and sent
  only actionable lifecycle, text, form, transfer, and disconnect events to
  the scenario queue. History rolls at its cap instead of terminating a live
  session.
- Changed the default offline Golem name from `BDP-Golem` to `BDPGolem` so a
  direct offline session satisfies Bedrock username validation.
- Passed validation checks as one comma-delimited array argument and expanded
  it in the PowerShell report script.

## Sample disposable scenarios

The local ignored fixture directory contains these JSONL scenarios:

- `basic-stability.jsonl` sends chat, waits for server traffic, holds a
  30-second stationary-input window, writes a report, and quits.
- `invalid-input.jsonl` exercises deterministic unsupported-operation failure.

The fixture server is configured with `xbox-auth=off` and the Go command uses
`--auth none`. Its report must therefore keep `authenticated_upstream` as
`not_observed` and remain supplemental evidence.

## Verification

Focused tests passed for `internal/golemnet`, `internal/bedrock`, and
`cmd/bedrock-debug-proxy`. The Axolotl-PM fixture loaded both PHP plugins, the
Go Golem reached spawn, sent client input for the full 30-second scenario
window, and closed cleanly after the startup-table and queue fixes. The exact
stamped binary run, canonical capture manifest, supplemental report, and
validation report are recorded after the clean follow-up commit and build.

No public server was contacted and no retail Minecraft client was opened.
