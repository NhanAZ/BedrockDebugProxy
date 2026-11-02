# AUDIT-0256 - Produce canonical captures and release reports from Go Golem

## Scope

This change connects `golem network` to the existing BedrockDebugProxy capture,
resource-pack, automatic-artifact, analysis, and validation-report contracts.
It does not open Minecraft for Windows and does not change the PHP GolemAgent
mailbox contract.

## Implementation

- Added evidence callbacks to the Go network client for connection, packet,
  resource-pack, and spawn observations.
- Added a direct-client capture adapter that records post-transport packet
  payloads and decoded latest-protocol views in both directions, connection and
  spawn lifecycle events, bounded GameData, and transfer-hop identity.
- Reused `bedrock.RecordResourcePacks` for encrypted archive preservation and
  ContentKey-scoped derived decryption. No offline key lookup, guessing,
  arbitrary retrieval, or fallback encryption path was added.
- Added `--capture`, `--decrypt-resource-packs`, `--validation-report`,
  `--validation-result`, `--server-label`, and repeatable `--validation-check`
  options to `golem network`.
- Automatic release checks are derived only from observed evidence. The CLI
  marks the session stability check as passing only after a successful scenario
  `stability` window totaling at least 30 seconds with client input packets.
  Server-specific review checks, including Enchanted route review, remain
  explicit inputs.
- Kept full gameplay claims out of the report. The current client provides
  scenario-driven forms, chat, Transfer following, and stationary
  `PlayerAuthInput`; movement, inventory, combat, block interaction, and
  rendering are not claimed.

## Security review

Server packets, form JSON, transfer destinations, and scenario input remain
untrusted. The implementation keeps bounded structured views, uses the existing
address validation and transfer-hop limit, copies packet payloads before
retention, avoids shell interpolation by passing validation-script arguments as
separate process arguments, and does not expose authentication tokens in
reports. Capture and derived pack files remain operator-owned local artifacts.

## Paired consumer check

The downstream compatibility fixture was updated and published separately at
`NhanAZ-Tools/BedrockPackArchive` commit
`a270f30edda03e4552b158b444f045c9b7870169`. It now carries synthetic
`operation_mode=golem_network`, `direct_client=true`, and automatic-artifact
metadata while preserving archive, decrypted archive, contents manifest,
plaintext `contents.json`, key-count, and index assertions. The importer
production path is unchanged because the capture and archive event contract is
unchanged.

## Verification record

Focused checks completed before this producer commit:

- `go test -count=1 ./internal/golemnet ./cmd/bedrock-debug-proxy`
- `go test ./cmd/bedrock-debug-proxy -run TestGolemCaptureWritesCanonicalEvidence -count=1`
- Paired consumer `go test -count=1 ./...`
- Paired consumer `go vet ./...`
- Paired consumer `go build -o bin/bedrock-pack-archive.exe .`
- Paired consumer `git diff --check`

The producer commit must additionally pass `tools/quality.ps1`, race tests,
stamped build, and the paired actual-producer fixture with
`BPA_TEST_BEDROCK_DEBUG_PROXY` set to that stamped binary before publication.

## Live validation plan

Use the stamped producer binary and an approved cached account. For The Hive,
exercise `golem network --server experience:The Hive --auth device --capture
... --decrypt-resource-packs --validation-report ...` with a scenario that
waits for spawn, runs `stability` for at least 30 seconds, observes the final
state, and quits. Confirm the closed capture has bidirectional raw and decoded
traffic, successful pack derivation for every encrypted offer, no decode or
blocking errors, and a complete automatic-artifact status. Repeat the same
protocol-client flow for Galaxite, Lifeboat, Mineville Zeqa, Enchanted, and
CubeCraft under the existing release matrix. For Enchanted, inventory the
current Hub form, follow observed nested forms and Transfers, and supply the
review checks only for the routes actually observed. Retain failed captures for
diagnosis and do not use a retail client.
