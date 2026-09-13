package proxy

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// transferForProxyAddress keeps the server's transfer semantics while routing
// the client back through this proxy. The original target is returned
// separately so the next upstream connection can be established after the
// client reconnects.
func transferForProxyAddress(transfer *packet.Transfer, proxyAddress string) (*packet.Transfer, string, error) {
	if transfer == nil {
		return nil, "", errors.New("transfer packet is nil")
	}
	host := strings.TrimSpace(transfer.Address)
	if host == "" {
		return nil, "", errors.New("transfer address is empty")
	}
	host = strings.Trim(host, "[]")
	target := net.JoinHostPort(host, strconv.Itoa(int(transfer.Port)))
	if err := validateAddress(target); err != nil {
		return nil, "", fmt.Errorf("invalid transfer target %q: %w", target, err)
	}
	if err := validateProxyTransferAddress(proxyAddress); err != nil {
		return nil, "", err
	}
	rewrittenHost, rewrittenPort, err := net.SplitHostPort(proxyAddress)
	if err != nil {
		return nil, "", fmt.Errorf("split proxy transfer address: %w", err)
	}
	port, err := strconv.ParseUint(rewrittenPort, 10, 16)
	if err != nil {
		return nil, "", fmt.Errorf("parse proxy transfer port: %w", err)
	}
	rewritten := &packet.Transfer{
		Address:     rewrittenHost,
		Port:        uint16(port),
		ReloadWorld: transfer.ReloadWorld,
	}
	rewritten.GatheringJoinInfo = transfer.GatheringJoinInfo
	return rewritten, target, nil
}

func validateProxyTransferAddress(address string) error {
	if err := validateAddress(address); err != nil {
		return fmt.Errorf("proxy listener cannot be used for Transfer: %w", err)
	}
	host, _, _ := net.SplitHostPort(address)
	host = strings.TrimSpace(strings.Trim(host, "[]"))
	if host == "" || host == "0.0.0.0" || host == "::" || host == "::0" {
		return fmt.Errorf("proxy listener address %q is not reachable by the Minecraft client; use a concrete LAN address with --listen", address)
	}
	return nil
}
