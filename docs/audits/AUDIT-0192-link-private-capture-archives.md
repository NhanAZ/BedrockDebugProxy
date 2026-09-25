# Audit AUDIT-0192 - Link private capture archives

- Commit subject: `docs: link private capture archives`
- Change type: `documentation`
- Runtime impact: `no`

## Intent

Make documentation references to preserved capture sessions open the corresponding private archive so local copies can be reviewed for cleanup.

## Scope

- Replace 131 plain or local-path references across audit records, decisions, fix history, and research notes with links to the matching private capture archive release.
- Update storage statements that described the archived captures as local-only.
- Add no capture data, code, protocol behavior, credentials, or public artifacts to the source repository.

## Evidence and provenance

The 131 references identify 35 unique sessions. Each session link was matched to a same-named repository in the `BedrockDebugProxy-Captures` organization. The organization listing returned exactly 35 repositories, all private, with no missing or unreferenced repositories. The linked local session directories are still present. One older Hive observation in `docs/research/resource-pack-negotiation.md` does not record a session identifier, so the note identifies that missing archive citation instead of guessing.

## Validation

- Automated checks: a documentation link-integrity scan found 131 valid archive links, 35 unique targets, and no remaining unlinked session references. `git diff --check` passed.
- Full project quality and test gates were not run because this change only edits Markdown documentation.
- Live validation: not applicable. This change has no runtime impact.

## Limitations and follow-up

The archive links require access to the private organization and will not be available to public readers. No local capture directories were deleted. The older Hive observation without a session identifier remains unlinked until its source can be identified.
