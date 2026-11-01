package proxy

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/NhanAZ/BedrockDebugProxy/internal/bedrock"
	"github.com/NhanAZ/BedrockDebugProxy/internal/capture"
	"github.com/NhanAZ/BedrockDebugProxy/internal/experience"
	"github.com/sandertv/gophertunnel/minecraft"
)

const transferNetherNetProbeTimeout = 10 * time.Second

func transferHTTPSAddress(target string) (string, error) {
	host, port, err := net.SplitHostPort(strings.TrimSpace(target))
	if err != nil {
		return "", fmt.Errorf("split transfer target %q: %w", target, err)
	}
	if strings.TrimSpace(host) == "" {
		return "", fmt.Errorf("transfer target %q has an empty host", target)
	}
	portNumber, err := strconv.ParseUint(port, 10, 16)
	if err != nil || portNumber == 0 {
		return "", fmt.Errorf("transfer target %q has an invalid port", target)
	}
	return "https://" + net.JoinHostPort(host, strconv.FormatUint(portNumber, 10)), nil
}

func transferHTTPSNetherNet(ctx context.Context, target string) (minecraft.Network, string, error) {
	address, err := transferHTTPSAddress(target)
	if err != nil {
		return nil, "", err
	}
	probeCtx, cancel := context.WithTimeout(ctx, transferNetherNetProbeTimeout)
	defer cancel()
	network, err := experience.NewHTTPSNetherNetNetwork(probeCtx, http.DefaultClient, address)
	if err != nil {
		return nil, "", err
	}
	return network, address, nil
}

func (r *Runner) prepareTransferHop(ctx context.Context, state *hopState, hop int, logger *slog.Logger, failures *bedrock.FailureSink, live *liveReporter) error {
	if network, address, err := transferHTTPSNetherNet(ctx, state.target); err == nil {
		state.transferNetwork = network
		state.transferAddress = address
		state.transferRouteProbed = true
		captureErr := r.recordAtHop("transfer.route_probe", capture.SeverityInfo, map[string]any{
			"address":   state.target,
			"transport": "nethernet-https",
			"timeout":   transferNetherNetProbeTimeout.String(),
		}, hop)
		setCaptureFailure(failures, captureErr)
		if captureErr != nil {
			return captureErr
		}
		live.Info("Transfer route selected HTTPS NetherNet - %s", address)
		return nil
	}

	if state.transferLocalAddr != nil {
		if err := r.probeTransferRoute(ctx, state.target, state.transferLocalAddr, state.transferClientGUID, hop, logger, failures, live); err != nil {
			live.Info("Transfer route probe failed - %s", err)
		}
		state.transferRouteProbed = true
	}
	return nil
}
