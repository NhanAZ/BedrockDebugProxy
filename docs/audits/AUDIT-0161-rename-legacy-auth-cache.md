# Audit AUDIT-0161 - Rename legacy authentication caches explicitly

- Commit subject: `Rename legacy authentication caches explicitly`
- Intent: let an operator replace the ambiguous legacy `default` label with a meaningful local account name without performing another Microsoft device login.
- Runtime impact: yes. A new local account-management command moves one cache file and updates the account registry. Packet forwarding, token contents, and capture formats are unchanged.

## Scope

- Add `rename OLD_ACCOUNT NEW_ACCOUNT` to the CLI.
- Move the source cache only when it exists and reject a destination that already has a profile.
- Preserve token bytes, local cache-age policy, and selected-default state.
- Roll back the file move if publishing the updated registry fails.
- Update the README, authentication guide, changelog, and fix history.

## Evidence and provenance

- The existing `AccountStore` registry and token-path validation remain the source of truth.
- The operation is local metadata migration. It does not inspect, decode, or rewrite OAuth token claims.
- No network, server, packet, capture, or authentication-provider behavior is changed.

## Automated validation

- `go test ./cmd/bedrock-debug-proxy ./internal/authcache`
- `git diff --check`
- Focused account-store tests cover legacy migration, selected-default preservation, token retention, and destination collision rejection.

## Live validation status

No external account or server session was used. Manual validation is limited to running `accounts`, executing `rename default <known-label>`, and confirming the named cache remains usable with an owner-controlled local server.

## Known limitations and follow-up

- The destination label is supplied by the operator and is not verified against the Microsoft identity or Minecraft username in the token.
- The command refuses to overwrite an existing destination. Use `logout` only after confirming which local cache should be retained.
