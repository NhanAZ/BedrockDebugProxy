# Audit AUDIT-0174

## Intent

Allow an active, slow featured-experience resource-pack exchange to complete instead of cancelling it at the five-minute boundary.

## Scope

Increase the bounded upstream login and resource-pack deadline from five to fifteen minutes. Update user-facing and agent-facing documentation, the Galaxite research note, fix history, and changelog. Packet ordering, resource-pack contents, capture storage, and authentication behavior are unchanged.

## Evidence and provenance

The Galaxite capture `captures/session-20260922T151050Z` at revision `a992d6c37f07ae04afb651ce2130712f9e4b209e` recorded 186 server-to-client chunk packets and 197 client-to-server chunk requests across the five-minute interval. The final event was an explicit `upstream.dial_error` timeout, with no decode error or resource-pack error. The earlier Galaxite research capture documented 460 advertised chunks and the same pacing pattern.

## Automated validation

- `gofmt` and project quality checks must pass for the updated constant and documentation.
- Existing timeout error tests continue to cover bounded failure reporting.

## Live validation status

Pending a fresh stamped Galaxite session at the new revision. The prior Hive baseline is not reusable because this runtime behavior changed.

## Known limitations and follow-up

Fifteen minutes is a bounded estimate based on one active Galaxite flow, not a universal resource-pack duration. If a later capture shows progress beyond this window, replace the fixed bound with a progress-aware policy only after deterministic tests and evidence.
