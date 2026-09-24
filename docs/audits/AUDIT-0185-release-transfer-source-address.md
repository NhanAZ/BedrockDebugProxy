# Audit 0185 - Release transfer source address

- Commit subject: `Release transfer source address before route probe`
- Change type: `runtime compatibility fix`
- Runtime impact: `yes`

## Intent

Allow transfer following to reuse the preceding upstream RakNet source address for the route probe and next-hop dial on platforms that keep a UDP socket bound until the peer acknowledges its close.

## Scope

- Add an explicit immediate-close capability to the in-tree go-raknet connection and propagate it through gophertunnel and the capture-observing wrapper.
- Use immediate close only after forwarding a server `Transfer`; normal connection-close behavior remains unchanged.
- Require the controlled loopback transfer test to observe a successful route probe, and test that RakNet immediate close releases its underlying packet connection.
- Update the existing transfer decision, handshake research note, patch provenance, fix history, changelog, and prior test-maintenance audit with the CI result.

## Evidence and provenance

GitHub Actions run `35946284078` for commit `18e66b72daf1e01d9aa9a6ee0cfb43d68be31a52` failed on Linux when the route probe attempted to bind the previous hop's local UDP address and the next-hop login failed afterward. The existing transfer decision and Enchanted handshake research already require the next hop to reuse the source address and RakNet GUID. Source review of the in-tree go-raknet close path confirmed that `Close` retains the socket while waiting for a peer acknowledgement, for up to five seconds.

## Validation

- Automated checks: `tools/quality.ps1 -Race` passed. The complete race-detected Go suite, build, viewer checks, module and license checks, static analysis (0 issues), and whitespace validation passed. The vulnerability scan found 0 reachable vulnerabilities, plus 4 in imported packages and 2 in required modules that the code does not appear to call.
- Focused loopback transfer test: passed locally on Windows after the change.
- go-raknet tests: passed locally, including the immediate-close regression test. The full vendored gophertunnel test suite also passed.
- CI: pending for the commit containing this record.
- Live validation: not run. A stamped binary must complete the required The Hive baseline and release server matrix.

## Limitations and follow-up

Immediate close sends a RakNet disconnect notification and releases the local socket without waiting for a close acknowledgement. This path is limited to transfer following; it does not alter normal close behavior. The local loopback result does not establish compatibility with a real featured-experience route. Release approval remains blocked on exact-revision CI and the required live validation reports.
