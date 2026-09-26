# Audit AUDIT-0197 - Allow packs-only smoke checks for development pull requests

- Commit subject: `docs: scope pull request validation to packs-only smoke checks`
- Change type: `documentation`
- Runtime impact: `no`

## Intent

Use the existing packs-only mode as scoped live evidence for routine development pull requests, while retaining full retail-client testing for official releases.

## Scope

- Update `AGENTS.md`, `README.md`, `CONTRIBUTING.md`, `docs/validation.md`, `docs/resource-packs.md`, and `validation/README.md` to distinguish the pre-spawn packs-only smoke check from full client validation.
- Clarify that packs-only does not prove a retail client can spawn or play, and require focused automated evidence for flows outside its boundary.
- No product code, CLI behavior, or report-tool behavior changed.

## Evidence and provenance

The existing `run --packs-only` help and implementation state that the mode performs an upstream login and resource-pack exchange, uses default Gophertunnel client metadata, and closes without a listener or world spawn. The existing validation-report tool supports `result=incomplete` and `normal_session=not_observed`; release readiness continues to require passing reports.

## Validation

- Automated validation: `tools/format.ps1` passed, `tools/quality.ps1` passed (PowerShell and JavaScript syntax, module integrity, license inventory, unit tests, build, static analysis, vulnerability scan, and tracked-capture checks), and `git diff --check` passed. The vulnerability scan reported no vulnerabilities in this project and noted six dependency vulnerabilities not called by this code.
- Live validation: not applicable.
- Capture or report: none.

## Limitations and follow-up

Packs-only smoke evidence is limited to the observed upstream login and pack flow. It cannot establish retail-client compatibility, world spawn, or gameplay. Those checks remain part of the official six-server release gate.
