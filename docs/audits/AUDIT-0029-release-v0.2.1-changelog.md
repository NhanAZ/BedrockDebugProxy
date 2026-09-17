# AUDIT-0029 - Prepare v0.2.1 release notes

- Commit subject: `docs: prepare v0.2.1 release notes`
- Change type: `documentation | release`
- Runtime impact: `no`

## Intent

Close the reviewed `Unreleased` entries for the v0.2.1 release while keeping a new empty section for subsequent work. The release follows the v0.2.0 baseline and contains the already validated 1.26.50 sub-chunk height-map correction plus the associated maintenance and research documentation.

## Scope

Move the user-facing entries in `CHANGELOG.md` into the dated v0.2.1 section and retain a single root changelog file. Add the release link and compare link that will be valid after the tag is published. This commit does not alter protocol code, tools, workflows, configuration, captures, validation reports, or build output.

## Evidence and provenance

- The candidate runtime revision is `8425d04ad5990ec9be4c1056d122ac2dde898870`.
- The v0.2.0 release is tagged at the preceding release baseline. Commits through `8425d04` were reviewed with `git log`, the per-commit audit records, the protocol research notes, and the changed-file diff.
- The five-server reports for `8425d04` pass for The Hive, Galaxite, Lifeboat, Mineville Zeqa (Prison flow), and Enchanted. The existing Mineville PvP failure remains historical evidence and is not replaced.

## Validation

- Automated checks: pending after this docs-only commit. The release workflow will run `tools/quality.ps1`, the docs-only release equivalence gate, the stamped build, `git diff --check`, and CI for the exact candidate commit.
- Live validation: pass on runtime revision `8425d04` for all five required release targets. This commit is documentation-only and will reuse those reports through the explicit allowlist gate.
- Capture or report: reports remain under ignored `validation/local/8425d04ad5990ec9be4c1056d122ac2dde898870/` and are not copied into Git history.

## Known limitations and follow-up

The release is not complete until the exact candidate commit is built, signed, pushed, verified in GitHub Actions, published with the reviewed notes and assets, and synchronized to `NhanAZ/BedrockDebugProxy-Backup`. The release does not claim universal server compatibility or resolve the pending Mojang pack-encryption proposal.
