# Audit 0106 - Preserve validation report timestamps

- Intent: keep capture provenance deterministic in sanitized validation reports.
- Scope: update `tools/create-validation-report.ps1` to format parsed capture timestamps as invariant UTC ISO 8601 values, document the correction in `CHANGELOG.md`, and index it in `docs/fix-history.md`.
- Affected interfaces: the validation report schema fields `capture.started_at` and `capture.ended_at`. No proxy wire behavior, capture schema, raw evidence, or runtime forwarding path changed.
- Evidence and provenance: the report generated from `captures/session-20260921T122228Z` showed locale-formatted timestamps after the existing script parsed the capture analysis. The canonical capture and `analyze` output retain the original UTC timestamps.
- Automated validation: run PowerShell syntax validation, `tools/quality.ps1 -LowMemory`, and `git diff --check` before commit.
- Local validation: regenerate a fresh report from the same local capture and verify both timestamp fields match the ISO 8601 UTC form with fractional seconds.
- Live validation: not applicable. This is a validation-tool and documentation change and does not contact any server.
- Known limitations: reports generated before this change are not rewritten, and the report remains a derived summary rather than raw packet evidence.
- Follow-up: use the corrected formatter for any future report timestamp fields.
