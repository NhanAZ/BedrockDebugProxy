# Audit AUDIT-0173

## Intent

Recover a potentially retransmitted RakNet `CONNECTION_REQUEST_ACCEPTED` during an upstream dial instead of aborting the session after the first `NEW_INCOMING_CONNECTION` response.

## Scope

Changed the vendored go-raknet dialer handler and added a deterministic regression test. Updated dependency provenance, fix history, research notes, and the changelog. Bedrock packet payloads, capture storage, authentication data, and transfer rewriting are unchanged.

## Evidence and provenance

The change is based on Galaxite capture `1de1bf87-bc59-47fe-949d-6f0eee6ea9fc` at runtime revision `730e6b972dd16c21a94c3b9c2e2119d68e2ea7ea`. The capture recorded `unexpected additional CONNECTION_REQUEST_ACCEPTED packet` after the upstream dialer had sent its initial response. The upstream go-raknet handler at commit `ea813dc668b5a2a2767cc5577bf77869c965f27a` has the same fatal branch. The local change remains within the MIT-licensed vendored dependency and is documented in its patch note.

## Automated validation

- `gofmt` passed for changed Go files.
- `go test ./...` passed in `third_party/go-raknet`.
- Full project quality gate and stamped build remain required before live validation.

## Live validation status

Pending a fresh stamped Galaxite session. The prior retry failed separately at the Microsoft authentication-chain request with `unexpected EOF` and is not used as evidence for this transport correction.

## Known limitations and follow-up

The recovery assumes a repeated acceptance is a retransmission of the same initial handshake and re-sends only the required `NEW_INCOMING_CONNECTION`. A new server capture must confirm that the session proceeds to Bedrock login and spawn. The local code does not retry authentication requests or suppress any captured error.
