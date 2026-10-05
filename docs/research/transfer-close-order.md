# Transfer completion after a client close

Reviewed on 2026-10-05 against BedrockDebugProxy revision `818a6e682ae4019d2b54f858f4fdc169c33b4346`.

## Confirmed local behavior

The runHop implementation in internal/proxy/proxy.go receives the first terminal forwarding-loop result, recognizes a transfer only in that result, and then closes both connections before receiving the second result. The server-to-client forwarder flushes a rewritten Transfer before recording timing and returning its transferError. A client may receive that flushed packet and close its old connection while the forwarder is still completing those steps. The downstream loop can therefore return context.Canceled first and the completed transfer second.

The completion-order regression uses a synthetic loopback target, source UDP address, and GUID. It verifies both orders and wrapped errors, and fails for the client-close-first order with the original selection behavior. The unrelated-write-failure case continues to reject transfer following. This is a new local completion-order root cause. FIX-0005 covers transfer routing, FIX-0013 covers racing shutdown diagnostics, and FIX-0141 covers immediate socket release.

## Focused correction

Select a second-result transfer only when the first result wraps context.Canceled. Keep the first-result transfer path unchanged. Preserve the original typed transferError, source address, GUID, and both terminal error strings. Release the upstream socket immediately before following the recovered transfer. Do not infer a transfer from an observed packet alone, suppress an unrelated terminal failure, change packet bytes, or special-case a server endpoint.

Capture schema and importer event contracts are unchanged. session.close already supports first_loop_error, second_loop_error, and transfer_target. BedrockPackArchive's paired synthetic fixture now imports both hop 1 and hop 2 archives, including a prior transfer close event and the child menu path, and checks archive digest, metadata bytes, contents.json, declared key count, source sequence, and index provenance. No importer production change is needed.

## Validation boundary

The existing two-hop loopback test covers transfer delivery, reconnect, the route probe, hop attribution, and capture verification. Completion-order tests separately cover the deterministic race. The paired audit records the canonical quality gate and stamped producer-verifier check. Live Enchanted menu traversal and a The Hive packs-only baseline remain distinct from retail-client play and the official release matrix. This note cites synthetic evidence only and contains no real capture, server asset, account identity, or key.
