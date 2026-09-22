# Audit AUDIT-0159 - Add low-friction Microsoft account profiles

- Commit subject: `Add low-friction Microsoft account profiles`
- Intent: allow more than one local Microsoft authentication cache without adding an account flag to every proxy run, while keeping the existing one-account workflow unchanged.
- Runtime impact: yes. `run --auth device` now resolves the selected default account through a local registry. Device authentication, token refresh, upstream login, packet forwarding, and capture formats remain unchanged.

## Scope

- Added `internal/authcache.AccountStore` with a versioned per-user registry, stable account labels, default selection, legacy-cache compatibility, local cache-age policy, and safe token paths.
- Added top-level `login`, `accounts`, `use`, and `logout` commands. The first successfully authenticated account becomes the default. `login --max-age` only controls local cache reuse and does not change provider-issued token expiry.
- Added an explicit forced-device option for intentional reauthentication. Existing cached refresh behavior remains the default for ordinary runs, and an unchanged cached token is not rewritten so a local cache-age policy does not reset on every start.
- When `login` is interactive and no flag is supplied, a single prompt accepts a local cache age. Enter selects no local limit. Non-interactive login defaults to no local limit without blocking.
- Updated the README, authentication guide, validation and release instructions, legal behavior description, changelog, and fix history.

## Evidence and provenance

- Existing `internal/authcache` token serialization and refresh behavior was reused. No token value, account claim, or authentication chain is copied between profiles.
- Microsoft lifetime terminology and links in [`docs/authentication.md`](../authentication.md) were reviewed against the official Microsoft Learn access-token, refresh-token, and device-code documentation on 2026-09-22.
- Account labels are local metadata only. Registry keys are validated before constructing paths, and token and registry files retain user-only permissions where supported.

## Automated validation

- `go test ./internal/authcache ./cmd/bedrock-debug-proxy`
- `git diff --check`
- Focused coverage includes legacy cache compatibility, named-account creation and switching, default replacement after logout, unsafe registry rejection, local-age parsing and expiry, unchanged-cache timestamp preservation, forced device selection, interactive and non-interactive cache-policy selection, and flexible `login ACCOUNT --max-age VALUE` argument order.

## Live validation status

No device login or external server session was performed for this change. Manual validation remains required with an owner-controlled local server: authenticate two explicitly authorized accounts, confirm `accounts` and `use`, run once with each default, then log out one profile and verify the other remains usable. Do not use this feature to bypass a server policy or enforcement decision.

## Known limitations and follow-up

- The provider controls access-token expiry, refresh-token rotation, and revocation. The local `max-age` setting is a reauthentication policy, not a token-lifetime override.
- The local account registry does not discover or verify the Minecraft username associated with a label. The operator should choose clear labels and inspect the provider's login result.
- Existing legacy caches are exposed as the synthetic `default` label until a named login creates the registry. No migration or token duplication is performed.
