# AUDIT-0238 - Allow bounded CI startup time for the viewer test

The Windows compatibility job in [Quality run 39450934911](https://github.com/NhanAZ/BedrockDebugProxy/actions/runs/39450934911) failed the new JavaScript rendering regression after 10.04 seconds, with process exit status 1 and no child output. This matches its ten-second deadline. The log does not establish whether Node startup or stdin handling caused the delay.

The test now uses the one-minute subprocess deadline already used by other executable-based tests, inherits the test context, and includes the context error in failure diagnostics. Its embedded source, input, function extraction, and assertions are unchanged. Production code and the capture contract are unchanged.

The focused regression passed 30 consecutive local Windows runs in 1.692 seconds. `tools/format.ps1` and `tools/quality.ps1 -Race` passed, with zero static-analysis findings and no vulnerabilities in called functions. Exact-commit CI remains required. The hosted delay did not reproduce locally.

Earlier scoped sessions for runtime `34472e23c379bc68d289a9f2de08580c99d8c3e5` passed on [The Hive](https://github.com/BedrockDebugProxy-Captures/session-20261020T153727Z/releases/tag/capture-archive) and [Galaxite](https://github.com/BedrockDebugProxy-Captures/session-20261020T154008Z/releases/tag/capture-archive). Both private archives have verified SHA-256 digests and retain their original revision. Their reports remain unchanged. This test-source change is outside the documentation-only release allowlist, so the candidate requires a fresh six-server matrix.
