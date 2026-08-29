# Protocol update workflow

## Purpose

Minecraft Bedrock protocol updates are evidence changes, not dependency-refresh chores. A successful compile only proves that local APIs still type-check. It does not prove packet IDs, field order, optional fields, version gates, login, encryption, compression, resource-pack transfer, spawn, or compatibility with different server implementations.

Use this workflow for a gophertunnel upgrade, a new Bedrock release, a packet definition change, or a regression that first appears after either one.

## Current baseline

The baseline at repository commit `669557f` is:

| Component | Pinned value | Evidence |
| --- | --- | --- |
| Go toolchain | `go1.26.6` | `go.mod` toolchain directive |
| gophertunnel | `v1.61.0` | `go.mod` |
| gophertunnel source | `283a5a97dfe65da94bcc0b401807f6aefa9e72ee` | Upstream tag revision reviewed in `docs/research/bedrock-ecosystem.md` |
| Bedrock protocol | `2169` | `minecraft/protocol/info.go` in gophertunnel `v1.61.0` |
| Bedrock game version | `1.26.45` | `minecraft/protocol/info.go` in gophertunnel `v1.61.0` |
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

Prefer authoritative runtime evidence and primary source definitions. Cross-check at least two independent sources when practical. Useful sources include the pinned gophertunnel and go-raknet revisions, Cloudburst Protocol, PrismarineJS bedrock-protocol, Endstone spyglass and protocol-dumper output, Dragonfly, owner-authorized Bedrock Dedicated Server captures, and captures from distinct public server categories.

Source agreement is not enough when all sources may share an outdated assumption. A live capture can confirm bytes and ordering for one server and version, but it does not prove that every valid server must behave identically.

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
