# Galaxite repeated RakNet acceptance

The first post-limit-fix Galaxite capture reached the upstream RakNet connection and then closed with a repeated `CONNECTION_REQUEST_ACCEPTED` error. The raw capture remains local because it contains endpoint and authenticated session evidence.

## Evidence

- Capture: `captures/session-20260922T141819Z`
- Capture ID: `1de1bf87-bc59-47fe-949d-6f0eee6ea9fc`
- Runtime revision: `730e6b972dd16c21a94c3b9c2e2119d68e2ea7ea`
- Event: `library.log` at sequence 23, `unexpected additional CONNECTION_REQUEST_ACCEPTED packet`
- Result: the upstream login did not reach the Bedrock login or spawn phase

The same build's second Galaxite attempt (`captures/session-20260922T142754Z`) failed earlier with `POST https://multiplayer.minecraft.net/authentication: unexpected EOF`. That is a separate authentication-service boundary and is not evidence about the RakNet acceptance path.

## Source comparison

The vendored go-raknet handler was based on `v1.15.2-0.20260705184311-0d1fd09e2cf6`. The current upstream handler at commit `ea813dc668b5a2a2767cc5577bf77869c965f27a` still returns an error for a second acceptance in `dialerConnectionHandler.handleConnectionRequestAccepted`. The upstream pull request [#60](https://github.com/Sandertv/go-raknet/pull/60) separately documents repeated open-connection challenges during the earlier MTU phase. Neither source provides a recovery path for a repeated acceptance after the first `NEW_INCOMING_CONNECTION`.

## Local correction

The dialer now treats a repeated acceptance as a retransmission and sends `NEW_INCOMING_CONNECTION` again using the new `PongTime`. The connection remains established if that response is accepted. `TestDialerResendsNewIncomingConnectionAfterRepeatedAcceptance` verifies the recovery without a real server. This is a transport reliability correction, not packet mutation, replay, authentication bypass, or server-specific routing.

## Validation status

The correction requires a new stamped Galaxite capture that reaches upstream login, spawn, and normal shutdown. The existing failed reports remain unchanged and are not upgraded by this code change.
