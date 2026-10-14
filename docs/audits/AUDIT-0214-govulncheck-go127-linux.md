# AUDIT-0214 - Update the scanner for Go 1.27 Linux analysis

- Commit subject: `build: update govulncheck for Go 1.27 Linux analysis`
- Change type: tooling
- Runtime impact: none

## Intent and scope

The initial paired push publishes producer `06344a964b1157951320c5b684228f9d221d7a56` and private consumer `8afee20944f45fad3faa545e7cdb13d923a26108`. Producer Windows CI and Dependency Graph succeed, but Linux quality fails in the old vulnerability scanner after race tests, build, and static analysis pass. Change the govulncheck pin from v1.1.4 to the reviewed stable v1.9.0. Keep the canonical gate, scan coverage, and error policy.

## Evidence and provenance

[The focused research note](../research/govulncheck-go127-linux.md) records the failed exact-revision CI, old-scanner Linux-target reproduction, immutable selected release, source review, and build-configuration boundary. This is an existing tool integration issue, not a reported product vulnerability. No upstream issue or pull request is created.

## Validation

- Old scanner v1.1.4 reproduces the Linux-target panic and exit 2 on the unchanged source.
- New scanner v1.9.0 completes the same Linux-target source analysis with exit 0 and zero reachable vulnerabilities. `tools/quality.ps1 -LowMemory` passes on Windows, including all tests, build, static analysis with zero issues, and the unchanged full vulnerability gate. The scanner reports one imported or required advisory outside called code.
- Fresh exact-revision Linux and Windows CI follow this commit.
- Consumer remote main matches `8afee20944f45fad3faa545e7cdb13d923a26108` and remains private. It has no configured Actions workflow or remote check run. Its prior local tests, vet, build, and paired verifier checks pass.

## Limits

Only developer tooling changes. Original live reports remain unchanged and the runtime revision is still `414eeb97157416eed8dfdc85f872101f40dd74af`. No retail gameplay or release is claimed. No paired consumer code or capture contract change is needed for a scanner pin.
