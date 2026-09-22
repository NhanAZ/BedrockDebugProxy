# Audit AUDIT-0175

## Intent

Reduce idle time during active upstream resource-pack exchanges while preserving packet contents, capture ordering, and the existing downstream delivery boundary.

## Scope

Add a bounded 100-request upstream chunk window, response validation, decoder-owned byte copies, and a bounded out-of-order reorder buffer in the vendored gophertunnel connection. Keep one resource pack active at a time, keep downstream listener delivery sequential with its existing pacing, and leave archive storage, decryption, and capture schema unchanged. Update resource-pack documentation, research, fix history, changelog, and third-party provenance.

## Evidence and provenance

The Galaxite capture `captures/session-20260922T151050Z` at revision `a992d6c37f07ae04afb651ce2130712f9e4b209e` recorded 197 requests and 186 responses during the previous five-minute deadline while the exchange was still active. The pinned gophertunnel loop was serial. The request-window behavior is supported by [gophertunnel PR #509](https://github.com/Sandertv/gophertunnel/pull/509), reviewed as an open upstream research reference on 2026-09-22. No source was copied from the PR.

## Automated validation

- `go test ./minecraft` in `third_party/gophertunnel` passes, including `resource_pack_pipeline_test.go` for chunk-count bounds, the initial request window, replenishment, requested-index validation, and response-byte copying.
- `gofmt` was run on the changed Go files.
- Project low-memory quality and stamped build checks are required before commit completion.

## Live validation status

Pending. The prior Hive report is not reusable because upstream resource-pack scheduling changed. A fresh stamped The Hive baseline and Galaxite session must reach login, spawn, and orderly shutdown at the exact revision.

## Known limitations and follow-up

The 100-chunk maximum is a bounded compatibility choice, not a claim about every server's optimal window. The downstream send delay remains because send-side congestion control has not been established. If a live server rejects concurrent requests, record the capture and classify the behavior before changing the generic window.
