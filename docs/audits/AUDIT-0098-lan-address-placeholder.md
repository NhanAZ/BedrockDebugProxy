# Audit 0098 - Use a LAN address placeholder in examples

## Intent

Prevent documentation users from copying an operator-specific LAN address that is not assigned on their machine.

## Scope

- Replace `192.168.1.10` in transfer, validation, and release command examples with `<LAN_IP>`.
- Keep `0.0.0.0` and `127.0.0.1` examples where they intentionally document wildcard and loopback binding.
- Add a changelog entry.
- No runtime, protocol, capture, or CLI behavior change.

## Evidence and provenance

The listener bind error is produced when the requested address is not assigned to a local interface. The previous example address was not a project invariant and differs from the operator's observed LAN address.

## Automated validation

- `git diff --check`
- Full `tools/quality.ps1` is required before commit.
- Repository search confirms no user-facing example retains `192.168.1.10`.

## Live validation

Not applicable. This change only corrects copy-paste documentation.

## Known limitations and follow-up

The operator must still choose an address reachable by the Minecraft client. A wildcard listener cannot be sent in a rewritten `Transfer` packet.
