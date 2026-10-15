# AUDIT-0221 - Prepare v0.4.0 release candidate

- Commit subject: `docs: prepare v0.4.0 release candidate`
- Change type: Markdown-only release preparation
- Runtime impact: none. The validated runtime revision is `aa103ac5d88ef4ca24a9e24f06789a60b8490864`.

## Scope and validation

Move the reviewed Unreleased changelog entries into v0.4.0 and record the exact-revision release findings. The six sanitized retail reports for The Hive, Galaxite, Lifeboat, Mineville Zeqa, Enchanted, and CubeCraft pass at the validated runtime revision. The Enchanted session followed a minigame transfer to hop 2. The CubeCraft session used the normal read-only proxy route after the maintainer acknowledged its published proxy-policy risk. Neither acknowledgement nor a successful test is a policy exception.

The runtime commit passed the local canonical quality gate and both jobs of [CI run 38701214802](https://github.com/NhanAZ/BedrockDebugProxy/actions/runs/38701214802). The exact-runtime release-readiness gate passed all six reports. Before tagging the Markdown-only candidate, rerun the quality, build, CI, and documentation-only equivalence checks at its candidate revision. The reports keep their original `tested_revision` and remain ignored locally until attached as sanitized release assets.

## Limits

The retail sessions exercise the named server flows with Bedrock 1.26.52 and protocol 2193. They do not prove compatibility with every server, client build, or future protocol. Raw captures, credentials, and third-party resource packs are excluded from the release assets and source history.
