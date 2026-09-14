# Audit 0005 - Exclude a prohibited proxy target from active validation

- Commit subject: `Exclude prohibited proxy target from release validation`
- Change type: Documentation, tooling, and agent process
- Runtime impact: None. No product binary or protocol behavior changed.
- Intent: Remove CubeCraft from the required pull request and release validation matrix, document the remaining server endpoints, and state the server-policy risk for proxy use.
- Scope: `AGENTS.md`, `README.md`, `CONTRIBUTING.md`, `docs/validation.md`, `docs/releasing.md`, `docs/session-artifacts.md`, `validation/README.md`, `docs/fix-history.md`, `docs/research/enchanted-transfer-handshake.md`, `tools/check-release-readiness.ps1`, and `docs/server-targets.md`.
- Evidence: CubeCraft's [Rules](https://help.cubecraft.net/en/article/cubecraft-rules-1403lij/) and [Allowed Mods and Clients](https://www.cubecraft.net/threads/allowed-mods-and-clients.228596/) both state that proxies in Minecraft Bedrock are not allowed. Pages reviewed 2026-09-14.
- Policy boundary: Historical CubeCraft research remains available for provenance, but the agent must not recommend or require CubeCraft in active testing. If the maintainer asks for a CubeCraft test, the agent should warn about the published policy and possible rejection, kick, or ban. No product-level warning was added because the request explicitly keeps this internal to the agent workflow.
- Automated checks: `git diff --check`; `tools/quality.ps1`.
- Live validation: Not applicable. This change does not alter runtime behavior.
- Follow-up: Keep the five-server matrix revision matched and recheck endpoint references when a service changes its routing or policy.
