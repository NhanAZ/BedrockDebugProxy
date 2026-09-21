package minecraft

import (
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/login"
)

func TestSetAndroidDataUsesCurrentVersion(t *testing.T) {
	data := &login.ClientData{
		DeviceOS:     protocol.DeviceWin32,
		GameVersion:  "1.26.50",
		LanguageCode: "en_GB",
	}

	setAndroidData(data)

	if data.DeviceOS != protocol.DeviceAndroid {
		t.Fatalf("device OS = %d, want Android", data.DeviceOS)
	}
	if data.GameVersion != protocol.CurrentVersion {
		t.Fatalf("game version = %q, want %q", data.GameVersion, protocol.CurrentVersion)
	}
}
