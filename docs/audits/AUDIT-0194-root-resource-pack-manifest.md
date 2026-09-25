# Audit AUDIT-0194 - Prefer root resource-pack manifests

- Commit subject: `Prefer root resource-pack manifests`
- Change type: `runtime`
- Runtime impact: `yes`

## Intent

Prevent a nested placeholder `manifest.json` from overriding the resource-pack identity declared by the archive-root manifest.

## Scope

- Prefer the exact root path in the vendored gophertunnel archive lookup, retaining the existing nested-file fallback.
- Add a synthetic regression test for root manifest identity selection.
- Update the changelog, fix history, and research note.
- Pair with BedrockPackArchive commit `95c919474fb2dfa9549527a2c5f6389fdc8f165f`. Its synthetic capture test verifies the existing importer contract, so no importer runtime code change was needed.

## Evidence and provenance

The Microsoft Learn Add-Ons manifest reference describes `manifest.json` as pack identification data and its header UUID and version fields. The symptom was reported during an owner-run OneBlockOnline pack-only session. That temporary capture was removed and is unavailable for independent review, so the capture-backed cause remains unverified. A synthetic ZIP regression test now reproduces the relevant entry ordering without private pack data. The vendored gophertunnel source is MIT licensed; its base and merged revisions are documented in `THIRD_PARTY_NOTICES.md`.

## Validation

- Automated checks: `tools/format.ps1` passed. `tools/quality.ps1 -LowMemory` passed, including Go tests, build, static analysis, and vulnerability scan. BedrockPackArchive `go test -count=1 ./...`, `go vet ./...`, and `go build ./...` passed. Its synthetic importer test stubs only the external `verify` process. BedrockPackArchive commit `95c919474fb2dfa9549527a2c5f6389fdc8f165f` was pushed to its existing private `main` branch. GitHub Actions are not configured in that repository. BedrockDebugProxy's pushed commit passed both jobs in run [36143436642](https://github.com/NhanAZ/BedrockDebugProxy/actions/runs/36143436642); the paired commit hashes are recorded in AUDIT-0195.
- Live validation: `pending` for the exact runtime revision. The Hive baseline and direct OneBlockOnline resource-pack flow have not been recorded as stamped reports for this revision.
- Capture or report: `none for this revision`. No captures, account data, keys, or third-party pack contents were added.

## Limitations and follow-up

The synthetic regression verifies parser selection, not retail client or server behavior. After the exact revision is committed, validate the normal The Hive client session and a OneBlockOnline pack exchange that exercises root-manifest selection. Confirm pack acceptance and identity, both traffic directions, packet decoding, and clean shutdown. If either flow fails, retain the closed capture locally for diagnosis and produce the sanitized report with the project tooling. The original symptom capture remains unavailable.
