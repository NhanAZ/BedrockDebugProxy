# AUDIT-0180 - Clarify the scope of the `--force` override

- Commit subject: `Clarify force project guidance override`
- Change type: `process`
- Runtime impact: `no`

## Intent

Clarify that a leading `--force` suspends all project-level guidance for that
prompt, rather than only individual rules that conflict with the requested
objective. Keep higher-priority platform and tool constraints outside its
scope.

## Scope

Update the `--force` section in `AGENTS.md` and its accepted decision in
`docs/decisions/0012-force-prompt-project-workflow-override.md`. No product
code, CLI behavior, capture format, or release artifact changes.

## Evidence and provenance

The maintainer clarified that `--force` is intended to override every rule in
the project's `AGENTS.md`, not OpenAI policies or higher-priority instructions.
This documentation makes that distinction explicit and keeps the override
local to the prompt containing the prefix.

## Validation

- Automated checks: `tools/quality.ps1 -LowMemory` passed. This included
  formatting, PowerShell and viewer JavaScript syntax, module tidiness and
  integrity, third-party license inventory, Go tests, build, static analysis
  with 0 issues, vulnerability scan with 0 reachable vulnerabilities,
  tracked-artifact checks, and whitespace validation. The scan noted 4
  imported-package and 2 required-module findings that the code does not
  appear to call.
- Live validation: `not_observed`
- Capture or report: `none`

## Limitations and follow-up

The prefix cannot override system or developer instructions, applicable OpenAI
policies, tool restrictions, or other higher-priority requirements. It does
not independently authorize actions that the prompt does not request. No
runtime impact.
