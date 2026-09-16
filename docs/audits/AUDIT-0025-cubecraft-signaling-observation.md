# AUDIT-0025 - Record CubeCraft signaling observation

- Commit subject: `docs: record CubeCraft signaling observation`
- Change type: `documentation`
- Runtime impact: `no`

## Intent

Preserve a new user-provided Discord observation about CubeCraft's possible HTTP signaling path, the corrected RakNet status, and a redacted identity JWT without treating it as confirmed production behavior.

## Scope

Update `docs/research/cubecraft-gophertunnel-discord-audit.md` with the observation, the official NetherNet endpoint semantics used to interpret it, and the boundary between identity assertions and the current RakNet proxy path. No code, CLI behavior, capture schema, validation target, or policy changes.

## Evidence and provenance

- User-provided Discord excerpt and screenshot received on 2026-09-17. The screenshot's `cpk` value and token are not stored.
- Mojang's [NetherNet HTTP Signaling guide](https://mojang.github.io/bedrock-protocol-docs/guides/nether-net-onboarding-guide/) was checked for `GET /v1/join`, `POST /v1/join/{networkId}`, identity assertions, and `cpk` semantics.
- The Discord report is community evidence. It does not prove CubeCraft production NetherNet support, the corrected `mc-raknet-allowed` value, or a Sentinel detection rule.

## Validation

- Automated checks: `git diff --check`.
- Live validation: `not_observed`
- Capture or report: `none`

## Limitations and follow-up

No runtime impact. Do not implement NetherNet or change CubeCraft validation policy from this observation. Revisit only with an authorized production endpoint test or a primary CubeCraft or Mojang source that confirms the transport and identity requirements.
