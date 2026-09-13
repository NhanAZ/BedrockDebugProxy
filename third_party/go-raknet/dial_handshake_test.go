package raknet

import (
	"context"
	"encoding/binary"
	"net"
	"net/netip"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/sandertv/go-raknet/internal/message"
)

func TestMTUSizesFor(t *testing.T) {
	tests := []struct {
		name string
		max  uint16
		want []uint16
	}{
		{name: "default", want: []uint16{1492, 1200, 576}},
		{name: "transfer", max: 1200, want: []uint16{1200, 576}},
		{name: "custom", max: 1000, want: []uint16{1000, 576}},
		{name: "minimum", max: 100, want: []uint16{576}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := mtuSizesFor(test.max); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("MTU sizes = %v, want %v", got, test.want)
			}
		})
	}
}

func TestOpenConnectionRefreshesSecurityCookie(t *testing.T) {
	const (
		oldCookie = uint32(0x10203040)
		newCookie = uint32(0x50607080)
	)
	conn := &cookieRefreshConn{packets: make(chan []byte, 2), oldCookie: oldCookie, newCookie: newCookie}
	state := &connState{
		conn:           conn,
		raddr:          testUDPAddr("192.0.2.10:19132"),
		mtu:            1200,
		serverSecurity: true,
		ticker:         time.NewTicker(time.Second / 2),
	}
	defer state.ticker.Stop()
	state.cookie.Store(oldCookie)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := state.openConnection(ctx); err != nil {
		t.Fatal(err)
	}
	if got := state.cookie.Load(); got != newCookie {
		t.Fatalf("security cookie = %#x, want %#x", got, newCookie)
	}
	if !conn.usedRefreshedCookie() {
		t.Fatal("OpenConnectionRequest2 was not resent with the refreshed cookie")
	}
}

type cookieRefreshConn struct {
	packets chan []byte

	mu          sync.Mutex
	oldCookie   uint32
	newCookie   uint32
	sentRefresh bool
	sentReply2  bool
}

func (c *cookieRefreshConn) Read(b []byte) (int, error) {
	packetBytes := <-c.packets
	return copy(b, packetBytes), nil
}

func (c *cookieRefreshConn) Write(b []byte) (int, error) {
	if len(b) < 21 || b[0] != message.IDOpenConnectionRequest2 {
		return len(b), nil
	}
	cookie := binary.BigEndian.Uint32(b[17:21])
	c.mu.Lock()
	defer c.mu.Unlock()
	switch {
	case cookie == c.oldCookie && !c.sentRefresh:
		c.sentRefresh = true
		response, _ := (&message.OpenConnectionReply1{
			ServerGUID:        2,
			ServerHasSecurity: true,
			Cookie:            c.newCookie,
			MTU:               1200,
		}).MarshalBinary()
		c.packets <- response
	case cookie == c.newCookie && !c.sentReply2:
		c.sentReply2 = true
		response, _ := (&message.OpenConnectionReply2{
			ServerGUID:    2,
			ClientAddress: netip.MustParseAddrPort("192.0.2.20:50000"),
			MTU:           1200,
		}).MarshalBinary()
		c.packets <- response
	}
	return len(b), nil
}

func (c *cookieRefreshConn) usedRefreshedCookie() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.sentReply2
}

func (*cookieRefreshConn) Close() error                     { return nil }
func (*cookieRefreshConn) LocalAddr() net.Addr              { return testUDPAddr("192.0.2.20:50000") }
func (*cookieRefreshConn) RemoteAddr() net.Addr             { return testUDPAddr("192.0.2.10:19132") }
func (*cookieRefreshConn) SetDeadline(time.Time) error      { return nil }
func (*cookieRefreshConn) SetReadDeadline(time.Time) error  { return nil }
func (*cookieRefreshConn) SetWriteDeadline(time.Time) error { return nil }

func testUDPAddr(address string) *net.UDPAddr {
	addr, _ := net.ResolveUDPAddr("udp", address)
	return addr
}
