# Go network-client Golem

Accepted on 2026-11-02 for the first implementation step after moving the PHP
GolemAgent companion to `NhanAZ-Plugins/GolemAgent`.

## Decision

Add a Go network-client mode as `bedrock-debug-proxy golem network`. Keep the
existing `golem --data` mailbox command temporarily so existing disposable
PocketMine and Axolotl workflows do not break while the network client gains
feature parity. The new client uses the repository's pinned gophertunnel,
go-raknet, experience resolution, and authentication-cache paths. It does not
copy protocol code from another client implementation.

The first boundary is login, resource-pack exchange, spawn, server form
inspection, bounded form responses, chat, Transfer following, and a private
structured session report. The client stores no credentials in the report and
does not claim that its report is a canonical capture or a release report.

## Rationale

The PHP companion simulates a Player inside a server. That is valuable for
plugin API tests but cannot exercise Microsoft authentication, server-side
login policy, RakNet or NetherNet transport, resource-pack negotiation, server
forms, or Featured Experience Transfer routing. A network client is the
smallest architecture that addresses those missing public-server boundaries.

Reusing the existing gophertunnel dialer avoids a second login and packet codec.
Reusing `experience.Resolve` keeps Featured and Creator Experience discovery in
one reviewed path. The network bot still has a separate session state machine
because its scenario actions and Transfer lifecycle differ from the observation
proxy's downstream forwarding lifecycle.

## Security and evidence boundaries

Scenario JSON, form JSON, server messages, and Transfer destinations are
untrusted. The implementation validates them at the boundary, caps retained
data, caps transfer hops, uses the existing authentication cache without
printing token values, and creates reports as private files. Packet mutation,
replay, injection, anti-cheat evasion, and authentication bypass remain out of
scope.

The network report is supplemental evidence. It does not replace a verified
BedrockDebugProxy capture, resource-pack decryption artifacts, or the existing
release report checks. A later implementation step must add canonical capture
integration, client tick and input behavior, complete pack decryption evidence,
and a stability check before the network bot can become the release driver.

No BedrockPackArchive change is required for this first step. The implementation
does not change the proxy's capture or pack archive producer contract and does
not write reconstructed pack output.
