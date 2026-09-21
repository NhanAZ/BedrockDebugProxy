# Audit 0100 - Add a low-memory quality-gate mode

## Intent

Provide a supported way to run the complete local quality gate without allowing Go's default parallelism to exhaust workstation memory.

## Scope

- Add an opt-in `-LowMemory` switch to `tools/quality.ps1`.
- Bound child Go commands to `GOMAXPROCS=2` and package parallelism `-p=1`.
- Restore any caller-provided `GOMAXPROCS` and `GOFLAGS` values when the script exits.
- Keep the default local and CI profiles unchanged.

## Evidence and provenance

The existing quality gate invokes tests, builds, golangci-lint, and govulncheck through Go. The bounded profile uses the same commands and checks, only changing their process-level concurrency.

## Automated validation

- `tools/quality.ps1 -LowMemory` passed, including tests, build, static analysis, vulnerability scan, and artifact checks.
- PowerShell syntax validation passed as part of the quality gate.
- `git diff --check` passed.

## Live validation

Not applicable. This changes only local validation-process resource usage.

## Known limitations and follow-up

The bounded profile is slower than the default profile. CI continues to use its existing default and race-enabled commands.
