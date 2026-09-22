package minecraft

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	cryptorand "crypto/rand"
	"encoding/json"
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

func TestOfflineEncodingOmitsAuthenticationBoundClaims(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P384(), cryptorand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	identity := login.IdentityData{
		XUID:           "1234567890",
		Identity:       "11111111-1111-4222-8222-111111111111",
		DisplayName:    "OfflineProbe",
		TitleID:        "896928775",
		PlayFabTitleID: "20CA2",
		PlayFabID:      "master-player-id",
	}
	clearXBLIdentityData(&identity)

	clientData := login.ClientData{}
	defaultClientData("127.0.0.1:19132", identity.DisplayName, &clientData, true)
	request := login.EncodeOffline(identity, clientData, key, false)
	parsed, _, authResult, err := login.Parse(request, nil)
	if err != nil {
		t.Fatalf("parse offline request: %v", err)
	}
	if authResult.XBOXLiveAuthenticated {
		t.Fatal("offline request was reported as Xbox-authenticated")
	}
	if parsed.XUID != "" || parsed.TitleID != "" || parsed.PlayFabTitleID != "" || parsed.PlayFabID != "" {
		t.Fatalf("authentication-bound claims were encoded: %#v", parsed)
	}
	if parsed.Identity != identity.Identity || parsed.DisplayName != identity.DisplayName {
		t.Fatalf("offline identity changed: got %#v, want %#v", parsed, identity)
	}
}

func TestAuthenticatedClientDataPreservesEmptyPlayFabID(t *testing.T) {
	clientData := login.ClientData{}
	defaultClientData("127.0.0.1:19132", "AuthenticatedProbe", &clientData, false)

	if clientData.PlayFabID != "" {
		t.Fatalf("authenticated ClientData.PlayFabID = %q, want empty when the client omitted it", clientData.PlayFabID)
	}
}

func TestAuthenticatedClientDataOmitsEmptyPlayFabIDFromJSON(t *testing.T) {
	clientData := login.ClientData{PlayFabID: ""}
	encoded, err := json.Marshal(clientData)
	if err != nil {
		t.Fatalf("marshal empty authenticated ClientData: %v", err)
	}
	if bytes.Contains(encoded, []byte(`"PlayFabId"`)) {
		t.Fatalf("empty authenticated ClientData still contains PlayFabId: %s", encoded)
	}

	clientData.PlayFabID = "0123456789abcdef"
	encoded, err = json.Marshal(clientData)
	if err != nil {
		t.Fatalf("marshal populated authenticated ClientData: %v", err)
	}
	if !bytes.Contains(encoded, []byte(`"PlayFabId":"0123456789abcdef"`)) {
		t.Fatalf("populated ClientData omitted PlayFabId: %s", encoded)
	}
}
