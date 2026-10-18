# AUDIT-0227 - Preserve transfer dial route-hint classification

## Investigation and scope

This is a focused correction to FIX-0136. The explicit pre-reconnect probe sets Network.BestEffortPing, but connectUpstream performs another bounded optional status ping through a separate observing network. That network lacked the marker. Gophertunnel prepareDialAddress continues to the original address when its ping fails, so the diagnostic could incorrectly block validation after a successful child session.

The implementation applies the existing marker only when transferRoute is true. It retains ping error identity and raw response evidence. Initial-dial ping errors, connection failures, packet forwarding, authentication, and capture schema are unchanged. No server-specific exception or retry is added.

## Regression and paired consumer

TestConnectUpstreamBoundsAddressDiscoveryPing now checks the existing deadline and continuation behavior plus the captured best_effort marker, severity, and retained error for initial and transfer dials. The transfer case fails before the correction and passes afterward. The initial case remains an error.

BedrockPackArchive's transferred-hop synthetic fixture adds the retained route-hint warning before its offer. It continues to require archive digest, root and nested manifests, plaintext contents.json, three synthetic key declarations, source sequence, hop, and child path. The importer contract is unchanged, so no importer production code change is required. Both repositories keep separate commits. The paired consumer SHA and final validation results are recorded before publication.

## Validation status and human follow-up

Focused regression passed, including a failing transfer case before the fix. The canonical tools/format.ps1 and tools/quality.ps1 -LowMemory checks passed with zero linter issues and no reported vulnerabilities. The paired consumer tests, vet, and whitespace checks passed at `5f1eaf94456774f319b39acdb821d944ceece63c`. The clean stamped build, actual producer-verifier fixture check, and live check follow this commit. Live evidence from the preceding runtime does not validate this revision.

For a human validation, select Enchanted, choose an available game in its current menu, and allow the normal transfer. Expect the next-hop login, complete resource-pack offer, and spawn. If its optional status ping fails, retain the best_effort warning and operation error. A next-hop connection failure must still fail. Keep the closed capture and sanitized report if any behavior differs. Headless archive collection does not establish retail gameplay or satisfy the official release gate.
