# AUDIT-0220 - Skip unsupported HTTPS NetherNet status ping

- Commit subject: `fix: skip unsupported HTTPS NetherNet ping`
- Change type: runtime networking
- Runtime impact: yes. HTTPS NetherNet routes no longer attempt RakNet address-discovery ping before dialing.

## Intent and evidence

The first stamped CubeCraft retail session after FIX-0158 spawned and forwarded traffic, but its validation report identified a blocking `transport.ping` error. Local event inspection showed `NetherNet.PingContext: not supported`. The existing vendored gophertunnel dialer ignores optional ping errors and dials the original address. `internal/experience/target.go` had already checked the advertised HTTPS endpoint status before making the target available. Skipping this unsupported operation is narrower than changing the capture error classification or inventing a synthetic ping response. The original incomplete report remains local for comparison.

## Validation

- Run the canonical quality gate and `git diff --check`.
- Build a clean stamped candidate and repeat retail CubeCraft validation. The report must show spawn, two-way traffic, decoded packets, no blocking errors, and `risk_acknowledged=pass`.
- Repeat the other five release targets at the same runtime revision. Their existing RakNet route pings remain enabled.

## Limitations

This change applies to HTTPS destinations only. It does not alter HTTPS signaling, authentication, encrypted gameplay, capture schema, or the behavior of ID-based NetherNet and RakNet destinations.
