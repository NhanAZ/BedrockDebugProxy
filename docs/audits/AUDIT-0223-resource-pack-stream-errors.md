# AUDIT-0223 - Preserve resource-pack stream failures

- Commit subject: `fix: preserve resource-pack stream read failures`
- Change type: resource-pack error propagation and regression coverage
- Runtime impact: failed source streams are rejected before ZIP parsing and their original errors remain observable.

## Intent and scope

The [capture failure review](../research/capture-failure-review-2026-10-18.md) records seven prioritized sessions, their verified private archives, and the distinction between external failures and the reproduced local defect. FIX-0160 checks copy and close failures in the retained resource reader and cleans up its temporary file. The successful archive path and capture contract remain unchanged. No third-party patch was copied and the existing MIT notice is retained.

The change includes the reader, synthetic before/within/after-archive failures, a successful byte-preserving read, a truncated chunked HTTP regression through the URL cache, resource-pack documentation, changelog, and dependency provenance. Viewer fields, CLI options, authentication, retry policy, and transfer behavior are unchanged.

## Validation

The three reader failure cases fail on the original source and pass after the correction. The full vendored resource package tests and focused URL cache tests pass. The canonical tools/format.ps1 and tools/quality.ps1 -LowMemory pass with Go 1.27.2, including root tests, build, viewer syntax, module and license checks, static analysis, vulnerability scanning, and whitespace. The first static-analysis pass found unchecked test-only hijack/close handling, which was corrected before the passing gate. Full consumer tests, vet, build, whitespace, and its fixture with the existing actual producer verifier pass. The follow-up audit records the clean stamped producer fixture check, both final Git revisions, and live scope.

## Paired consumer

The BedrockPackArchive update extends the synthetic fixture to a 0.4.0 producer and validates unchanged root/nested manifests, archive digest, plaintext metadata, contents.json, key count, both hops, and index provenance. Its version gate accepts the reviewed numeric 0.4.x producer needed for live testing alongside 0.3.x. No importer data-path change is required. The consumer commit is `cbfa0879f24b141873bef4042e8de048e1f38cef`. The follow-up paired audit records this runtime commit as well.

## Limits and human validation

This does not fix external TLS/EOF, CDN resets, DNS outages, or backend refusal. Use an authorized retail client to select an Enchanted child mode and verify Transfer, child pack exchange, child spawn, and a continued connection. If it fails, retain the full capture and terminal diagnostics. The Hive packs-only baseline and an Enchanted automated child exchange remain distinct from retail gameplay and the release matrix.
