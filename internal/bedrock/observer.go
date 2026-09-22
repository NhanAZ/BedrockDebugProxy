package bedrock

import (
	"context"
	"net"
	"sync"

	"github.com/NhanAZ/BedrockDebugProxy/internal/capture"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type flow struct {
	local        string
	remote       string
	connectionID string
	hop          int
}

type RawObserver func(channel string, direction capture.Direction, header packet.Header)

type Observer struct {
	recorder  *capture.Recorder
	failures  *FailureSink
	sessionID string
	hop       int
	raw       RawObserver

	mu    sync.Mutex
	flows map[string][]flow
}

func NewObserver(recorder *capture.Recorder, failures *FailureSink, sessionID string, hop int, raw RawObserver) *Observer {
	return &Observer{
		recorder:  recorder,
		failures:  failures,
		sessionID: sessionID,
		hop:       hop,
		raw:       raw,
		flows:     make(map[string][]flow),
	}
}

func (o *Observer) SetFlow(channel string, local, remote net.Addr) {
	o.SetFlowWithIdentity(channel, local, remote, "", 0)
}

// SetFlowWithIdentity records a physical transport flow and the connection
// identity that was selected when that flow was opened. A listener can have
// an old and a new flow alive briefly during a transfer, so packet callbacks
// must use the matching physical flow rather than a mutable pending state.
func (o *Observer) SetFlowWithIdentity(channel string, local, remote net.Addr, connectionID string, hop int) {
	if local == nil || remote == nil {
		return
	}
	o.mu.Lock()
	flows := o.flows[channel]
	updated := false
	for index := range flows {
		if flows[index].local == local.String() && flows[index].remote == remote.String() {
			flows[index] = flow{local: local.String(), remote: remote.String(), connectionID: connectionID, hop: hop}
			updated = true
			break
		}
	}
	if !updated {
		flows = append(flows, flow{local: local.String(), remote: remote.String(), connectionID: connectionID, hop: hop})
	}
	o.flows[channel] = flows
	o.mu.Unlock()
}

func (o *Observer) PacketFunc(channel, connectionID string) func(packet.Header, []byte, net.Addr, net.Addr) {
	return o.PacketFuncDynamic(channel, func() string { return connectionID })
}

// PacketFuncDynamic is the packet callback variant used by a long-lived
// listener that accepts more than one physical connection over a session.
// The callback is evaluated for each packet so capture events retain the
// current hop and connection identity after a server transfer.
func (o *Observer) PacketFuncDynamic(channel string, connectionID func() string) func(packet.Header, []byte, net.Addr, net.Addr) {
	return func(header packet.Header, payload []byte, source, destination net.Addr) {
		direction, observedFlow := o.classify(channel, source, destination)
		observedConnectionID := connectionID()
		observedHop := o.currentHop()
		if observedFlow != nil && observedFlow.connectionID != "" {
			observedConnectionID = observedFlow.connectionID
			observedHop = observedFlow.hop
		}
		if o.raw != nil {
			o.raw(channel, direction, header)
		}
		_, err := o.recorder.Record(context.Background(), capture.Record{
			Event: capture.Event{
				SessionID:    o.sessionID,
				ConnectionID: observedConnectionID,
				Hop:          observedHop,
				Kind:         "packet.raw",
				Direction:    direction,
				Channel:      channel,
				Stage:        "post_decrypt_decompress_frame",
				Source:       endpoint(source),
				Destination:  endpoint(destination),
				Packet: &capture.PacketInfo{
					ID:              header.PacketID,
					SenderSubClient: header.SenderSubClient,
					TargetSubClient: header.TargetSubClient,
					DecodeStatus:    "not_yet_decoded",
				},
			},
			Raw:            payload,
			MediaType:      "application/octet-stream",
			Representation: "bedrock_packet_payload",
		})
		if err != nil {
			o.failures.Set(err)
		}
	}
}

func (o *Observer) SetHop(hop int) {
	o.mu.Lock()
	o.hop = hop
	o.mu.Unlock()
}

func (o *Observer) currentHop() int {
	o.mu.Lock()
	hop := o.hop
	o.mu.Unlock()
	return hop
}

func (o *Observer) classify(channel string, source, destination net.Addr) (capture.Direction, *flow) {
	if source == nil || destination == nil {
		return capture.DirectionUnknown, nil
	}
	o.mu.Lock()
	flows := append([]flow(nil), o.flows[channel]...)
	o.mu.Unlock()
	for index := len(flows) - 1; index >= 0; index-- {
		known := flows[index]
		read := source.String() == known.remote && destination.String() == known.local
		write := source.String() == known.local && destination.String() == known.remote
		if !read && !write {
			continue
		}
		switch channel {
		case "downstream":
			if read {
				return capture.DirectionClientToServer, &known
			}
			if write {
				return capture.DirectionServerToClient, &known
			}
		case "upstream":
			if read {
				return capture.DirectionServerToClient, &known
			}
			if write {
				return capture.DirectionClientToServer, &known
			}
		}
	}
	return capture.DirectionUnknown, nil
}

func endpoint(address net.Addr) *capture.Endpoint {
	if address == nil {
		return nil
	}
	return &capture.Endpoint{Network: address.Network(), Address: address.String()}
}
