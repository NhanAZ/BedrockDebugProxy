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

func TestClearXBLIdentityDataRemovesAuthenticationBoundFields(t *testing.T) {
	data := &login.IdentityData{
		XUID:           "1234567890",
		Identity:       "11111111-1111-1111-1111-111111111111",
		DisplayName:    "OfflineProbe",
		TitleID:        "896928775",
		PlayFabTitleID: "20CA2",
		PlayFabID:      "master-player-id",
	}

	clearXBLIdentityData(data)

	if data.XUID != "" || data.TitleID != "" || data.PlayFabTitleID != "" || data.PlayFabID != "" {
		t.Fatalf("authentication-bound identity fields were retained: %#v", data)
	}
	if data.Identity != "11111111-1111-1111-1111-111111111111" || data.DisplayName != "OfflineProbe" {
		t.Fatalf("offline identity fields changed: %#v", data)
	}
}
