package bedrock

import (
	"context"
	"net"
	"sync"

	"github.com/NhanAZ/BedrockDebugProxy/internal/capture"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

type flow struct {
	local  string
	remote string
}

type Observer struct {
	recorder  *capture.Recorder
	failures  *FailureSink
	sessionID string
	hop       int

	mu    sync.Mutex
	flows map[string]flow
}

func NewObserver(recorder *capture.Recorder, failures *FailureSink, sessionID string, hop int) *Observer {
	return &Observer{
		recorder:  recorder,
		failures:  failures,
		sessionID: sessionID,
		hop:       hop,
		flows:     make(map[string]flow),
	}
}

func (o *Observer) SetFlow(channel string, local, remote net.Addr) {
	if local == nil || remote == nil {
		return
	}
	o.mu.Lock()
	o.flows[channel] = flow{local: local.String(), remote: remote.String()}
	o.mu.Unlock()
}

func (o *Observer) PacketFunc(channel, connectionID string) func(packet.Header, []byte, net.Addr, net.Addr) {
	return func(header packet.Header, payload []byte, source, destination net.Addr) {
		direction := o.classify(channel, source, destination)
		_, err := o.recorder.Record(context.Background(), capture.Record{
			Event: capture.Event{
				SessionID:    o.sessionID,
				ConnectionID: connectionID,
				Hop:          o.hop,
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

func (o *Observer) classify(channel string, source, destination net.Addr) capture.Direction {
	if source == nil || destination == nil {
		return capture.DirectionUnknown
	}
	o.mu.Lock()
	known, ok := o.flows[channel]
	o.mu.Unlock()
	if !ok {
		return capture.DirectionUnknown
	}
	read := source.String() == known.remote && destination.String() == known.local
	write := source.String() == known.local && destination.String() == known.remote
	switch channel {
	case "downstream":
		if read {
			return capture.DirectionClientToServer
		}
		if write {
			return capture.DirectionServerToClient
		}
	case "upstream":
		if read {
			return capture.DirectionServerToClient
		}
		if write {
			return capture.DirectionClientToServer
		}
	}
	return capture.DirectionUnknown
}

func endpoint(address net.Addr) *capture.Endpoint {
	if address == nil {
		return nil
	}
	return &capture.Endpoint{Network: address.Network(), Address: address.String()}
}
