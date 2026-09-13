package raknet

import "testing"

func TestDialerUsesConfiguredClientGUID(t *testing.T) {
	const want int64 = -123456789
	if got := (Dialer{ClientGUID: want}).clientGUID(); got != want {
		t.Fatalf("configured client GUID = %d, want %d", got, want)
	}
}

func TestDialerGeneratesDistinctNegativeClientGUIDs(t *testing.T) {
	dialer := Dialer{}
	first, second := dialer.clientGUID(), dialer.clientGUID()
	if first >= 0 || second >= 0 || first == second {
		t.Fatalf("generated client GUIDs = %d and %d, want distinct negative values", first, second)
	}
}

func TestConnExposesDialledClientGUID(t *testing.T) {
	conn := &Conn{clientGUID: 42}
	if got := conn.ClientGUID(); got != 42 {
		t.Fatalf("connection client GUID = %d, want 42", got)
	}
}
