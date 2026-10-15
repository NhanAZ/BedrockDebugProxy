# AUDIT-0219 - Connect HTTPS NetherNet Experiences

- Commit subject: `fix: connect HTTPS NetherNet Experiences`
- Change type: runtime networking
- Runtime impact: yes. `experience:` can now select the HTTPS NetherNet signaling endpoint advertised by Minecraft services.

## Intent and scope

The v0.4.0 CubeCraft release check failed before listening because the authenticated join result declared `NetherNet` and supplied a hostname and port rather than `netherNetId`. The existing resolver accepted only the ID-based NetherNet signaling paths. This change handles that response shape through the installed go-nethernet HTTPS endpoint client. It checks `/v1/join` before accepting the target, retains existing RakNet and ID-based signaling behavior, and does not downgrade to HTTP or use a guessed fallback endpoint. The public `run` syntax and capture schema are unchanged.

## Evidence and provenance

The current authenticated service response was inspected only for field names and value shapes. Its advertised host and port returned `200` for the HTTPS NetherNet status endpoint and `426` for plaintext HTTP. No token, endpoint value, or private server message is retained in this record. The [routing research note](../research/cubecraft-experience-routing-2026-10-15.md) classifies this observation. The [Mojang NetherNet onboarding guide](https://github.com/Mojang/bedrock-protocol-docs/blob/main/additional_docs/NetherNetOnboardingGuide.md), inspected on 2026-10-15, describes the HTTPS status and SDP exchange. The installed [`go-nethernet` v1.0.24 endpoint client](https://github.com/df-mc/go-nethernet/blob/v1.0.24/endpoint/client.go) already implements that exchange. This commit uses its public API and copies no upstream source.

## Validation

- Focused Experience destination and HTTPS status tests: pass. Existing RakNet, legacy NetherNet, and JSON-RPC fixtures remain covered.
- Development-build live resolution: pass. The authenticated CubeCraft selector reached the listener after the HTTPS status check.
- Development-build `--packs-only` session: authenticated upstream login and the offered resource-pack exchange completed. The capture closed and verified; no world spawn was requested.
- Canonical `tools/quality.ps1 -LowMemory` and `git diff --check`: pass. The gate completed formatting, PowerShell and JavaScript syntax, module checks, full tests, build, static analysis with zero issues, and a vulnerability scan with zero reachable findings. One imported or required advisory was outside called code.
- Retail client join, gameplay, and all six exact-revision release reports: pending a clean stamped build.

## Limitations and follow-up

The HTTPS endpoint is selected only for a `NetherNet` response with a host and port and no ID. The resolver does not infer HTTP support, alter the server's selected transport, or bypass its authentication or identity checks. A packs-only session proves the upstream login and pack boundary, not retail spawn. Validate The Hive and CubeCraft with a real client at the stamped runtime revision, then complete the six-server release matrix before publishing v0.4.0.
