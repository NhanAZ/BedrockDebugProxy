# Audit AUDIT-0078 - Record post-fix vanilla-client loopback evidence

## Intent

Record the completed local validation that followed the offline identity-field fix. This is a documentation-only audit and has no runtime impact.

## Scope

- Updated `docs/research/local-identity-parity.md` with the real Windows Bedrock client loopback result from the rebuilt binary.
- Recorded the closed ProxyFingerprint report's safe summary without copying identifiers, raw tokens, skins, or packet payloads.

## Evidence and provenance

- The rebuilt `bin/bedrock-debug-proxy.exe` served an operator-controlled PocketMine server on loopback.
- The session completed login, resource-pack negotiation, spawn, gameplay traffic, and a normal client quit over approximately 220.9 seconds.
- The closed local report contained 4,350 packets and no `auth_client_playfab_mismatch` signal. It only reported the expected loopback and self-signed-authentication signals.

## Automated validation

- No runtime code changed. The evidence was read from the closed local report and the proxy log.
- `git diff --check` is required before committing this documentation-only change.

## Live validation

The validation used only the operator-controlled local PocketMine server. No public server was contacted and no anti-cheat or authentication policy was bypassed.

## Known limitations and follow-up

This evidence confirms local offline identity parity for a real vanilla client. It does not establish how any public server classifies a proxy, and it does not replace an authorized authenticated-server test.
