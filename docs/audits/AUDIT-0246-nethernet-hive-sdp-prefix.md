# AUDIT-0246 - Fix The Hive NetherNet SDP candidate prefix

## Scope

Investigate the report that a retail Minecraft client could join The Hive while
the agent-controlled BedrockDebugProxy path failed during NetherNet signaling.
The failure was observed before Bedrock login, resource-pack exchange, or
world spawn.

## Root cause

The selected go-nethernet v1.0.24 implementation formatted a locally gathered
ICE candidate with a `candidate:` prefix and then embedded that value under the
SDP `a=candidate` attribute. The resulting value had a duplicated prefix. The
upstream commit [`60826b5e5a8f13498de588eacbbabefcc9402883`](https://github.com/df-mc/go-nethernet/commit/60826b5e5a8f13498de588eacbbabefcc9402883)
separates the formatting used for SDP attributes from the formatting used for
trickle-candidate signals and explicitly identifies the fix as resolving The
Hive connectivity issues.

The HTTP endpoint and address were not the defect. The successful fixed run
used The Hive's HTTPS NetherNet endpoint, and no server-specific request
mutation or policy bypass was added.

## Implementation

- Pin `github.com/df-mc/go-nethernet` to
  `v1.0.27-0.20261027203254-60826b5e5a8f` in the root and replacement-module
  graphs.
- Add a regression test that creates a full-ICE offer and rejects the malformed
  `a=candidate:candidate:` form.
- Record the dependency provenance and live evidence in the protocol review,
  fix index, notices, and Unreleased changelog.

## Validation

- Focused Experience tests passed, including the SDP candidate-prefix regression.
- Candidate binary: `0.5.0-nethernet-minimal-fix-test`, protocol 2225,
  Bedrock 1.26.61, source tree based on `a8a314f63446e695b8fd29563e9c08fec6e725de`
  plus this dependency/test change.
- Agent-controlled The Hive flow passed on 2026-11-01: authenticated
  NetherNet negotiation, 25 resource packs, spawn, 35 seconds of stable
  bidirectional traffic, 6,453 packets read, 35 packets written, and automatic
  capture verification with 23,581 events and 8,305 blobs.
- The same endpoint and cached account had failed with both the previous
  candidate and the old v0.4.2 binary, each at the HTTP 400 SDP-offer step.

This audit records candidate evidence. It is not a release approval and does
not replace the exact-commit six-server matrix.
