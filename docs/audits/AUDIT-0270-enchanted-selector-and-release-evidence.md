# AUDIT-0270 - Correct Enchanted selector and record release validation

- Commit subject: `Correct Enchanted selector and record release validation`
- Change type: `documentation`
- Runtime impact: `no`

## Intent

Correct stale release instructions after the old `experience:Enchanted` selector stopped resolving, and record the exact-revision release gate evidence.

## Scope

Update the Enchanted selector in the release, validation, server target, and Golem command documentation. Update the changelog, close the related fix-history entries, and record the completed six-server evidence for the runtime change in `AUDIT-0269`. Report labels remain `Enchanted`. No runtime or capture behavior changes.

## Evidence and provenance

The live Experience resolver rejected `experience:Enchanted` as not found. The exact current entry `experience:Enchanted [Skyblock]` authenticated, reached spawn, captured the automatically sent Hub form, and closed immediately. This report provides protocol-client evidence for the selected Experience, not retail rendering or gameplay.

## Validation

- Automated checks: `tools/quality.ps1` passed. The stamped docs-candidate build, hosted CI, and docs-only release equivalence remain pending.
- Live validation: pass on runtime revision `84ab31396cc2b80aa094c47d924489a69581ed43` for all six required release targets.
- Capture or report: six sanitized reports under the ignored `validation/local/84ab31396cc2b80aa094c47d924489a69581ed43/` directory. Every corresponding local capture passed `bedrock-debug-proxy verify`.

## Limitations and follow-up

This commit is documentation-only and introduces no new live-session claim. Keep the original report revisions unchanged. Complete the docs-only equivalence gate before tagging the release.
