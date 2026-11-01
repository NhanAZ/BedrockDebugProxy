# AUDIT-0248 - Use HTTPS NetherNet for transferred hops

- Commit subject: `Fix HTTPS NetherNet transfer hops`
- Change type: `runtime`
- Runtime impact: `yes`

## Intent

Allow transferred child servers that expose HTTPS NetherNet on the server-provided host and port to complete the authenticated child connection. Preserve compatibility with direct RakNet transfer endpoints.

## Scope

- Classify each transfer endpoint through the existing authenticated HTTPS NetherNet status path before reconnecting.
- Use the authenticated NetherNet network and HTTPS target when the endpoint is supported.
- Retain the previous transfer RakNet source address, client GUID, and MTU fallback when HTTPS NetherNet is unavailable.
- Add address-normalization regression coverage for hostnames, IPv6 addresses, and invalid ports.
- Keep packet capture boundaries, authentication requirements, and server policy behavior unchanged.

## Evidence and provenance

The affected Enchanted child endpoints returned HTTPS NetherNet status responses while direct RakNet handshakes stopped before connection acceptance. A focused agent-controlled Fishing session reached the child server, completed its resource packs, spawned, and remained stable after selecting HTTPS NetherNet. A separate Box PvP session confirmed the same route without regressing nested-form handling. The implementation reuses the project's existing authenticated NetherNet network instead of adding server-specific packet mutation or policy-evasion behavior.

## Validation

- Automated checks: `go test ./...`, `tools/quality.ps1 -LowMemory`, and the stamped candidate build.
- Live validation: `pending` at commit creation; the exact candidate must pass the six-server release matrix before tagging.
- Capture or report: sanitized release reports are generated only for the exact stamped candidate revision and are kept out of source history.

## Limitations and follow-up

HTTPS status probing is bounded to ten seconds to allow the first TLS/DNS setup to complete. Endpoints that do not answer as HTTPS NetherNet continue through the existing RakNet path. Release validation must still confirm The Hive, Galaxite, Lifeboat, Mineville Zeqa, Enchanted, and CubeCraft independently, with Enchanted's current form and route matrix reviewed from fresh Hub sessions.

## Security and privacy

No credentials, captures, server-provided addresses, resource-pack keys, decrypted packs, or third-party payloads are stored in this audit record. The route probe uses the existing certificate-validated HTTPS client and authenticated identity flow.
