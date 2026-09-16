# Protocol update workflow

## Purpose

Minecraft Bedrock protocol updates are evidence changes, not dependency-refresh chores. A successful compile only proves that local APIs still type-check. It does not prove packet IDs, field order, optional fields, version gates, login, encryption, compression, resource-pack transfer, spawn, or compatibility with different server implementations.

Use this workflow for a gophertunnel upgrade, a new Bedrock release, a packet definition change, or a regression that first appears after either one.

## Current baseline

The baseline introduced by the Bedrock 1.26.50 protocol update is:

| Component | Pinned value | Evidence |
| --- | --- | --- |
| Go toolchain | `go1.26.6` | `go.mod` toolchain directive |
| gophertunnel | `v1.61.0` plus reviewed upstream 1.26.50 changes and local compatibility patches | `go.mod`, `third_party/gophertunnel/BEDROCKDEBUGPROXY_PATCH.md`, and `docs/research/protocol-1.26.50.md` |
| gophertunnel source | `283a5a97dfe65da94bcc0b401807f6aefa9e72ee`, `481f3bd138766304a73d7a0412a47a87acec15ed`, and `b8bd7357c24fcba3e32f940d1afb1d3b4e43b96d` | Stable base, 1.26.50 feature branch, and required-field correction |
| Bedrock protocol | `2193` | Mojang `bedrock-protocol-docs` release `v1.26.50` and `minecraft/protocol/info.go` |
| Bedrock game version | `1.26.50` | Mojang `bedrock-protocol-docs` release `v1.26.50` and `minecraft/protocol/info.go` |
| go-raknet | `v1.15.2-0.20260705184311-0d1fd09e2cf6` | `go.mod` |
| Capture schema | `bedrockdebugproxy.capture.v1` | `internal/capture/schema.go` |

Run `bedrock-debug-proxy version` to report the protocol and game version compiled into a binary. Record the binary output, commit, target Minecraft version, and server software with every live capture used as update evidence.

## Required evidence

Do not add or change a packet ID, field, type, serialization order, version gate, required sequence, or compatibility workaround from intuition. For each material change, record:

- the exact claim being implemented
- source URLs and immutable revisions or release versions
- the protocol version covered by each source
- whether the evidence is a definition, executable implementation, test, packet capture, or observed server behavior
- disagreements and the reason for choosing one interpretation
- remaining unknowns and how raw capture preserves them

Prefer authoritative runtime evidence and primary source definitions. Cross-check at least two independent sources when practical. Use the catalog below to select sources with different origins instead of counting forks or generated outputs as independent agreement.

Source agreement is not enough when all sources may share an outdated assumption. A live capture can confirm bytes and ordering for one server and version, but it does not prove that every valid server must behave identically.

## Protocol reference catalog

Record the exact revision, release, file path, protocol version, access date when useful, and license for every source used to support a wire claim. The repository name alone is not evidence.

