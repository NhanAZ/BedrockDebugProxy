# AUDIT-0225 - Defer upstream resource-pack completion

## Intent and scope

Fix upstream world startup overlapping downstream resource-pack delivery. This is a runtime change to the existing terminating login path, not a server-specific timeout, packet rewrite, or synthetic gameplay response. The [decision](../decisions/0014-defer-upstream-pack-completion.md) records the state transition and sources. [FIX-0161](../fix-history.md#fix-0161---defer-upstream-world-startup-during-downstream-pack-delivery) indexes the root cause.

The change adds one opt-in boundary to the retained gophertunnel dialer, resumes it through DoSpawnContext after downstream pack completion, and obtains upstream GameData only after the resumed spawn completes. The nested tests cover the deferred boundary, missing archives, conflicting modes, cancellation, and concurrent resume calls. The existing mixed URL/chunk loopback now checks completion order and still checks offer metadata and archive identity. Capture schema, packet codecs, archive bytes, pack keys, download pacing, and public CLI syntax are unchanged. No viewer contract changes.

## Evidence and provenance

The earlier authenticated [OP SMP failure](https://github.com/BedrockDebugProxy-Captures/session-20261018T165644Z/releases/tag/capture-archive) retains one Hub archive and seven child archives. Upstream gameplay begins while the client is still receiving packs and the remote endpoint disconnects before that delivery finishes. The [previous audit](AUDIT-0224-paired-stream-error-validation.md) records its original unresolved classification.

A controlled client with a cold pack cache reproduced the startup failure. A warm-cache control completed spawn and 60 seconds of traffic. Delaying only player initialisation failed, so that experimental source change was removed. Deferring resource-pack completion then completed cold-cache OP SMP delivery, spawn, and 60 seconds of traffic in a preliminary unstamped working build. These preliminary runs do not establish an exact committed runtime or retail gameplay. The follow-up audit will record stamped captures and sanitized reports before this task is considered complete.

The protocol and retained MIT implementation were inspected directly. This patch and its tests are independently written, and third-party provenance notes were updated. No upstream issue or pull request was posted.

## Paired consumer

BedrockPackArchive commit `a8114b8d1e1b89bec697a1228d00efaff964a0c3` extends the synthetic root and hop 2 fixtures with ordered downstream and upstream pack completion followed by session.spawned. The importer contract is unchanged, so no production importer change is needed. The fixtures retain archive digests, root and nested manifests, plaintext contents.json, three declared synthetic keys, source sequence, generator version, hop, and menu path. Full consumer tests, vet, build, and whitespace checks pass. The follow-up records the producer's final SHA and actual stamped-verifier check. Consumer fixtures contain no real captures or pack data.

## Validation

The old early-completion setting fails the loopback ordering assertion with upstream sequence 48 preceding downstream sequence 74. The deferred setting passes. Focused root tests include ordinary bidirectional forwarding, mixed packs, and both packs-only boundaries. Focused root and nested race checks pass, as do the full nested minecraft tests. The canonical tools/format.ps1 and tools/quality.ps1 -LowMemory workflow passes, including the full root tests, build, static analysis, module and license checks, viewer asset checks, and whitespace checks. The vulnerability scan reports no reachable project vulnerability, with uncalled dependency findings retained in its output.

The next automated owner-controlled check uses the stamped build, a new cold client cache, and the Enchanted Hub to OP SMP transfer. It must retain one Hub and seven child packs, record downstream completion before upstream completion, record both hops spawned, and remain connected for 60 seconds after child spawn. Box PvP, Mines, and The Hive provide related-flow and baseline checks. If a check fails, preserve its closed capture and inspect completion order, Disconnect, decode, and transport errors. Retail rendering, player movement, and official release validation remain unobserved.
