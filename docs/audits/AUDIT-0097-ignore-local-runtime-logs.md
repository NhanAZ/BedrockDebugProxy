# Audit 0097 - Ignore local runtime logs

## Intent

Keep operator-generated proxy and local-server logs out of Git status so the canonical validation build can stamp a committed source revision without touching those files.

## Scope

- Add root-scoped `.gitignore` patterns for `proxy-agent-*.log` and `server.log`.
- Preserve all existing log files in place.
- No runtime, capture, protocol, dependency, or CLI behavior change.

## Evidence and provenance

Before this change, `tools/build.ps1` refused to build because the working tree contained these untracked runtime logs. The patterns are limited to the exact root-level names produced by the local workflow.

## Automated validation

- `git diff --check`
- Full `tools/quality.ps1` is required before commit.
- `tools/build.ps1 -Version dev` will be run after this commit to verify the stamped revision.

## Live validation

Not applicable. This is a repository hygiene change only.

## Known limitations and follow-up

Existing log files remain local and are not retroactively deleted. Other ad hoc log names are intentionally not ignored.
