# Audit AUDIT-0075 - Record ClientData parity in negotiation evidence

## Intent

Make a local parity investigation able to distinguish the protocol implementation version from
the game version and device OS reported by each login's `ClientData`.

## Scope

- Updated `internal/proxy/proxy.go` to add the downstream and upstream `ClientData` game versions,
  device OS values, and explicit match booleans to `session.negotiated`.
- Updated `internal/analysis/analysis.go` and its test to expose the optional values in the
  machine-readable analysis summary.
- Updated the proxy integration assertion, capture-format documentation, fix history, and
  changelog.

## Evidence and provenance

The Hive incident research note records a client-reported version difference between the closest
successful capture and the affected capture. Existing `session.connection_metadata` snapshots
already contain the source values. This change adds a compact comparison without changing packet
forwarding, authentication, or transport behavior.

## Automated validation

- `gofmt` passed for all changed Go files.
- `go test ./internal/analysis ./internal/proxy` passed.
- `git diff --check` passed.

## Live validation

No live server was contacted. The previously completed operator-controlled PocketMine loopback
comparison remains the relevant runtime evidence. This commit changes capture metadata only.

## Known limitations and follow-up

The new fields do not make authenticated sessions identical to a vanilla client and do not
identify any public server's enforcement rule. Older captures omit the new optional summary
fields and remain valid.
