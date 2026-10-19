# AUDIT-0234 - Accept full CLI sessions as release evidence

The maintainer explicitly requested that AGENTS.md and the validation workflow accept CLI tests without Minecraft and that version 0.4.1 be released. Decision 0015 records this deliberate policy change separately from its tooling implementation.

The agent guide now permits full live CLI sessions with authenticated upstream access, complete pack exchange, spawn, at least 30 seconds of subsequent bidirectional decoded traffic, and verified closed captures with complete artifacts. The six-server matrix, exact revisions, CubeCraft acknowledgement, known pack-bearing Enchanted transfer, privacy boundaries, and failure investigation remain required. Packs-only and offline simulations remain insufficient. Retail rendering and gameplay are not claimed by CLI reports.

This commit changes policy documentation only. No runtime, dependency, packet, capture, pack output, importer, viewer, or authentication default changes. The accompanying workflow change supplies regression tests. Canonical tools/format.ps1, tools/quality.ps1 -LowMemory, and whitespace checks passed before these commits, with zero static-analysis issues and no called vulnerabilities. Historical reports and source captures remain unchanged. Fresh candidate CLI reports are required before the requested release.
