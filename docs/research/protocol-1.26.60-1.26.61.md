# Bedrock 1.26.60 and 1.26.61 protocol update

This note records the evidence for moving the selected Bedrock protocol from
`2193` to `2225`. The project advertises the latest stable game metadata,
`1.26.61`, while using the reviewed `1.26.60` packet codec because the
official `1.26.61` release reports no packet changes.

Sources were checked on 2026-11-01.

## Evidence and decision

| Claim | Protocol version | Source and revision | Independent check | Decision | Confidence |
| --- | --- | --- | --- | --- | --- |
| Stable Bedrock 1.26.60 uses network protocol 2225 and introduces packet changes | `2225` | Mojang [`v1.26.60` release](https://github.com/Mojang/bedrock-protocol-docs/releases/tag/v1.26.60), tag commit [`0879ab5c4c70be1bd520e32a8c49ea4a155616d6`](https://github.com/Mojang/bedrock-protocol-docs/tree/0879ab5c4c70be1bd520e32a8c49ea4a155616d6) | The official [protocol changelog](https://mojang.github.io/bedrock-protocol-docs/changelog/) lists 64 packet changes for 1.26.60 | Import the reviewed 1.26.60 packet definitions and select protocol 2225 | Confirmed |
| Stable Bedrock 1.26.61 uses network protocol 2225 and has no packet changes from 1.26.60 | `2225` | Mojang [`v1.26.61` release](https://github.com/Mojang/bedrock-protocol-docs/releases/tag/v1.26.61), tag commit [`21e84e70fc6385d40d448ca8a2e12d959e9d185e`](https://github.com/Mojang/bedrock-protocol-docs/tree/21e84e70fc6385d40d448ca8a2e12d959e9d185e) | The official [protocol changelog](https://mojang.github.io/bedrock-protocol-docs/changelog/) lists zero packet changes for 1.26.61 | Advertise game version 1.26.61 without another packet-code import | Confirmed |
| The upstream 1.26.60 implementation adds the new packet IDs and supporting structures | `2225` | Sandertv gophertunnel `v1.63.0`, commit [`8c30047d86617a56e1f1871e3a1d47ec4f670d99`](https://github.com/Sandertv/gophertunnel/commit/8c30047d86617a56e1f1871e3a1d47ec4f670d99), MIT-licensed source | The packet IDs and registrations match the official release metadata and the existing packet-pool architecture | Import the codec changes while preserving BedrockDebugProxy's local observation hooks | Confirmed |
| Four 1.26.60 packet layouts were corrected after the initial upstream release | `2225` | Sandertv gophertunnel commit [`d4836105abdc5de0496cf62b69413a591af5f8fb`](https://github.com/Sandertv/gophertunnel/commit/d4836105abdc5de0496cf62b69413a591af5f8fb), MIT-licensed source | The corrected field optionality and enum encodings are consistent with the official 1.26.60 schema release | Import the corrections for camera spline, player video capture, and structure block update packets | Confirmed |
| 1.26.60 clients can send a single-port NetherNet URL in login metadata | `2225` | Sandertv gophertunnel commit [`531fd12d9a67985af5fa9552e88745350da87bfb`](https://github.com/Sandertv/gophertunnel/commit/531fd12d9a67985af5fa9552e88745350da87bfb), MIT-licensed source | The existing validator already handles the older two-port form; the new branch is narrow and retains the original validation rules | Accept both observed URL forms in `ClientData.Validate` | Confirmed |
| The selected NetherNet transport must omit the literal `candidate:` prefix from SDP candidate attribute values | NetherNet signaling | go-nethernet commit [`60826b5e5a8f13498de588eacbbabefcc9402883`](https://github.com/df-mc/go-nethernet/commit/60826b5e5a8f13498de588eacbbabefcc9402883), whose subject explicitly identifies The Hive connectivity | The older dependency emitted `a=candidate:candidate:...`; the fixed candidate reached The Hive, completed 25 packs, spawned, retained 35 seconds of bidirectional traffic, and verified its capture | Pin the fixed upstream snapshot in both module graphs and keep a regression test for the malformed prefix | Confirmed |

## Scope and disagreements

The selected source is the official stable 1.26.60 schema plus the stable
1.26.61 no-packet-change release. Preview-only 1.26.70 schemas are excluded.
The upstream 1.26.60 commit also upgrades `go-nethernet` to `v1.0.25`; that
dependency contains separate client-identity and endpoint changes and is not
needed to decode protocol 2225 or validate the new login URL form. A later
live The Hive failure exposed an independent dependency defect: the released
implementation duplicated the `candidate:` token when embedding ICE
candidates in SDP. The project now pins the reviewed upstream snapshot
`v1.0.27-0.20261027203254-60826b5e5a8f`, which contains the targeted fix from
`60826b5`, in both module graphs. This is a NetherNet interoperability fix,
not a packet-codec change or a server-specific workaround.

The in-tree gophertunnel copy remains a composite rather than a wholesale
upstream replacement. Existing BedrockDebugProxy resource-pack, raw-packet,
capture, transfer, NBT, and transport patches are retained. The imported
upstream code is MIT-licensed and its provenance remains recorded in
`third_party/gophertunnel/BEDROCKDEBUGPROXY_PATCH.md` and
`THIRD_PARTY_NOTICES.md`.

## Validation limits

The automated evidence covers protocol metadata, packet registration, exact
encodings for newly added packets, corrected optional fields, malformed input,
the SDP candidate-prefix regression, and the existing capture/forwarding
regressions. A candidate binary using the fixed dependency completed the
agent-controlled The Hive flow on 2026-11-01: 25 packs, spawn, 35 seconds of
bidirectional traffic, and automatic capture verification. This is protocol
interoperability evidence, not a retail rendering or gameplay claim. The
complete six-server release matrix still requires a fresh exact-revision run.
Unknown packets and any live decode uncertainty remain preserved by the raw
capture boundary.
