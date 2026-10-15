# Audit AUDIT-0165 - Show the selected account in run output and the viewer

## Intent

Identify the selected local Microsoft account label in authenticated run output and in the capture viewer overview after multiple account support was added.

## Scope

- Record `options.values.account_name` in authenticated capture manifests.
- Print the selected account label when `run --auth device` starts.
- Add an Account metric to the viewer header and load it from the manifest.
- Update capture, authentication, fix-history, and changelog documentation.

No packet forwarding, authentication protocol, token storage, or identity verification behavior changed.

## Evidence and provenance

The account label comes from the local account registry already selected by `run`. It is operator metadata only and must not be interpreted as a verified Minecraft identity. The capture format and authentication documentation state this boundary explicitly.

## Automated validation

- Focused viewer tests cover the manifest value and Account header marker.
- `git diff --check` passed.
- `powershell -NoProfile -ExecutionPolicy Bypass -File tools/quality.ps1 -LowMemory` passed, including viewer JavaScript syntax validation, all tests, build, static analysis, and the vulnerability scan.
- A local `run --auth device` smoke check printed the selected account label before listening.

## Live validation

Pass via a clean committed build at `[local viewer URL]` using a local closed capture with synthetic `account_name=ExampleAccount` metadata. The Account metric rendered `ExampleAccount`, and the browser console had no errors or warnings. No public-server or real-client validation is implied by this presentation-only change.

## Known limitations and follow-up

Unauthenticated runs display `-`. The displayed label does not prove which Minecraft identity a provider token represents. Rebuild any release candidate from the final commit before release validation.
