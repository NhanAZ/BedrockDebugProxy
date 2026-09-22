# Audit AUDIT-0169 - Report closed-capture verification progress

## Intent

Make long capture verification visibly active in both the run terminal and its live viewer.

## Scope

- Add coarse event and blob counters to the existing capture verification path.
- Publish verification progress through the live viewer session response while the closed capture is checked.
- Keep machine-readable `verify` output on stdout and send progress diagnostics to stderr.
- Update tests, fix history, and the user-facing changelog.

The capture schema, packet forwarding path, artifact contents, and final verification result are unchanged.

## Evidence and provenance

The owner observed a 925,354-event capture remain silent for several minutes after automatic artifacts completed. The synchronous verifier scans event lines and unique blob files, but previously exposed only its final result. The live viewer also attempted its own blocking verification after the manifest closed.

## Automated validation

- `go test ./internal/capture ./internal/viewer ./cmd/bedrock-debug-proxy` passed.
- `node --check internal/viewer/assets/app.js` passed.
- `git diff --check` passed.
- `tools/quality.ps1 -LowMemory` passed, including formatting, tests, build, static analysis, vulnerability scan, tracked-capture checks, and whitespace checks.

## Live validation

Browser smoke validation used a local 100,000-event closed capture with an externally managed live viewer. The page displayed `Verifying capture - scanned 60,000 events and verified 0 blobs`, then advanced to `Verifying capture - checking manifest` while the scan continued. The viewer API regression test confirms that completion clears progress and returns the session to `verified`. No public-server or real-client validation is required for this observability-only change.

## Known limitations and follow-up

Progress is intentionally coarse and reports every 10,000 events plus start and completion milestones. It is an activity indicator, not a precise time estimate or percentage.
