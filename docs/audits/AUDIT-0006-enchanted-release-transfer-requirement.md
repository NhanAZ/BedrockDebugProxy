# Audit 0006 - Require an Enchanted minigame transfer in release validation

- Commit subject: `Require Enchanted minigame transfer for release validation`
- Change type: Documentation and validation process
- Runtime impact: None. No product binary or protocol behavior changed.
- Intent: Make the Enchanted release gate exercise the real form-driven minigame transfer instead of accepting a hub-only session.
- Scope: `docs/validation.md`, `docs/releasing.md`, `docs/server-targets.md`, and `validation/README.md`.
- Evidence: The maintained Enchanted validation flow receives a server-sent form when the player joins, uses the form response to select a minigame, emits `Transfer`, and requires a next-hop resource-pack exchange before minigame spawn. The requirement is based on the project's existing Enchanted transfer research and live validation history.
- Pass criteria: After the server sends its form, select any available minigame, observe the `Transfer`, follow the next hop, complete its resource-pack exchange when offered, reach minigame spawn, and record `transfer_following=pass` plus `resource_pack_transfer=pass`.
- Automated checks: `git diff --check`; `tools/quality.ps1`.
- Live validation: Not applicable to this documentation-only change. The next Enchanted release report must use the updated flow.
- Follow-up: Treat a hub-only Enchanted session as incomplete for release readiness.
