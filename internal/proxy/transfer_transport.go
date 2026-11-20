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
	"github.com/google/uuid"
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

func serviceNetherNetTransferID(network minecraft.Network, upstreamAddress, target string) (string, bool) {
	if network == nil {
		return "", false
	}
	if _, err := uuid.Parse(strings.TrimSpace(upstreamAddress)); err != nil {
		return "", false
	}
	host, port, err := net.SplitHostPort(strings.TrimSpace(target))
	if err != nil {
		return "", false
	}
	portNumber, err := strconv.ParseUint(port, 10, 16)
	if err != nil || portNumber != 0 {
		return "", false
	}
	id, err := uuid.Parse(strings.TrimSpace(host))
	if err != nil {
		return "", false
	}
	return id.String(), true
}

func (r *Runner) prepareTransferHop(ctx context.Context, state *hopState, hop int, logger *slog.Logger, failures *bedrock.FailureSink, live *liveReporter) error {
	if id, ok := serviceNetherNetTransferID(r.config.UpstreamNetwork, r.config.UpstreamAddress, state.target); ok {
		state.transferNetwork = r.config.UpstreamNetwork
		state.transferAddress = id
		state.transferRouteProbed = true
		captureErr := r.recordAtHop("transfer.route_selection", capture.SeverityInfo, map[string]any{
			"transport":          "nethernet-signaling",
			"address_type":       "network_id",
			"reused_signaling":   true,
			"selection_evidence": "upstream_network_id_and_zero_port_transfer",
		}, hop)
		setCaptureFailure(failures, captureErr)
		if captureErr != nil {
			return captureErr
		}
		live.Info("Transfer route selected the existing NetherNet signaling connection for a server-sent network ID")
		return nil
	}
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
	} else {
		captureErr := r.recordAtHop("transfer.route_probe", capture.SeverityWarn, map[string]any{
			"address":            state.target,
			"transport":          "nethernet-https",
			"fallback_transport": "raknet",
			"timeout":            transferNetherNetProbeTimeout.String(),
			"best_effort":        true,
			"error":              err.Error(),
			"error_type":         fmt.Sprintf("%T", err),
		}, hop)
		setCaptureFailure(failures, captureErr)
		if captureErr != nil {
			return captureErr
		}
		live.Info("Transfer HTTPS NetherNet probe failed - %s, falling back to RakNet", err)
	}

	if state.transferLocalAddr != nil {
		if err := r.probeTransferRoute(ctx, state.target, state.transferLocalAddr, state.transferClientGUID, hop, logger, failures, live); err != nil {
			live.Info("Transfer route probe failed - %s", err)
		}
		state.transferRouteProbed = true
	}
	return nil
}
