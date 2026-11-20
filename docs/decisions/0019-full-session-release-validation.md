# Require post-spawn stability for release validation

## Decision

Every official release requires authenticated Go `golem network` sessions for The Hive, Galaxite, Lifeboat, Mineville Zeqa, Enchanted, and CubeCraft. Each session must complete its offered packs, reach spawn, retain bidirectional decoded traffic for at least 30 seconds after spawn, and close with verified capture integrity and complete artifacts. Decrypt every encrypted offered pack. The report must pass `cli_session`, `authenticated_upstream`, and `post_spawn_stability`.

Use the stationary `stability` scenario operation for the 30-second post-spawn window, then send `quit`. Do not request movement, select lobby forms or items, or enter minigames. Capture Enchanted's automatic Hub form without submitting a choice. For CubeCraft, collect the lobby offer and close without selecting a minigame. A retail Minecraft client is not required.

Enchanted child-route inspection remains part of pack acquisition. Use its current forms and Transfers, retain route history, and classify unavailable or no-transfer outcomes from observed packets. Do not explore unrelated public minigames to satisfy release validation.

## Rationale

The maintainer requires a full authenticated protocol session that remains stable after spawn. The stationary heartbeat observes the joined session without adding movement or gameplay. The Golem client records and validates the same protocol evidence without opening a retail client.

## Evidence boundary

Passing reports establish authentication, pack exchange, spawn, the observed post-spawn interval, bidirectional decoded traffic, and capture integrity for the named target. They do not establish retail rendering or gameplay. Packs-only reports remain insufficient for release.

This decision supersedes [decision 0018](0018-minimal-spawn-release-validation.md) and restores the 30-second stability requirement from [decision 0015](0015-cli-release-validation.md). Enchanted route acquisition remains governed by [decision 0016](0016-context-aware-enchanted-validation.md).
