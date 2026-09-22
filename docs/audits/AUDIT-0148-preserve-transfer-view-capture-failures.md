# Audit 0148 - Preserve Transfer view capture failures

## Intent

Prevent a recorder failure while capturing the derived Transfer rewrite view from becoming a nil-pointer panic on typed forwarding paths that do not retain raw packet ownership.

## Scope

- Updated `internal/proxy/proxy.go` to make raw-packet cleanup conditional in the derived Transfer view failure path.
- Added `TestRecordTransferRewriteFailureAllowsTypedTransferWithoutRawRead` in `internal/proxy/transfer_test.go`.
- Updated `docs/research/capture-observability.md`, `docs/fix-history.md`, and `CHANGELOG.md`.

## Evidence and provenance

- The forwarding path sets `rawRead` to nil when a packet is already decoded by gophertunnel or when typed forwarding is required.
- The adjacent Transfer rewrite error path already used conditional cleanup, while the derived-view recording error path still called `Release` unconditionally.
- The regression test closes the canonical recorder and exercises the derived-view failure with a nil raw packet.

## Automated validation

- `TestRecordTransferRewriteFailureAllowsTypedTransferWithoutRawRead` passes.
- The full project quality gate and `git diff --check` are required before commit.

## Live validation status

No live or public-server validation was performed. The failure path is local recorder error handling.

## Known limitations and follow-up

Normal Transfer forwarding and opt-in hop following remain covered by the existing integration test. Live capture-writer failure behavior remains unverified.
