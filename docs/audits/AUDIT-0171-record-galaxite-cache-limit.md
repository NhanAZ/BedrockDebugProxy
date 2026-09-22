# Audit AUDIT-0171

## Intent

Record the Galaxite release-gate decode failure and its source comparison without changing runtime behavior.

## Scope

Added `docs/research/galaxite-client-cache-blob-status-limit.md`. The note records the exact capture evidence, vendored gophertunnel limit, comparison with BedrockProtocol, source revisions, and the focused follow-up required before a decoder change. No product, protocol, capture, or dependency code changed.

## Evidence and provenance

The evidence is capture `fe9d3e3b-aa80-4976-a6ec-31db99dd1db1` from runtime revision `93435c5119e2899dc1958d8cec0fa63447bd611c`. The source comparison uses the vendored files and immutable upstream revisions linked in the research note.

## Automated validation

- `git diff --check` passed.
- Documentation-only change. No runtime tests are required.

## Live validation status

The Galaxite report remains `fail` for the exact runtime revision because one blocking decode error was observed. No new live session was run for this documentation-only record.

## Known limitations and follow-up

The source comparison establishes a likely safety-limit mismatch but does not yet prove the correct replacement limit or justify a generic bound change. Add a focused fixture and regression test before implementing a fix.
