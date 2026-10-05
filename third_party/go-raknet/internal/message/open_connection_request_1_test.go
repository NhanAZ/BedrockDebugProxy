package message

import (
	"sync"
	"testing"
)

func TestOpenConnectionRequest1ConcurrentMarshal(t *testing.T) {
	const workers = 32
	start := make(chan struct{})
	var group sync.WaitGroup
	for worker := range workers {
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			for iteration := range 64 {
				mtu := uint16(600 + worker*8 + iteration%8)
				pk := OpenConnectionRequest1{ClientProtocol: 11, MTU: mtu}
				data, err := pk.MarshalBinary()
				if err != nil {
					t.Errorf("marshal MTU %d: %v", mtu, err)
					return
				}
				if len(data) != int(mtu)-28 || data[0] != IDOpenConnectionRequest1 || data[17] != pk.ClientProtocol {
					t.Errorf("invalid connection request for MTU %d", mtu)
					return
				}
			}
		}()
	}
	close(start)
	group.Wait()
}