| Source | Primary role in a protocol review | Review caution |
| --- | --- | --- |
| [Mojang bedrock-protocol-docs](https://github.com/Mojang/bedrock-protocol-docs) | Official released packet schemas, IDs, fields, and version metadata | Match the release tag to the target game version. The files are specifications under Mojang's stated terms, not project source to copy. |
| [Sandertv gophertunnel](https://github.com/Sandertv/gophertunnel) | Selected Go session and protocol implementation | Inspect both stable and relevant development branches. Preserve documented local patches when importing changes. |
| [Sandertv go-raknet](https://github.com/Sandertv/go-raknet) | Selected RakNet transport implementation | RakNet behavior does not prove Bedrock packet layout. Preserve the local GUID patch. |
| [Cloudburst Protocol](https://github.com/CloudburstMC/Protocol) | Independent Java packet codecs and protocol metadata | Generated metadata may lag an official release. Resolve version disagreements against primary evidence. |
| [PrismarineJS bedrock-protocol](https://github.com/PrismarineJS/bedrock-protocol) and [minecraft-data](https://github.com/PrismarineJS/minecraft-data) | JavaScript codecs, relay behavior, and multi-version data | Protocol definitions may be split across repositories and may lag the newest release. |
| [Endstone spyglass](https://github.com/EndstoneMC/spyglass) | Windows Bedrock client packet decode diagnostics | Diagnostic output is observed evidence for the exercised flow, not a complete definition. |
| [Endstone endstone](https://github.com/EndstoneMC/endstone) | Bedrock Dedicated Server integration and runtime behavior | BDS call sites show implementation behavior but may not expose raw wire details. |
| [Endstone endweave](https://github.com/EndstoneMC/endweave) | Cross-version translation and explicit version gates | Translation code can contain compatibility policy rather than a canonical layout. |
| [axolotl-pm PocketMine-MP](https://github.com/axolotl-pm/PocketMine-MP) | PHP server call sites and packet handling behavior | Pair it with its protocol library at a matching revision. |
| [axolotl-pm BedrockProtocol](https://github.com/axolotl-pm/BedrockProtocol) | PHP packet serializers and protocol structures | Confirm its advertised Minecraft and protocol version before using a definition. |
| [Altay](https://github.com/altayofficial/Altay) | Alternate PocketMine-derived server behavior | Establish fork lineage so inherited code is not counted as independent confirmation. |
| [BetterAltay](https://github.com/BetterAltayBedrock/BetterAltay) | Additional PocketMine-derived server behavior | Establish fork lineage and local modifications before treating it as separate evidence. |
| [Endstone protocol-dumper](https://github.com/EndstoneMC/protocol-dumper) | Extracted Bedrock Dedicated Server protocol metadata | Record the exact BDS build used to generate the output. |
| Owner-authorized captures | Actual ordered bytes, timing, directions, and server-specific behavior | A capture proves only the exercised version, server, route, and flow. Retain raw evidence. |

## Update procedure

### 1. Establish a clean baseline

Start from a clean working tree and a passing commit. Run:

```powershell
.\tools\quality.ps1
go run .\cmd\bedrock-debug-proxy version
```

Preserve at least one verified capture from the current build for comparison. Record which parts have only automated coverage and which have live client or server validation.

### 2. Define the smallest update scope

Identify whether the required change belongs in the dependency version, the Bedrock adapter, the capture model, or documentation. Do not combine a protocol bump with unrelated refactoring, capture schema changes, new features, or cleanup.

Before adding local protocol code, search the installed dependency and existing adapters for the same functionality. Prefer a supported upstream API or a narrow adapter over a second packet implementation.

### 3. Build an evidence table

Create or update a note under `docs/research/` before coding. One row per claim is sufficient.

| Claim | Protocol version | Source and revision | Independent check | Decision | Confidence |
| --- | --- | --- | --- | --- | --- |
| Example packet field is optional after field X | Exact ID | Definition link | Capture or second implementation | Gate or retain unknown bytes | Confirmed, inferred, or unknown |

If sources disagree, stop and classify the difference. Check protocol versions, experiments, feature flags, optional fields, transport boundaries, server-specific workarounds, and generated-code lag. Do not select a source merely because it is easiest to copy.

### 4. Apply the focused change

For a dependency bump, inspect the upstream diff between the old and new revisions. Pay particular attention to:

- `minecraft/protocol/info.go`, packet pools, packet IDs, and codecs
- login, encryption handshake, compression negotiation, and batching
- unknown and invalid packet handling
- `minecraft.Network`, `PacketFunc`, `ReadPacket`, `Context`, and `Latency` extension points
- resource-pack announcements, download paths, chunking, content keys, and listener delivery
- authentication, transfer, and spawn behavior

Preserve BedrockDebugProxy observation hooks and raw evidence boundaries. If an upstream API no longer exposes a required boundary, document the fidelity loss before choosing a fork or compatibility layer.

### 5. Add focused automated evidence

Run the full quality gate and add tests proportional to the changed behavior. Protocol work commonly needs:

- encode and decode round trips for changed packets or structs
- exact binary fixtures for IDs, field order, optionals, and version gates
- malformed and truncated input tests that retain raw bytes and decode errors
- login and spawn flow coverage
- compression, encryption, batching, and framing boundary checks
- no-pack, single-pack, multi-pack, RakNet, HTTP, and encrypted resource-pack cases when affected
- capture verification and analysis of unknown or partially decoded packets

Do not update a fixture from new output without independently proving that the new bytes are correct.

### 6. Validate across real server categories

Automated tests do not replace interoperability testing. Exercise the categories relevant to the change:

- an owner-controlled Bedrock Dedicated Server
- a Featured Experience when the flow can be tested safely
- a Creator Experience
- another server implementation or public Other Server

For each target, record the server category and known software or infrastructure, login and resource-pack flow, protocol ID, first failure or divergence, expected result, and capture path. Do not require optional packets to arrive in the same order across targets unless the protocol evidence proves that ordering.

At minimum, verify connection, resource-pack negotiation, spawn, bidirectional packet forwarding, graceful disconnect, final capture verification, and `analyze` output. If a target fails, retain the raw capture and compare the earliest divergent packet or library error before adding a workaround.

### 7. Review regressions before workarounds

When a bug appears after an update, first re-check packet IDs, field order, optional fields, version gates, decode boundaries, compression negotiation, and assumptions inherited from the old version. Determine whether it affects one server category or all tested implementations.

A compatibility workaround requires evidence of valid behavior that the normal path cannot represent. It must be narrow, documented, tested, and visible in capture metadata when it changes fidelity or interpretation.

### 8. Commit and request human validation

Keep the dependency bump, adapter fix, and unrelated cleanup separate when they can be reverted independently. Update the research note and baseline table in the behavior-changing commit.

After each protocol, networking, or resource-pack commit, give the developer a concrete test plan. Name the target categories, exact flow, packets or events to observe, expected result, and capture or log to return on failure. Report automated and live validation separately.

## Completion gate

A protocol update is complete only when:

- every changed wire claim has traceable evidence
- source disagreements are resolved or explicitly marked unknown
- raw, unknown, malformed, and decoded observations remain representable
- the full quality gate passes
- focused fixtures and regression tests pass
- dependency and source licenses remain compatible and provenance is updated
- the compiled protocol and game version are visible in CLI and capture metadata
- required live test categories have results, or the missing validation is explicitly reported as pending

If the protocol evidence is insufficient, preserve the bytes and uncertainty and stop. Do not guess in order to make the update appear complete.
