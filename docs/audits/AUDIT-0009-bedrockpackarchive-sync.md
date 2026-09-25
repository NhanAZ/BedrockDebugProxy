# AUDIT-0009 - Require BedrockPackArchive synchronization for resource-pack changes

- Commit subject: `Require BedrockPackArchive synchronization for resource-pack changes`
- Change type: `documentation | workflow`
- Runtime impact: `no`

## Intent

Require coordinated BedrockDebugProxy and BedrockPackArchive changes whenever resource-pack code changes, so the pack archiver stays compatible with the capture producer.

## Scope

This change updates only `AGENTS.md` and this audit record. It defines required cross-repository review, paired updates, validation, private-remote handling, and commit provenance. No product code, capture schema, dependency, or runtime behavior changed.

## Evidence and provenance

- Reviewed the private BedrockPackArchive repository at commit `e4413bfb6c0ac2d7c11a4e382c4deb08e33b4357`.
- Its `main.go` imports `resource_pack.archive`, `resource_pack.decrypted_archive`, and `resource_pack.contents_manifest` data from `bedrockdebugproxy.capture.v1`. Its README documents the supported BedrockDebugProxy version range and reconstruction behavior.
- The new requirement makes producer-consumer review mandatory and keeps real captures, authentication state, `ContentKey` values, and decrypted pack data out of the companion repository's regression fixtures.

## Validation

- Automated checks: `pass - git diff --check -- AGENTS.md`
- Live validation: `not applicable - documentation and workflow only`
- Capture or report: `none`

## Limitations and follow-up

This record changes the required workflow but does not modify BedrockPackArchive. Future resource-pack implementation changes must include and validate the paired update before they are considered complete. Existing uncommitted worktree changes are not included in this policy-only commit.
