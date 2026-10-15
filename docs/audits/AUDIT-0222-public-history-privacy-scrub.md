# AUDIT-0222 - Scrub private identifiers before public visibility

- Commit subject: `docs: scrub private identifiers before public visibility`
- Change type: documentation and test-fixture privacy cleanup
- Runtime impact: none.

## Intent and scope

Prepare the existing private repository for the maintainer-requested visibility change. Remove personal account names, a local network address, token-scoped local viewer URLs, and workstation paths from tracked text and retained Git history. Remove research and audit documents that reproduce the maintainer's individual public-server enforcement incident or first-person Discord messages. Keep general protocol findings, official server policy links, and test behavior with synthetic names.

Rewrite the branch and tag history so the old text is no longer reachable from those refs. Preserve the pre-cleanup repository in the private recovery remote and an ignored local bundle. Re-sign the v0.4.0 tag at the rewritten release commit. The release assets and validated runtime are not modified.

## Validation

- Search all commits reachable from the rewritten branch and tags for the removed account labels, LAN address, workstation paths, private incident notes, and token-scoped localhost URLs.
- Confirm no capture, authentication cache, private key, or resource-pack data is tracked.
- Run `git diff --check`, compare the rewritten main tree with the reviewed cleanup tree, inspect tag signatures, and scan the published release assets for those identifiers.
- Application tests are not run because the changes only replace test labels and redact or remove documentation; no executable behavior changes.

## GitHub pull request refs

After the rewritten branch and tags were pushed, GitHub still exposed generated refs for closed pull requests. PR #4 and PR #5 point to pre-cleanup commit trees that contain the removed private markers. They are a merged Dependabot update and a closed Dependabot update. GitHub's REST API rejected deletion of `refs/pull/4/head` and `refs/pull/5/head` with HTTP 422 because `refs/pull/*` is read-only.

Repository visibility was briefly changed to public, then returned to private as soon as these refs were found. The repository remains private until GitHub removes or otherwise makes the old pull request objects inaccessible. The current rewritten branch and tags are clean, but these GitHub-managed refs prevent claiming that the public repository is ready.

The available repository traffic view data ends on 2026-10-14 and therefore cannot establish whether anyone accessed the repository during the brief public interval on 2026-10-15.

## Limitations

The repository will retain public technical provenance and the maintainer's public project identity. The private recovery backup retains the pre-cleanup history and remains private. The pre-cleanup pull request refs are a separate GitHub-managed retention boundary and still require resolution.
