package packet

import (
	"bytes"
	"strings"
	"testing"
)

func TestDecoderAllowsConfiguredBatchPacketCount(t *testing.T) {
	const maxPackets = 1024
	decoder := NewDecoder(bytes.NewReader(testBatch(maxPackets)))
	decoder.SetMaxBatchPacketCount(maxPackets)

	packets, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if len(packets) != maxPackets {
		t.Fatalf("Decode() returned %d packets, want %d", len(packets), maxPackets)
	}
}

func TestDecoderRejectsBatchAboveConfiguredPacketCount(t *testing.T) {
	const maxPackets = 1024
	decoder := NewDecoder(bytes.NewReader(testBatch(maxPackets + 1)))
	decoder.SetMaxBatchPacketCount(maxPackets)

	_, err := decoder.Decode()
	if err == nil || !strings.Contains(err.Error(), "number of packets exceeds max=1024") {
		t.Fatalf("Decode() error = %v, want configured packet-count limit error", err)
	}
}

func TestDecoderRetainsDefaultBatchPacketCount(t *testing.T) {
	decoder := NewDecoder(bytes.NewReader(testBatch(defaultMaximumInBatch + 1)))

	_, err := decoder.Decode()
	if err == nil || !strings.Contains(err.Error(), "number of packets exceeds max=812") {
		t.Fatalf("Decode() error = %v, want default packet-count limit error", err)
	}
}

func TestDecoderCanDisableBatchPacketLimit(t *testing.T) {
	decoder := NewDecoder(bytes.NewReader(testBatch(defaultMaximumInBatch + 1)))
	decoder.DisableBatchPacketLimit()

	packets, err := decoder.Decode()
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if len(packets) != defaultMaximumInBatch+1 {
		t.Fatalf("Decode() returned %d packets, want %d", len(packets), defaultMaximumInBatch+1)
	}
}

func testBatch(packetCount int) []byte {
	var batch bytes.Buffer
	_ = batch.WriteByte(header)
	var length [5]byte
	for range packetCount {
		_, _ = batch.Write(length[:putVaruint32(&length, 1)])
		_ = batch.WriteByte(0)
	}
	return batch.Bytes()
}
