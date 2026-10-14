# AUDIT-0210 - Update Go security maintenance versions

- Commit subject: `build: update Go and networking security patches`
- Change type: tooling
- Runtime impact: yes

## Intent and scope

The clean baseline's canonical vulnerability scan fails with Go 1.27.0. Select the current stable Go 1.27.2 and x/net v0.60.0 with their required graph in the root and replacement gophertunnel modules. Update the released golangci-lint pin to v2.14.0 because v2.13.2 cannot read the new compiler's export format. The dependency and checker review is separate from the subsequent Bedrock metadata update. The same networking graph is selected in BedrockPackArchive.

## Evidence and provenance

[The maintenance review](../research/maintenance-2026-10-15.md) records the baseline failure, official advisory, stable release evidence, retained source provenance, and the unchanged capture and archive contract. No dependency source is copied and no checker is disabled.

## Validation

- `tools/format.ps1` and `tools/quality.ps1 -LowMemory` pass with Go 1.27.2 and golangci-lint v2.14.0. Static analysis reports zero issues and govulncheck reports zero reachable vulnerabilities. One imported or required vulnerability remains outside called code. The previous Go 1.27.0 vulnerability failure and v2.13.2 export-format failure are recorded in the research note.
- BedrockPackArchive `go test -count=1 ./...`, `go vet ./...`, build, and `git diff --check` pass at consumer commit `816d23aba51f95c5799a256389cf22c5526d9087`. The actual stamped producer verifier is pending.
- Replacement gophertunnel minecraft, NBT, and packet tests and vet pass. Replacement go-raknet `go test -race -count=1 ./...` passes.
- Public-server checks for this committed revision are pending. Retail-client gameplay is not observed.

## Limitations

The Go toolchain directive selects a minimum preferred toolchain under Go's automatic selection. A newer installed Go version may be used. The quality gate remains responsible for detecting reachable vulnerabilities in the actual compiler used. No official release or remote publication is performed by this commit.
