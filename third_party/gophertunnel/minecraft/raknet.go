package minecraft

import (
	"context"
	"github.com/sandertv/go-raknet"
	"log/slog"
	"net"
)

// RakNet is an implementation of a RakNet v10 Network.
type RakNet struct {
	l *slog.Logger

	// LocalAddr optionally pins the UDP source address used for dialing and
	// pinging. Transfer-aware proxies can reuse the source port from the
	// preceding RakNet hop when a featured experience routes by UDP flow.
	LocalAddr *net.UDPAddr
	// MaxMTU optionally caps MTU discovery for this connection.
	MaxMTU uint16
}

// NewRakNet returns a RakNet network using l for transport diagnostics.
// Optional transfer settings can be assigned on the returned value before it
// is passed to a Minecraft dialer.
func NewRakNet(l *slog.Logger) RakNet {
	return RakNet{l: l}
}

func (r RakNet) dialer() raknet.Dialer {
	dialer := raknet.Dialer{ErrorLog: r.l, MaxMTU: r.MaxMTU}
	if r.LocalAddr != nil {
		dialer.UpstreamDialer = &net.Dialer{LocalAddr: r.LocalAddr}
	}
	return dialer
}

// DialContext ...
func (r RakNet) DialContext(ctx context.Context, address string) (net.Conn, error) {
	return r.dialer().DialContext(ctx, address)
}

// PingContext ...
func (r RakNet) PingContext(ctx context.Context, address string) (response []byte, err error) {
	return r.dialer().PingContext(ctx, address)
}

// Listen ...
func (r RakNet) Listen(address string) (NetworkListener, error) {
	return raknet.ListenConfig{ErrorLog: r.l.With("net origin", "raknet")}.Listen(address)
}

// init registers the RakNet network.
func init() {
	RegisterNetwork("raknet", func(l *slog.Logger) Network { return RakNet{l: l} })
}
