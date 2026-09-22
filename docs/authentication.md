# Authentication accounts

BedrockDebugProxy keeps Microsoft authentication in per-user local cache files. The normal `run` command uses one selected default account, so everyday proxy commands do not need an account flag.

## First login

The first interactive authentication must have a local account label. This prevents a new installation from silently assigning the token to the synthetic `default` profile. Start the proxy normally, or run the login command directly:

```powershell
.\bin\bedrock-debug-proxy.exe login
```

When no account exists, the terminal asks for the Microsoft account username or a local label. Entering a non-empty label is required and the first successful account becomes the default. If accounts already exist, pressing Enter at this prompt reauthenticates the selected default instead.

The usual proxy command remains short after the first login:

```powershell
.\bin\bedrock-debug-proxy.exe run --upstream "<UPSTREAM>" --auth device
```

The provider refreshes the cached Microsoft token when needed. A failed refresh starts device authentication again.

## Multiple accounts

Use a local label only when adding or switching accounts:

```powershell
# Add or reauthenticate an account.
.\bin\bedrock-debug-proxy.exe login NhanAZ

# See the selected account and cache state.
.\bin\bedrock-debug-proxy.exe accounts

# Select the account used by future run commands.
.\bin\bedrock-debug-proxy.exe use NhanAZ

# Rename a legacy default cache without reauthenticating it.
.\bin\bedrock-debug-proxy.exe rename default SampleAccount

# Sign out the default account, or name another account explicitly.
.\bin\bedrock-debug-proxy.exe logout
.\bin\bedrock-debug-proxy.exe logout OtherAccount
```

`login` without a label prompts for a label in an interactive terminal. A non-empty label adds or reauthenticates that profile. Pressing Enter is allowed only after at least one account exists, and reauthenticates the current default. The first successful account becomes the default. `use` changes only the local selection. `logout` removes only the selected account's local token cache. If the default is removed while another account remains, the remaining account with the stable first label becomes the default.

`rename OLD_ACCOUNT NEW_ACCOUNT` moves an existing local cache and preserves its token, local cache-age policy, and selected-default state. It never overwrites an existing destination. This is useful for converting the legacy `default` cache to a meaningful label without signing in again. The label is local metadata and is not verified against the Microsoft or Minecraft identity in the token.

Account labels are local names, not claims about the Minecraft username inside the token. The account registry never stores access or refresh token values.

## Local cache age

When `login` runs in a terminal without `--max-age`, it asks for a local cache age. Press Enter for no local limit, or enter a value such as `7d`. This is the low-friction default. Non-interactive login uses no local limit automatically so scripts do not wait for input.

An optional local maximum cache age can require device authentication again even if the provider could still refresh the token:

```powershell
.\bin\bedrock-debug-proxy.exe login NhanAZ --max-age 7d
```

The flag is useful for scripts or for skipping the prompt. The accepted forms are a positive Go duration such as `168h` or a day value such as `7d`. `none` and `0` remove the local limit. This setting does not change Microsoft's token lifetime. Microsoft controls the access token and refresh token expiry, rotation, and revocation. The proxy only applies the local policy when choosing whether to reuse the cache.

## Storage and privacy

On Windows, the legacy single-account cache is `%AppData%\BedrockDebugProxy\auth-token.json`. Additional named accounts use `%AppData%\BedrockDebugProxy\accounts\<label>.json`, and the local selection is in `%AppData%\BedrockDebugProxy\accounts.json`. Existing legacy caches continue to work as the `default` account without a migration prompt. A new installation with no cache cannot create an unnamed `default` profile through an interactive first login.

These files contain Microsoft authentication secrets. They are written with user-only permissions where the platform supports them, are ignored by Git, and must not be copied into captures, issues, logs, or commits. `logout` removes the local cache but does not revoke the Microsoft session or sign out other Microsoft applications.

Microsoft token lifetime references reviewed on 2026-09-22:

- [Access tokens](https://learn.microsoft.com/en-us/entra/identity-platform/access-tokens)
- [Refresh tokens](https://learn.microsoft.com/en-us/entra/identity-platform/refresh-tokens)
- [OAuth 2.0 device authorization grant](https://learn.microsoft.com/en-us/entra/identity-platform/v2-oauth2-device-code)
