# ADR 0004 - Preserve same-protocol packet bytes on the bridge

## Status

Accepted on 2026-09-14.

## Context

The bridge must expose decoded packet views without changing the packet bytes sent to a server or client when no rewrite is requested. Before this decision, the forwarding loop called `ReadPacket` and `WritePacket` for every application packet. That path is semantically convenient, but it decodes and serializes each packet again. Serialization can change optional-field encoding, packet ordering inside a converted batch, or other wire details even when the decoded packet has the same meaning.

The latest Galaxite capture ended with an upstream Bedrock `Disconnect` carrying `A connection issue occurred` after the resource-pack exchange and normal spawn. Gophertunnel consumes that disconnect while closing the upstream connection, so the capture proves that the server sent the disconnect but does not identify the server-side rule that caused it. A direct vanilla client was reported to complete the same flow. This is sufficient evidence to remove an avoidable proxy-side re-serialization difference, but not to claim that every server accepts a terminating proxy.

## Decision

When both bridge endpoints expose the project-specific raw packet adapter and negotiate the same Bedrock protocol ID:

1. Read the packet after gophertunnel transport decryption, decompression, and framing while retaining its complete packet header and payload.
2. Decode an owned clone for the structured capture and live summaries.
3. Forward the original packet bytes through the destination encoder without decoding and re-marshalling them.
4. Keep the existing typed path for protocol mismatches, packets already consumed and returned as decoded values by the login state machine, and the `Transfer` rewrite used by `--follow-transfers`.

The raw adapter is limited to the current gophertunnel protocol boundary. It does not bypass encryption, authentication, batching, compression, RakNet, or server policy. It also does not mutate, drop, inject, or replay gameplay packets.

## Consequences

Same-protocol gameplay and resource-pack packets now retain their observed wire representation across the bridge while remaining available as decoded capture views. The capture records the raw packet once, and the forward timing event identifies whether raw or typed forwarding was used.

Protocol conversion and intentional transfer-address rewriting still use the typed path and therefore remain visible exceptions. A raw forwarding success is not proof of vanilla identity, and it does not make proxy use permitted by any server. The Galaxite result must be re-tested with a stamped binary before this change is considered live-compatible.

## Evidence and provenance

- Sandertv gophertunnel `v1.61.0`, commit [`283a5a97dfe65da94bcc0b401807f6aefa9e72ee`](https://github.com/Sandertv/gophertunnel/tree/283a5a97dfe65da94bcc0b401807f6aefa9e72ee), MIT. The packet hook and connection decoder establish the raw packet boundary and the internal disconnect handling.
- BedrockDebugProxy capture `session-20260914T092052Z`, capture ID `6c6ab952-e3f0-4a8c-a4f5-0ca31e16a006`, generated from revision `2818a674915109993e994de5953c83ff50d3b165`. The session was closed and verified with 17,697 events and 5,689 blobs; its final upstream error was the server disconnect described above.
- The raw adapter and bridge change are independent BedrockDebugProxy code. No source was copied from `bedrock-tool/bedrocktool` or another external repository.
