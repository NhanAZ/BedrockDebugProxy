# AUDIT-0274 - Bound HTTPS NetherNet transfer probes by the hop budget

- Change type: `runtime`
- Runtime impact: `yes`

## Intent

Allow a valid HTTPS NetherNet transfer route to finish discovery when its endpoint needs more than the former fixed 10 seconds, while preserving a finite wait and the configured shorter per-hop budget.

## Root cause

An Enchanted lobby `Transfer` to Dragons reached `tutorial.wyville.connect.enchanted.gg:19132`. The proxy's HTTPS `/v1/join` route probe expired after its hard-coded 10 seconds, then fell back to RakNet. An independent Golem session entered through the Enchanted lobby, selected Dragons from the live form, followed that same transfer, spawned, and received 10 child packs. This classifies the proxy's short route-probe timeout as the cause. It does not indicate a Dragons maintenance state.

## Scope

The HTTPS NetherNet route probe now uses the configured upstream dial budget, capped at one minute. The caller context can still impose an earlier deadline. Successful and failed probe events retain the effective timeout. No transport fallback behavior changed.

## Validation

- Automated checks: focused route-probe regression tests and the full `tools/quality.ps1` gate pass. The exact stamped build is pending.
- Live validation: Golem successfully followed the current Enchanted form's Dragons transfer and received all offered child packs. Exact-revision proxy validation is pending.
- Capture or report: no capture is cited. The troubleshooting capture includes decrypted third-party pack content and is not suitable for a project evidence archive.

## Limitations and follow-up

Run the stamped build against Enchanted, select Dragons from its current server-sent form, and verify the proxy selects HTTPS NetherNet, completes the child pack exchange, reaches spawn, and closes a verified capture. Do not infer server unavailability from a failed timeout alone. This focused check does not satisfy the six-server official release gate.
