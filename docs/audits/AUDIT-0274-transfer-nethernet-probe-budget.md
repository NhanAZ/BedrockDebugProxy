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

- Automated checks: the full `tools/quality.ps1 -Race -LowMemory` gate passed on runtime revision `39b29f141b918ad0b00cab2cb90fbb862aff56af`. Hosted run [44881075259](https://github.com/NhanAZ/BedrockDebugProxy/actions/runs/44881075259) passed on Linux and Windows.
- Live validation: the stamped v0.6.2 proxy accepted a local Go Golem client, inspected the current Enchanted Hub form, and followed its Dragons selection. Hop 2 recorded a successful `nethernet-https` route probe with a one-minute bound and no fallback. The child exchanged 10 resource-pack archives and spawned before the client closed. The Hub and child offers contained 11 archives total, all marked unencrypted.
- Capture: [session-1](https://github.com/BedrockDebugProxy-Captures/5568f83e-c3e5-429e-a0b4-43c047f0e8d5/releases/tag/capture-archive). The private archive is keyed by its unique manifest capture ID because the event-level `session_id` restarts at `session-1` for separate processes. The capture closed with 38,070 events and 17,060 blobs, zero write errors, dropped events, truncations, or decode errors, and a complete artifact index. The capture verifier passed.
- Paired consumer: BedrockPackArchive commit [e0c08f4](https://github.com/NhanAZ-Tools/BedrockPackArchive/commit/e0c08f4a2f22cc40c2382377bf8f3b62d13bb48f) closes Enchanted child clients promptly after spawn. The exact stamped producer compatibility fixture passed against the v0.6.2 binary.

## Limitations and follow-up

This focused route check does not replace the six-server release matrix. The six-server gate passed separately on the same runtime revision and is recorded in AUDIT-0275. Do not infer server unavailability from a failed timeout alone.
