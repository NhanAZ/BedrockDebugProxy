package experience

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/service"
)

func TestIsSelector(t *testing.T) {
	for _, value := range []string{"experience:The Hive", "EXPERIENCE:CubeCraft"} {
		if !IsSelector(value) {
			t.Fatalf("IsSelector(%q) = false", value)
		}
	}
	if IsSelector("example.org:19132") {
		t.Fatal("literal address was treated as an experience selector")
	}
}

func TestFindFeaturedServerUsesExactCaseInsensitiveNameOrID(t *testing.T) {
	servers := []featuredServer{
		{Name: "The Hive", ExperienceID: "first"},
		{Name: "Hive Games", ExperienceID: "second"},
	}
	server, ok := findFeaturedServer(servers, "the hive")
	if !ok || server.ExperienceID != "first" {
		t.Fatalf("findFeaturedServer() = %#v, %t", server, ok)
	}
	server, ok = findFeaturedServer(servers, "second")
	if !ok || server.Name != "Hive Games" {
		t.Fatalf("findFeaturedServer() by ID = %#v, %t", server, ok)
	}
	if _, ok := findFeaturedServer(servers, "Hive"); ok {
		t.Fatal("partial name unexpectedly matched")
	}
}

func TestListFeaturedServers(t *testing.T) {
	server := newTLSServer(t, func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/v2.0/discovery/blob/client" || request.Method != http.MethodPost {
			t.Fatalf("request = %s %s", request.Method, request.URL.Path)
		}
		if request.Header.Get("Authorization") != "MCToken test" {
			t.Fatalf("Authorization = %q", request.Header.Get("Authorization"))
		}
		_, _ = w.Write([]byte(`{"data":{"Items":[{"Title":{"NEUTRAL":"The Hive"},"DisplayProperties":{"url":"example.org","port":19132,"experienceId":"123"}}]}}`))
	})
	servers, err := listFeaturedServers(context.Background(), server.Client(), server.URL, &service.Token{AuthorizationHeader: "MCToken test"})
	if err != nil {
		t.Fatal(err)
	}
	if len(servers) != 1 || servers[0].Name != "The Hive" || servers[0].Address != "example.org:19132" || servers[0].ExperienceID != "123" {
		t.Fatalf("servers = %#v", servers)
	}
}

func TestJoinExperienceDestinations(t *testing.T) {
	experienceID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	netherNetID := "22222222-2222-2222-2222-222222222222"
	tests := []struct {
		name           string
		response       string
		defaultAddress string
		want           resolvedDestination
		wantErr        string
	}{
		{
			name:     "raknet",
			response: `{"result":{"networkProtocol":"RakNet","ipV4Address":"192.0.2.10","port":19132}}`,
			want:     resolvedDestination{Address: "192.0.2.10:19132", Transport: transportRakNet},
		},
		{
			name:     "legacy nethernet",
			response: `{"result":{"networkProtocol":"NetherNet","netherNetId":"` + netherNetID + `"}}`,
			want:     resolvedDestination{Address: netherNetID, Transport: transportNetherNet},
		},
		{
			name:     "json rpc nethernet",
			response: `{"result":{"networkProtocol":"NetherNet_JsonRpc","netherNetId":"` + netherNetID + `"}}`,
			want:     resolvedDestination{Address: netherNetID, Transport: transportJSONRPC},
		},
		{
			name:           "default uses discovery address",
			response:       `{"result":{"networkProtocol":"Default"}}`,
			defaultAddress: "hive.example:19132",
			want:           resolvedDestination{Address: "hive.example:19132", Transport: transportRakNet},
		},
		{
			name:           "default prefers join address",
			response:       `{"result":{"networkProtocol":"Default","ipV4Address":"192.0.2.11","port":19133}}`,
			defaultAddress: "hive.example:19132",
			want:           resolvedDestination{Address: "192.0.2.11:19133", Transport: transportRakNet},
		},
		{
			name:     "default without discovery address",
			response: `{"result":{"networkProtocol":"Default"}}`,
			wantErr:  "no usable destination",
		},
		{
			name:     "missing destination",
			response: `{"result":{"networkProtocol":"RakNet","port":19132}}`,
			wantErr:  "no usable destination",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := newTLSServer(t, func(w http.ResponseWriter, request *http.Request) {
				if request.URL.Path != "/api/v2.0/join/experience" || request.Method != http.MethodPost {
					t.Fatalf("request = %s %s", request.Method, request.URL.Path)
				}
				var body struct {
					ExperienceID string `json:"experienceId"`
				}
				if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
					t.Fatal(err)
				}
				if body.ExperienceID != experienceID.String() {
					t.Fatalf("experienceId = %q", body.ExperienceID)
				}
				_, _ = w.Write([]byte(test.response))
			})
			got, err := joinExperience(context.Background(), server.Client(), server.URL, &service.Token{}, experienceID, test.defaultAddress)
			if test.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantErr) {
					t.Fatalf("joinExperience() error = %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != test.want {
				t.Fatalf("joinExperience() = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestDecodeReceivedMessagesAcceptsOneOrMany(t *testing.T) {
	one, err := decodeReceivedMessages(json.RawMessage(`{"From":"one","Message":"payload"}`))
	if err != nil || len(one) != 1 || one[0].From != "one" {
		t.Fatalf("single = %#v, %v", one, err)
	}
	many, err := decodeReceivedMessages(json.RawMessage(`[{"From":"one"},{"From":"two"}]`))
	if err != nil || len(many) != 2 || many[1].From != "two" {
		t.Fatalf("many = %#v, %v", many, err)
	}
}

func newTLSServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	server := httptest.NewTLSServer(handler)
	t.Cleanup(server.Close)
	return server
}
