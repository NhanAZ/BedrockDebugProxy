# AUDIT-0244 - Make release validation agent-controlled

- Commit subject: `Make release validation agent-controlled`
- Change type: `process`
- Runtime impact: `no`

## Intent

Make the release workflow explicit that agents own the validation client and that a maintainer does not need to open Minecraft for the six-server release matrix. Use Golem CLI for applicable disposable local server and plugin scenarios.

## Scope

Update `AGENTS.md`, `CONTRIBUTING.md`, `README.md`, `CHANGELOG.md`, `docs/releasing.md`, `docs/validation.md`, `docs/decisions/0015-cli-release-validation.md`, `docs/server-targets.md`, `docs/golem.md`, and `validation/README.md`. The public release matrix remains an authenticated protocol-client workflow with the existing six servers, captures, report schema, Enchanted transfer checks, and CubeCraft risk acknowledgement. The repository Golem CLI remains a server-side PocketMine/Axolotl simulation and is documented as supplemental local evidence, not as a network client.

## Evidence and provenance

The current Golem CLI contract targets a private GolemAgent mailbox on a running development server. Its documented runtime skips Xbox authentication and real network login, and its journal is not accepted by the proxy capture verifier. The updated process separates this simulation evidence from public protocol-client reports while removing the unnecessary retail-client requirement.

## Validation

- Automated checks: documentation diff review and `git diff --check` are required; the existing runtime quality evidence remains recorded in `AUDIT-0243`.
- Live validation: `not_observed`
- Capture or report: `none`

## Limitations and follow-up

This policy change does not create a network-capable Golem client. A future network adapter would need its own source review, authentication boundary, capture compatibility, and release-gate decision before it could replace the reviewed protocol client. No retail rendering or gameplay claim is made.

## Security and privacy

Golem remains restricted to a disposable local server and private mailbox. The workflow continues to prohibit credentials, raw captures, resource-pack keys, decrypted assets, and private server discussions in source or sanitized reports.
