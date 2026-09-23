# AUDIT-0179 - Define the `--force` project workflow override

- Commit subject: `Define force prompt override`
- Change type: `process`
- Runtime impact: `no`

## Intent

Define a prompt prefix that lets the maintainer override repository-level agent
workflow rules for one request while preserving higher-priority platform and
tool constraints.

## Scope

Update `AGENTS.md` and add `docs/decisions/0012-force-prompt-project-workflow-override.md`.
No product code, CLI behavior, capture format, or release artifact changes.

## Evidence and provenance

The maintainer explicitly clarified that `--force` should override rules
established by the project's `AGENTS.md`, not OpenAI policies or higher-priority
instructions. The behavior is documented with that scope and applies only to
the prompt where the prefix appears.

## Validation

- Automated checks: `tools/quality.ps1 -LowMemory` passed. This included Go tests, build, static analysis with 0 issues, vulnerability scan with 0 reachable vulnerabilities, module and license checks, tracked-artifact checks, and whitespace validation. The scan noted 4 imported-package and 2 required-module findings that the code does not appear to call.
- Live validation: `not_observed`
- Capture or report: `none`

## Limitations and follow-up

The prefix cannot override system or developer instructions, applicable OpenAI
policies, or tool restrictions. It does not independently authorize actions
that the prompt does not request. No runtime impact.
