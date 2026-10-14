# Govulncheck compatibility with Go 1.27 Linux analysis

## Reproduction and classification

[Quality run 38614252382](https://github.com/NhanAZ/BedrockDebugProxy/actions/runs/38614252382) tests producer revision `06344a964b1157951320c5b684228f9d221d7a56`. Windows tests and build pass. Linux formatting, module checks, license inventory, the full race-enabled test suite, build, and static analysis pass. The subsequent govulncheck v1.1.4 invocation panics with `unexpected expr: *ast.KeyValueExpr` in `golang.org/x/tools/go/ssa.(*builder).expr0`, using x/tools v0.29.0 and Go 1.27.2.

This is a toolchain integration incompatibility. The failure is not a vulnerability finding and does not establish a runtime defect in BedrockDebugProxy. The precise Linux-selected function that triggers the unsupported SSA expression has not been isolated. Do not infer that all targets fail. The same scanner passed the prior Windows-only local gate.

A Windows-built v1.1.4 scanner reproduces the same panic when analyzing this unchanged module with GOOS=linux under Go 1.27.2. The source scan exits 2. The scanner binary itself remains a Windows executable. Only its package analysis target changes, so no Linux executable is run on Windows.

## Selected source and scope

The latest released [golang.org/x/vuln v1.9.0](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck@v1.9.0) is published on 2026-10-11. Its tag resolves to `e672cbe8fbc8c76e28fcb6db0efd9ce354e0dac5`. The immutable [go.mod](https://github.com/golang/vuln/blob/e672cbe8fbc8c76e28fcb6db0efd9ce354e0dac5/go.mod) selects x/tools v0.52.0 and Go 1.26.0. The GitHub API source matches the downloaded module apart from line endings. The source and existing BSD license were reviewed on 2026-10-15. No tool source is copied into the product.

Update only the govulncheck pin in tools/quality.ps1. Keep the existing source-analysis command, package scope, default symbol-level scan, vulnerability database, failure propagation, and canonical quality workflow. Do not suppress findings, switch to a module-only scan, or change runtime dependencies to accommodate the scanner. The scanner remains a development tool and is not part of the compiled proxy.

## Validation

The old Windows-built scanner exits 2 with the same SSA panic on the Linux-target source graph. The v1.9.0 scanner completes that same analysis with exit 0 and zero reachable vulnerabilities. The canonical `tools/quality.ps1 -LowMemory` Windows gate also passes with the new pin, zero static-analysis issues, and zero reachable vulnerabilities. The scan reports one imported or required advisory outside called code.

Push the focused tooling commit and require fresh Linux and Windows CI results. A passing Linux-target analysis on Windows is useful regression evidence, but it does not replace the actual Linux CI run. [AUDIT-0214](../audits/AUDIT-0214-govulncheck-go127-linux.md) records the matched consumer and initial failed CI revision.

Live sessions remain tied to runtime `414eeb97157416eed8dfdc85f872101f40dd74af`. Their facts and tested revisions are unchanged. This tooling-only correction adds no packet, network, capture, viewer, or importer behavior. It does not complete the retail release matrix or claim a new live session.
