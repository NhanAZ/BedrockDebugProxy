# Audit AUDIT-0170

## Intent

Allow the maintainer to start an owner-controlled live validation session through a direct internal request without requiring confidential server messages to be disclosed or stored in project artifacts.

## Scope

Updated `AGENTS.md`, `docs/server-targets.md`, `docs/validation.md`, `docs/releasing.md`, and `CHANGELOG.md`. The release workflow treats the maintainer's direct internal request as sufficient authorization and records only observed technical evidence. This commit has no runtime impact and does not alter proxy, capture, or protocol behavior.

## Evidence and provenance

The maintainer stated that the Hive authorization is confidential and should not be disclosed. No private message, token, address, or captured payload was added to the repository. Public policy references remain unchanged and are still presented as policy context rather than as permission.

## Automated validation

- `git diff --check` passed.
- Documentation-only change. No runtime tests are required.

## Live validation status

No new live session is required for this documentation-only change. The existing Hive capture remains tied to runtime revision `93435c5119e2899dc1958d8cec0fa63447bd611c`; its local report contains only observed technical evidence.

## Known limitations and follow-up

The workflow does not claim that a server's public policy changed. The maintainer remains responsible for the scope and consequences of the test. CubeCraft retains its separate explicit risk acknowledgement gate.
