# AUDIT-0228 - Record paired archive-client validation

## Intent and scope

This documentation-only record completes the maintenance evidence for resource stream failures, deferred upstream pack completion, and transfer dial route hints. It changes no runtime, capture schema, protocol definition, viewer asset, or pack byte.

The validated producer is 0.4.1-dev at `f0d8efe894a3acd5031efb7d968cf660b9820018`. The paired private BedrockPackArchive main is `8d219dd83b66bc1a81645e04e18635395c9fe498`. Its immediately preceding runtime implementation is `36356657acd8e91f32008835920469f3bbc24211`. The final consumer commit changes documentation only. The preceding paired consumer fixture commit is `5f1eaf94456774f319b39acdb821d944ceece63c`.

## Automated and live evidence

The producer's canonical formatting and LowMemory quality gate passed for the runtime revision, including focused transfer-dial and completion-order regressions. The consumer's full tests, vet, build, whitespace check, synthetic navigation and shutdown tests, and actual stamped producer-verifier importer matrix passed. The fixtures preserve archive digests, initial and transferred provenance, root and nested manifests, plaintext contents.json, and synthetic per-file key declarations. Real captures and keys are excluded from the fixtures.

The stamped producer completed authenticated headless world startup for The Hive in [session-20261018T185114Z](https://github.com/BedrockDebugProxy-Captures/session-20261018T185114Z/releases/tag/capture-archive). Its sanitized machine-readable report passes the automatic evidence requirements and remains incomplete because normal_session is not_observed. Matching scoped reports were generated for CubeCraft, Galaxite, Lifeboat, Mineville Zeqa, and Enchanted. None is relabeled as retail-client validation or an official release pass.

The Lifeboat and Enchanted scoped reports also retain failed transport writes during connection shutdown and therefore do not pass every automatic requirement. Their completed pack exchange and closed capture do not override that conservative report result. The other four scoped reports pass the automatic requirements. No write-error exclusion or report fact is changed to obtain a passing gate.

Enchanted completed a current child startup in [session-20261018T183325Z](https://github.com/BedrockDebugProxy-Captures/session-20261018T183325Z/releases/tag/capture-archive). Optional route-probe failures remain observable, while the actual connection and pack exchange succeed. The new consumer implementation completed the Mineville DUELS transfer and child startup in [session-20261018T191642Z](https://github.com/BedrockDebugProxy-Captures/session-20261018T191642Z/releases/tag/capture-archive). Its current walk has zero prediction corrections. This is a separate normal loopback client. The proxy does not replay or inject old client packets.

BEDWARS received an upstream startup refusal after retaining its offer in [session-20261018T191827Z](https://github.com/BedrockDebugProxy-Captures/session-20261018T191827Z/releases/tag/capture-archive). The capture remains failed, the consumer reports failure and continues later routes, and that failed child source is excluded from new imports. No endpoint bypass, server-specific protocol workaround, or suppressed error is added.

Each linked capture is an unchanged complete private release archive. Repository privacy and the uploaded asset digest were verified against the prepared ZIP and README SHA-256. The maintainer explicitly authorized this batch's private capture archival, including derived packs, and accepted the CubeCraft connection risk for this batch.

## Limitations and follow-up

The authenticated client reached five reviewed Mineville routes. The combined six-route command still fails because BEDWARS does not spawn. Dimension Clash's new-account tutorial remains outside the implemented consumer navigator, while the maintainer's completed form-driven 3vs3 capture remains a distinct legacy source. Initial offers and world startup do not prove exhaustive child-mode coverage, rendered gameplay, or movement indistinguishable from human input.

For human validation of these producer changes, enter The Hive for the ordinary baseline, then use Enchanted's current game menu and allow the server transfer. Observe downstream pack completion before upstream completion, child login, complete offered archives, and world spawn. A transfer ping may remain a retained best-effort warning. A failed child connection must still fail. Keep the closed capture and sanitized report if the flow differs. The six-server retail gate remains required before any official BedrockDebugProxy release. No producer release is made by this task.
