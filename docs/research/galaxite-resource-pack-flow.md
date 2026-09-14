# Galaxite resource-pack exchange observation

This note records a live observation used to tune the upstream connection deadline. It is evidence about one session, not a claim about every Galaxite backend or future pack set.

## Capture

- Server label: Galaxite
- Selector: `experience:Galaxite`
- Tested revision: `0fc7ef843d7d7f9d91b988a5c9e1dbabd94e2290`
- Protocol: Bedrock `1.26.45`, protocol `2169`
- Capture: `captures/session-20260914T074555Z`
- Observed: 2026-09-14

The capture showed an upstream connection followed by 19 `ResourcePackDataInfo` packets. The advertised chunk sizes were 102,400 bytes, with 460 chunks expected in total. The server continued returning `ResourcePackChunkData` while the proxy issued the corresponding `ResourcePackChunkRequest` packets, but only 161 chunks had arrived when the previous one-minute dial deadline expired. No Bedrock `Disconnect` packet was observed from Galaxite before the proxy cancelled the connection.

## Project impact

The one-minute deadline in the proxy was too short for this active multi-pack exchange. The proxy therefore produced its own timeout and the client saw a disconnect while still loading resource packs. The deadline is now five minutes. This is a bounded wait, not a packet mutation or a server-specific protocol workaround. Raw packet and transport evidence remain in the capture so a later revision can replace the duration with a progress-aware policy if more server observations justify it.
