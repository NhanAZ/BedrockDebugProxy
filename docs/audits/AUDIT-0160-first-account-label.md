# Audit AUDIT-0160 - Require a first account label

- Commit subject: `Require a first account label`
- Intent: prevent a new interactive installation from silently assigning its first Microsoft token to the synthetic `default` profile.
- Runtime impact: yes. First interactive `login` and first interactive `run --auth device` now require a non-empty local account label. Existing account selection and token handling remain unchanged.

## Scope

- Prompt for a Microsoft account username or local label when no account is configured.
- Reject an empty first label and return an actionable error for non-interactive first runs.
- Preserve Enter-for-default behavior for unlabeled login after at least one account exists.
- Reuse one buffered input reader across the label and cache-age prompts.
- Update the README, authentication guide, changelog, and fix history.

## Evidence and provenance

- The account registry and token cache implementation from `958cef0` remain the source of account storage and provider behavior.
- Labels are local metadata only. The prompt does not inspect or assert the Minecraft username contained in the Microsoft token.
- No token, capture, packet, or server behavior is changed by this workflow adjustment.

## Automated validation

- `go test ./cmd/bedrock-debug-proxy ./internal/authcache`
- `git diff --check`
- Focused tests cover required first-label rejection and preserving the next cache-age line through the shared prompt reader.

## Live validation status

No external account or server session was used for this change. Manual validation remains required in an interactive terminal with an owner-controlled local server: start from an empty account registry, confirm the required label prompt, complete device authentication, verify the named default with `accounts`, and repeat an unlabeled login to confirm Enter selects that default.

## Known limitations and follow-up

- Existing legacy `%AppData%\\BedrockDebugProxy\\auth-token.json` caches remain exposed as `default` for compatibility and are not renamed automatically.
- Microsoft account identity and Minecraft username are not discovered or verified by the local label prompt.
