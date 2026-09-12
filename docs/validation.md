# Real-world validation workflow

Automated tests establish that the implementation matches local expectations. They do not establish that those expectations match live Minecraft Bedrock clients and servers. Runtime, networking, protocol, authentication, session, resource-pack, capture, encoding, and decoding changes remain awaiting manual validation until this workflow is complete.

For the end-to-end pull request checklist, use [`../CONTRIBUTING.md`](../CONTRIBUTING.md). For the ordered six-server release checklist, use [`releasing.md`](releasing.md). This document defines how to create and judge the validation evidence used by both workflows.

The automated suite includes a loopback integration session over RakNet. It exercises offline login, resource-pack negotiation with no packs, StartGame and spawn, typed packet forwarding in both directions, clean shutdown, raw packet capture, decoded events, connection metadata, and GameData snapshots. This catches local forwarding and capture regressions without accounts or public infrastructure. It does not exercise Microsoft authentication, the retail Minecraft client, public-server routing, transfers, live resource packs, or server-specific behavior and therefore does not replace the gates below.

## Development and pull requests

The Hive is the minimum baseline for every runtime-affecting change and every pull request that changes code or observable behavior. A short session is sufficient only when it demonstrates successful authentication and connection, spawn, normal client interaction, traffic in both directions, packet decoding, stable shutdown, and no unexplained decode or protocol errors.

A change to a specific flow also requires direct validation of that flow. For example, resource-pack decryption requires both a The Hive baseline report and an owner-controlled encrypted-pack report that observes the supported decryption path. If the target does not exercise the changed flow, record `not_observed` and an `incomplete` result rather than treating a general connection as sufficient.

Do not approve or merge a runtime-affecting pull request without reports for the exact revision. Documentation, comments, formatting, and changes proven not to affect runtime may be exempt. When impact is uncertain, require validation.

## Stamped test build

Commit the revision, ensure the working tree is clean, run the quality gate, and create a binary whose capture manifest records the exact Git commit.

```powershell
.\tools\quality.ps1
.\tools\build.ps1 -Version dev
.\bin\bedrock-debug-proxy.exe version
```

`tools/build.ps1` refuses a dirty tree. Do not use an unstamped `go build` binary for validation evidence because its capture cannot prove which revision was tested.

## Capture and report

Run the stamped binary and complete the required live session. Stop it cleanly, then generate the sanitized report with project tooling rather than adding another command to the product CLI.

```powershell
.\bin\bedrock-debug-proxy.exe `
    run `
    --listen 0.0.0.0:19132 `
    --upstream "experience:The Hive" `
    --auth device
.\tools\create-validation-report.ps1 `
    -Binary .\bin\bedrock-debug-proxy.exe `
    -Capture .\captures\<SESSION> `
    -Server "The Hive" `
    -Result pass `
    -Check @("normal_session=pass") `
    -Output .\validation\local\<REVISION>\the-hive.json
```

Device authentication is cached after the first successful login. Wait for `Listening on` before connecting the client. Required baseline reports use the default Xbox-authenticated downstream path, normally through an entry in Minecraft's Servers tab. A LAN World entry uses self-signed client authentication and requires `--allow-unauthenticated-client` on a trusted LAN. Validate that opt-in separately when the LAN flow changes rather than using it as a substitute for the default authentication gate.

Use public server labels rather than addresses in `-Server`. A manual check uses `lowercase_name=pass`, `lowercase_name=fail`, or `lowercase_name=not_observed`. Add one for each flow reviewed, such as `resource_pack_transfer=pass` or `resource_pack_decryption=pass`.

A report cannot claim `pass` unless `normal_session=pass`, every other manual check passes, and capture-derived evidence proves all of the following facts.

- The capture is closed and free of verifier issues, write errors, dropped events, and truncated events.
- The upstream connected and the session negotiated, spawned, and closed.
- Client-to-server and server-to-client events were observed.
- Decoded packet events were observed.
- No packet decode, structured snapshot, resource-pack decryption, or other blocking error event was recorded.
- Sessions declaring automatic artifact folders have an error-free completed `artifacts/status.json` matching the capture ID, build revision, and final event sequence. See [session artifacts](session-artifacts.md) for the output and its limits.

The manifest `complete` flag remains in the report but is not by itself a pass gate. The current capture schema sets it to `false` whenever a known observation boundary is declared, including the intentional absence of raw UDP and RakNet acknowledgement frames. The stricter pass gate uses the verifier plus explicit write, dropped, and truncated counters so a documented scope boundary is not confused with data loss inside the supported boundary.

Normal connection termination currently produces a read-side `transport.payload` error and `bridge.read_error` evidence when a read loop ends. The report preserves counts for those events as termination errors without automatically treating them as protocol failures. The developer must still inspect them and may mark the session `fail` when their timing or cause is abnormal. Transport writes, library errors, resource-pack derivation errors, and all other error kinds block an automatic pass.

The report derives revision, timestamps, client and upstream versions, protocol IDs, traffic counts, resource-pack counts, and safe error-kind counts from the capture. It omits addresses, packet payloads, error messages, player identifiers, content keys, resource-pack metadata, decrypted assets, and the text of capture limitations because a dynamic limitation may contain a pack identifier. It retains only the limitation count. Review the JSON before uploading it. Keep the raw capture local and sensitive.

If an anomaly exists, use `-Result fail`. If a server or required flow could not be tested, use `-Result incomplete` and the applicable `not_observed` check. Never omit a failed or unavailable baseline silently. A report is evidence of one observed session, not a guarantee for other servers or later revisions.

Generate candidate reports under the ignored `validation/local/<REVISION>/` directory. Upload the sanitized JSON as pull request or release evidence without adding it to the candidate commit. Committing a report would change the pull request head or release revision after the tested binary was stamped. Keep any long-term archive as a release asset or in a deliberate later documentation-only record that preserves the original `tested_revision`.

Do not commit `.bdpcap`, raw captures, authentication state, resource-pack keys, or third-party assets.

Add `--decrypt-resource-packs` only when the change or release test explicitly needs to validate the authorized resource-pack decryption flow. It is not part of the default connection sanity test.

## Release gate

Every official release requires a full client session and a revision-matched report for each baseline server.

- The Hive
- CubeCraft
- Galaxite
- Lifeboat
- Mineville Zeqa (`experience:Mineville Zeqa`)
- Enchanted

For each server, check connection, authentication, resource packs when offered, spawn, both traffic directions, packet decoding, stability, and unexplained protocol or decode errors. Do not assume the six servers use the same packet order, timing, optional packets, software, or infrastructure.

A baseline failure starts an investigation. Determine whether the cause is a proxy bug, protocol misunderstanding, valid server behavior, or an external outage before changing code. Do not add a server-specific workaround merely to make the release matrix green. If an external condition prevents testing, retain an `incomplete` report and keep the release not ready until the maintainer explicitly resolves the gate.

After generating all six reports for one revision, run `tools/check-release-readiness.ps1 -Revision <40-character-commit>`. It checks that every required label has a passing report with automatic and manual gates for that exact revision. It does not replace human review of the local captures.
