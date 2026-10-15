# AUDIT-0222 - Scrub private identifiers before public visibility

- Commit subject: `docs: scrub private identifiers before public visibility`
- Change type: documentation and test-fixture privacy cleanup
- Runtime impact: none.

## Intent and scope

Prepare the existing private repository for the maintainer-requested visibility change. Remove personal account names, a local network address, token-scoped local viewer URLs, and workstation paths from tracked text and retained Git history. Remove research and audit documents that reproduce the maintainer's individual public-server enforcement incident or first-person Discord messages. Keep general protocol findings, official server policy links, and test behavior with synthetic names.

Rewrite the local branch and tag history so the old text is no longer reachable from public refs. Preserve the pre-cleanup repository in the private recovery remote and an ignored local bundle. Re-sign the v0.4.0 tag at the rewritten release commit. The release assets and validated runtime are not modified.

## Validation

- Search the working tree and every rewritten public ref for the removed account labels, LAN address, workstation paths, private incident notes, and token-scoped localhost URLs.
- Confirm no capture, authentication cache, private key, or resource-pack data is tracked.
- Run `git diff --check` and inspect rewritten refs, tag signatures, the published release, and private backup refs before changing repository visibility.
- Application tests are not run because the changes only replace test labels and redact or remove documentation; no executable behavior changes.

## Limitations

The repository will retain public technical provenance and the maintainer's public project identity. The private recovery backup retains the pre-cleanup history and remains private.
