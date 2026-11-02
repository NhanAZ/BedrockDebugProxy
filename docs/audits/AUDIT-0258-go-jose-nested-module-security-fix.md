# AUDIT-0258 - Close the nested go-jose security manifest gap

- Commit subject: `Fix nested go-jose dependency security alert`
- Change type: dependency security remediation
- Security alert: [Dependabot alert #18](https://github.com/NhanAZ/BedrockDebugProxy/security/dependabot/18)
- Advisory: [GHSA-mgcc-cww5-fp2q](https://github.com/go-jose/go-jose/security/advisories/GHSA-mgcc-cww5-fp2q)
- CVE: `CVE-2026-116587`

## Intent and scope

Dependabot alert #18 reported a high-severity denial-of-service vulnerability
in `github.com/go-jose/go-jose/v4` versions through `v4.1.5`. A crafted RSA
JWK could make RSA key validation consume excessive CPU. The upstream fix is
`v4.1.6`, which bounds accepted RSA modulus sizes before expensive validation.

PR #8 updated the root `go.mod` and `go.sum` to `v4.1.6`, but the replaced
`third_party/gophertunnel` module retained its own `v4.1.4` requirement and
checksums. The root build selected `v4.1.6`, but the nested manifest remained a
vulnerable dependency declaration and kept the Dependabot alert open.

## Implementation

- Updated `third_party/gophertunnel/go.mod` from `v4.1.4` to `v4.1.6`.
- Replaced the nested module's `go.sum` entries with the verified `v4.1.6`
  checksums.
- Added an Unreleased changelog entry documenting the security remediation.

No application authentication, protocol, capture, or packet behavior was
changed. No workaround or advisory dismissal was used.

## Dependency evidence

The selected root build list resolves only `github.com/go-jose/go-jose/v4`
`v4.1.6`. The module is reachable through the authentication and transport
dependency path used by the CLI. The nested gophertunnel manifest now declares
the same patched version.

## Verification

- `go list -m -json github.com/go-jose/go-jose/v4` selected `v4.1.6`.
- `go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 ./...` reported no
  vulnerabilities in called code.
- The prior merge commit's GitHub Quality and Windows compatibility workflows
  passed before this nested-manifest correction. The new commit must pass the
  same workflows after it is pushed.
- No live server session is required because this change only updates module
  metadata and checksums.

## Known limitation and follow-up

GitHub Dependabot rescans the default branch asynchronously after the push.
The alert should close automatically once the dependency graph observes both
module manifests at `v4.1.6`. If it remains open after the new default-branch
workflow completes, inspect the alert's dependency graph again before taking
any dismissal action.
