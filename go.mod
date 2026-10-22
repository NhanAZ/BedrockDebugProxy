module github.com/NhanAZ/BedrockDebugProxy

go 1.26.0

toolchain go1.27.3

require (
	github.com/coder/websocket v1.8.15
	github.com/df-mc/go-nethernet v1.0.24
	github.com/df-mc/go-playfab/v2 v2.0.3
	github.com/df-mc/go-xsapi/v2 v2.0.3
	github.com/go-gl/mathgl v1.2.0
	github.com/google/uuid v1.6.0
	github.com/sandertv/go-raknet v1.15.2-0.20260705184311-0d1fd09e2cf6
	github.com/sandertv/gophertunnel v1.61.0
	golang.org/x/oauth2 v0.37.0
)

// Enchanted and some other featured experiences omit ChunkRadiusUpdated during
// the upstream spawn sequence. Keep the small compatibility patch in-tree
// until it is available in the selected upstream gophertunnel release.
replace github.com/sandertv/gophertunnel => ./third_party/gophertunnel

// Keep the RakNet client GUID observable and reusable across featured-experience
// transfers. The local copy is a narrow MIT-compatible patch over the selected
// upstream revision; see third_party/go-raknet/BEDROCKDEBUGPROXY_PATCH.md.
replace github.com/sandertv/go-raknet => ./third_party/go-raknet

require (
	github.com/coreos/go-oidc/v3 v3.17.0 // indirect
	github.com/df-mc/jsonc v1.0.5 // indirect
	github.com/go-jose/go-jose/v4 v4.1.4 // indirect
	github.com/klauspost/compress v1.18.4 // indirect
	github.com/pion/datachannel v1.6.3 // indirect
	github.com/pion/dtls/v3 v3.1.9 // indirect
	github.com/pion/ice/v4 v4.4.4 // indirect
	github.com/pion/interceptor v0.1.49 // indirect
	github.com/pion/logging v0.2.4 // indirect
	github.com/pion/mdns/v2 v2.2.1 // indirect
	github.com/pion/randutil v0.1.0 // indirect
	github.com/pion/rtcp v1.2.17 // indirect
	github.com/pion/rtp v1.10.5 // indirect
	github.com/pion/sctp v1.11.3 // indirect
	github.com/pion/sdp/v3 v3.0.20 // indirect
	github.com/pion/srtp/v3 v3.1.0 // indirect
	github.com/pion/stun/v4 v4.0.1 // indirect
	github.com/pion/transport/v5 v5.0.1 // indirect
	github.com/pion/turn/v5 v5.1.2 // indirect
	github.com/pion/webrtc/v4 v4.2.21-0.20260920133716-91bfc6c2039f // indirect
	github.com/wlynxg/anet v0.0.5 // indirect
	golang.org/x/crypto v0.57.0 // indirect
	golang.org/x/exp v0.0.0-20240909161429-701f63a606c0 // indirect
	golang.org/x/net v0.60.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/text v0.42.0 // indirect
	golang.org/x/time v0.14.0 // indirect
)
