package minecraft

import (
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func TestEmptyPackDownloadRequestPreservesState(t *testing.T) {
	for _, active := range []bool{false, true} {
		t.Run(map[bool]string{false: "before download", true: "during download"}[active], func(t *testing.T) {
			conn := &Conn{}
			expectedID := uint32(packet.IDResourcePackClientResponse)
			if active {
				conn.packQueue = &resourcePackQueue{currentOffset: 42}
				expectedID = packet.IDResourcePackChunkRequest
			}
			queue := conn.packQueue
			conn.expect(expectedID)
			if err := conn.handleResourcePackClientResponse(&packet.ResourcePackClientResponse{Response: packet.PackResponseSendPacks}); err != nil {
				t.Fatalf("empty download request: %v", err)
			}
			if conn.packQueue != queue {
				t.Fatal("empty request replaced the download queue")
			}
			if active && conn.packQueue.currentOffset != 42 {
				t.Fatal("empty request changed the active download offset")
			}
			expected := conn.expectedIDs.Load().([]uint32)
			if len(expected) != 1 || expected[0] != expectedID {
				t.Fatalf("empty request changed expected packets: %v", expected)
			}
			if conn.loggedIn {
				t.Fatal("empty request completed login")
			}
		})
	}
}
