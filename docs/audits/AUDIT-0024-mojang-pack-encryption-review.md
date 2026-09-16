# AUDIT-0024 - Review Mojang pack encryption proposal

- Commit subject: `docs: record Mojang pack encryption review`
- Change type: `documentation`
- Runtime impact: `no`

## Intent

Record the findings from Mojang `bedrock-protocol-docs` pull request #63 before changing the resource-pack decryption algorithm.

## Scope

Update `docs/research/resource-pack-encryption.md` with the exact contributor revision, the proposed 26.50 stream-asset mode, the uninterpreted header bytes, the content-identity distinction, and the decision to defer code changes until the proposal is accepted or independently corroborated. No production code, capture schema, CLI behavior, or changelog entry changes.

## Evidence and provenance

- Mojang pull request [#63](https://github.com/Mojang/bedrock-protocol-docs/pull/63), open when reviewed on 2026-09-17.
- Contributor revision `2288c5ba945786a90f7bf78717162cee883e8ce7`, file `additional_docs/PackEncryption.md`.
- Current implementation in `internal/resourcepack/decrypt.go` and existing synthetic decryption tests.
- The proposed `AES-256-CTR` stream-asset behavior and header interpretation are documented as pending facts, not accepted protocol requirements.

## Validation

- Automated checks: `git diff --check`; focused resource-pack tests pass before this documentation-only change.
- Live validation: `not_observed`
- Capture or report: `none`

## Limitations and follow-up

No runtime impact. Revisit the decoder after PR #63 is merged or an independent primary source confirms the behavior. If confirmed, add a narrow CTR implementation, mode-aware tests, a relaxed header check, and advertised content-identity validation.
