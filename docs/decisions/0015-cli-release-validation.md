# Accept authenticated CLI sessions for release validation

## Decision

On 2026-10-19 the maintainer explicitly requested CLI validation without a retail Minecraft client and release of BedrockDebugProxy 0.4.1. A full live session driven by a CLI client may satisfy each required server baseline. The upstream must authenticate with the approved account, resolve the Experience, complete offered packs, reach spawn, retain bidirectional decoded traffic for at least 30 seconds after spawn, and close with verified capture integrity and complete artifacts.

CLI reports use cli_session=pass, authenticated_upstream=pass, and post_spawn_stability=pass. Retail reports retain normal_session=pass. Both paths require every automatic check and every supplied review check to pass. Packs-only, local offline loopback tests, and the server-side Golem simulation cannot replace the live matrix. Report labels, exact tested revisions, failed-session handling, CubeCraft risk acknowledgement, and the required Enchanted child transfer remain enforced.

The CLI is an ordinary downstream protocol client of the stamped proxy. It may use the existing self-signed client opt-in only on a concrete loopback listener. The proxy's upstream still uses the approved cached Microsoft account. This does not change product authentication defaults or authorize public unauthenticated listeners, packet mutation in the proxy, policy evasion, or arbitrary minigame exploration.

## Evidence boundary

This decision changes the release acceptance criterion. It does not turn CLI sessions into evidence of retail rendering, input handling, UI, or visible gameplay. Those claims require their own observations. Preserve earlier incomplete reports and generate new reports under the current criterion without modifying captures or tested_revision. Any non-documentation candidate change still requires a fresh exact-revision matrix.

The rationale is the project's protocol-observation mission. Full CLI sessions exercise authentication, resource packs, spawn, transport, packet forwarding, decoding, and capture closure without requiring the retail application. Visible gameplay remains an optional, separately described evidence layer.
