# AUDIT-0240 - Update the compression security patch

GitHub Dependabot alert 17 reports GHSA-259r-337f-4rfw in github.com/klauspost/compress through 1.18.6. The project selected 1.18.4. The local gophertunnel replacement imports flate and s2. The scanner did not find a call to the vulnerable NewDict function, but the affected runtime dependency should be updated before release.

The [upstream v1.18.7 release](https://github.com/klauspost/compress/releases/tag/v1.18.7) identifies an s2.NewDict out-of-bounds-read fix and describes a security release. This change selects 1.18.7 in the root module, updates checksums and the Unreleased changelog, and retains existing transport integration and license notices. No capture contract, protocol definition, CLI, or decryption behavior changes.

The canonical race-enabled quality gate must pass before commit. Its packet, compression, capture, transfer, and loopback coverage exercises the existing integration. The security-updated exact revision requires new CI and a fresh stamped six-server matrix. Historical October 20 reports remain unchanged. The maintainer accepted the CubeCraft connection risk for that new matrix. Release publication remains gated by the completed matrix and archive verification.

Go 1.27.3 with compress 1.18.7 passed tools/format.ps1 and tools/quality.ps1 -Race. Static analysis found zero issues. Vulnerability analysis found zero affected called functions and zero affected imported packages. No suppression was added.
