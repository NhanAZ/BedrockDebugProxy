# AUDIT-0063 - The Hive proxy-ban investigation record

- Commit subject: `docs: record The Hive proxy-ban investigation`
- Change type: `documentation`
- Runtime impact: `no`

## Intent

Record the evidence and policy boundary for the 2026-09-20 The Hive `Unfair Advantage` ban investigation. Correct an earlier overconfident conclusion that the `SrDuck84` upstream identity could be excluded, while avoiding an unsupported claim about Hive's private detector.

## Scope

Add a research note for the affected and comparison captures, link the investigation from `docs/fix-history.md`, and make the Hive authorization requirement explicit in `docs/server-targets.md` and `docs/validation.md`. No source, protocol model, capture, credential, or runtime behavior changes.

## Evidence and provenance

The note records the closed captures [session-20260920T135412Z](https://github.com/BedrockDebugProxy-Captures/session-20260920T135412Z/releases/tag/capture-archive) and [session-20260917T034859Z](https://github.com/BedrockDebugProxy-Captures/session-20260917T034859Z/releases/tag/capture-archive), the 23-capture search for `SrDuck84`, the runtime-path comparison between their generator commits, Hive's official VPN/proxy and client-modification policies, Hive's supported-version page and public changelog, and Mojang's protocol changelog. Sensitive identifiers and payloads are intentionally omitted. Confirmed facts are separated from hypotheses and unresolved questions.

## Validation

- Automated checks: `git diff --check` and targeted documentation searches are required before commit.
- Live validation: `not_observed`. This documentation change does not authorize a new Hive connection.
- Capture or report: the cited captures are stored in private capture repositories and remain sensitive evidence.

## Limitations and follow-up

The exact Hive detector and the causal role of the `SrDuck84` identity remain unknown. Public changelog entries do not prove that no private detector or blocklist rollout occurred. No runtime fix is proposed. Future live validation requires current policy review, explicit authorization, and a safe test environment.

No runtime impact.
