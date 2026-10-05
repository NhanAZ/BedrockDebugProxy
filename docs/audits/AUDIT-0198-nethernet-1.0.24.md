# AUDIT-0198 - Update NetherNet transport to v1.0.24

- Commit subject: `deps: update go-nethernet to v1.0.24`
- Change type: `runtime | networking`
- Runtime impact: `yes`

## Intent and scope

Adopt the released NetherNet transport corrections after confirming that the stable Bedrock protocol remains 2193 and game version 1.26.51. Update the root and nested gophertunnel module graphs, required Pion versions, dependency notices, and changelog. No wholesale transport or packet-tree replacement is performed.

## Evidence and provenance

The [maintenance review](../research/protocol-maintenance-2026-10-05.md) records the v1.0.23 to v1.0.24 diff and immutable upstream revision `90bfcd1be9a4762a2d0c06243f880d195408561f`. This is MIT dependency use. The update allows an empty identity-provider domain in the envelope while retaining JWT and signed-fingerprint verification. The project authenticated dialer still supplies its existing nonempty domain. The previously documented server-token temporal policy is unchanged.

The direct-dependency notices also correct older version labels to match the existing root go.mod. Those label corrections do not update those dependencies.

## Validation

- Automated checks: `tools/format.ps1`, `tools/quality.ps1 -LowMemory`, and `git diff --check` passed with Go 1.26.6 on the reviewed maintenance source. The quality gate included viewer asset and HTTP tests, JavaScript syntax, all root tests, build, static analysis with zero issues, module integrity, and 36 compiled-module license checks. Govulncheck found no project-reachable vulnerability and six uncalled dependency findings. The v1.0.24 upstream module suite passed `go test -count=1 ./... -timeout 120s` from its module directory, including its transport, identity, candidate, endpoint, and discovery tests. Root-graph transport and endpoint tests also passed. A root-graph attempt to include unused discovery tests lacked that package's optional checksums, so the complete suite was rerun successfully in the upstream module rather than adding an unused runtime dependency.
- Live validation: `not_observed`.
- Capture or report: none. Test traffic is synthetic and local.

## Limitations and human validation

Use an authorized NetherNet Experience or owner-controlled HTTP signaling endpoint and exercise login, resource-pack exchange, cancellation, and disconnect. Confirm that resolution reports NetherNet rather than RakNet, the transport identity reaches the observation wrapper, and the closed capture verifies. Return the closed capture and first failure log if negotiation fails. A RakNet-only The Hive session cannot validate these changes. Retail-client and official release validation remain pending.
