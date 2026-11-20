# Require lobby spawn, then close release sessions

## Decision

Official release validation uses the authenticated Go `golem network` client for The Hive, Galaxite, Lifeboat, Mineville Zeqa, Enchanted, and CubeCraft. Each session must authenticate, complete every offered resource-pack exchange, successfully decrypt each encrypted offered pack, reach spawn, and close immediately. The capture and generated artifacts must pass integrity, bidirectional traffic, packet decoding, and blocking-error checks.

The release client does not wait after spawn, move, use lobby items, submit forms, or enter minigames. Enchanted's automatically sent Hub form remains in the capture as server runtime data, but no child selection or route matrix is required for release. CubeCraft must close after the lobby spawn without selecting a minigame. `run --packs-only` remains a development and acquisition check, but cannot pass the release gate because it stops before spawn.

## Rationale

The release baseline needs evidence that an authenticated client can join each required Experience, receive and decrypt its offered packs, and arrive at the lobby. A longer idle interval and gameplay interactions do not improve that join-and-pack signal for this project. Short sessions also follow the maintainer's instruction to leave the connection once the lobby is reached.

Child-route discovery is a separate pack-acquisition task. It uses current decoded forms, inventory, entity, and transfer evidence to decide which route to inspect and records pack-bearing routes for later refreshes. Historical release reports and captures remain unchanged.

## Evidence boundary

These protocol-client reports establish only the captured authenticated join, pack exchange, spawn, and clean closure. They do not establish retail rendering, movement, gameplay, client compatibility through the proxy relay, or coverage of unselected child modes. Focused tests and loopback evidence remain required for code paths outside the observed direct Golem session. The optional `stability` scenario can still be used for investigations and does not become a release requirement.

This decision supersedes the release gate in [decision 0015](0015-cli-release-validation.md) and the Enchanted release traversal requirements in [decision 0016](0016-context-aware-enchanted-validation.md). Decision 0016's dynamic-form guidance still applies to pack acquisition.
