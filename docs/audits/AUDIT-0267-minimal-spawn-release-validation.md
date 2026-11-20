# AUDIT-0267 - Require minimal spawn release validation

- Commit subject: `Require minimal spawn release validation`
- Change type: `tooling`
- Runtime impact: `yes`

## Intent

Align the official release gate with the maintainer's requested evidence: an authenticated client must complete the offered packs, reach the lobby spawn, and close immediately. The user does not need to operate Minecraft. Release validation does not require movement, lobby interaction, minigames, or a post-spawn waiting interval.

## Scope and provenance

The Go Golem report now marks a clean authenticated spawn-and-close session as a passing CLI session without adding a stability check. The report generator and six-server readiness gate keep the automatic requirements for connection, negotiation, spawn, bidirectional decoded traffic, completed packs, encrypted-pack decryption, artifacts, and capture integrity. Enchanted's automatic Hub form remains captured, but child-route traversal belongs to pack acquisition. `--packs-only` remains insufficient because it stops before spawn.

This follows the maintainer's direct instruction in the current release task. It does not change packet handling, authentication, resource-pack transfer, the capture schema, or server routing. The paired BedrockPackArchive update is commit `7ded565cb93d37ce528e7cb672b1adb01c667ba4`, pushed to its existing `main` branch. It accepts the producer's 0.6.x version and adds `Rob a Mob` to its separate Experience acquisition order. The fixture against the stamped v0.6.0 producer remains pending.

The verbose Go vulnerability scan found no reachable vulnerable symbols. Its package findings concern `golang.org/x/net/html` `Parse` and `Render`, while the vendored tunnel code uses only `NewTokenizer`; `compress/zstd` is present only through a dependency's tests while production uses S2; and the project does not use `golang.org/x/crypto/openpgp`. The official reports are [GO-2026-6860](https://pkg.go.dev/vuln/GO-2026-6860), [GO-2026-6866](https://pkg.go.dev/vuln/GO-2026-6866), [GO-2026-6865](https://pkg.go.dev/vuln/GO-2026-6865), and [GO-2026-5932](https://pkg.go.dev/vuln/GO-2026-5932). These findings do not justify changing the active compression path or adding an unused replacement dependency in this release.

## Validation

- Automated checks: `.\tools\quality.ps1` passed, including tests, build, static analysis with 0 issues, tracked-capture scan, and whitespace checks. `govulncheck v1.9.0 -show verbose ./...` found 0 vulnerable symbols. It reported 2 non-reachable package advisories and 2 module-only advisories, triaged above. Focused report tests passed before the full quality gate.
- Paired consumer: BedrockPackArchive `go test -count=1 ./...`, `go vet ./...`, staged Windows build, and `git diff --check` passed. The exact producer fixture remains to be run against the stamped v0.6.0 binary.
- Live validation: `pending`; run the authenticated Go Golem six-server matrix on the exact stamped revision, complete and decrypt offered packs, reach spawn, then close immediately.
- Capture or report: sanitized per-server reports will be stored under the ignored `validation/local/<REVISION>/` directory and attached to the release. Raw captures remain local and are not release assets.

## Limitations and follow-up

This gate proves only the observed authenticated join, pack exchange, lobby spawn, and clean close. It does not prove retail rendering, movement, gameplay, or unselected child-route coverage. `--packs-only` is an acquisition check and cannot pass the release gate. Release remains blocked until all six exact-revision reports pass and the paired BedrockPackArchive fixture passes against the stamped producer.
