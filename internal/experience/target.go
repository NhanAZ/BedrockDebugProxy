package experience

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/df-mc/go-nethernet"
	"github.com/df-mc/go-playfab/v2"
	"github.com/df-mc/go-xsapi/v2"
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/auth"
	"github.com/sandertv/gophertunnel/minecraft/service"
	"golang.org/x/oauth2"
)

const (
	selectorPrefix      = "experience:"
	maxServiceResponse  = 8 << 20
	minecraftUserAgent  = "libhttpclient/1.0.0.0"
	identityDomain      = "https://authorization.franchise.minecraft-services.net/"
	transportRakNet     = "raknet"
	transportNetherNet  = "nethernet"
	transportJSONRPC    = "nethernet-jsonrpc"
	protocolJSONRPCName = "NetherNet_JsonRpc"
)

// Target is a resolved upstream destination. Network is nil for RakNet.
type Target struct {
	Selector     string
	Name         string
	ExperienceID string
	Address      string
	Transport    string
	Network      minecraft.Network
	TokenSource  oauth2.TokenSource

	closeOnce sync.Once
	closeFn   func() error
	closeErr  error
}

// Close releases service authentication and signaling resources.
func (t *Target) Close() error {
	t.closeOnce.Do(func() {
		if t.closeFn != nil {
			t.closeErr = t.closeFn()
		}
	})
	return t.closeErr
}

// IsSelector reports whether value uses the experience selector syntax.
func IsSelector(value string) bool {
	return len(value) >= len(selectorPrefix) && strings.EqualFold(value[:len(selectorPrefix)], selectorPrefix)
}

// Resolve resolves experience:<name-or-uuid> through Minecraft services. A
// literal HOST:PORT is returned unchanged and does not perform authentication.
func Resolve(ctx context.Context, value string, live oauth2.TokenSource, httpClient *http.Client) (*Target, error) {
	if !IsSelector(value) {
		return &Target{Selector: value, Address: value, Transport: transportRakNet, TokenSource: live}, nil
	}
	name := strings.TrimSpace(value[len(selectorPrefix):])
	if name == "" {
		return nil, errors.New("experience selector is empty")
	}
	if live == nil {
		return nil, errors.New("experience targets require device authentication")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	session, err := newServiceSession(ctx, live, httpClient)
	if err != nil {
		return nil, fmt.Errorf("authenticate Minecraft services: %w", err)
	}
	target, err := session.resolve(ctx, value, name)
	if err != nil {
		_ = session.Close()
		return nil, err
	}
	return target, nil
}

type serviceSession struct {
	live          oauth2.TokenSource
	httpClient    *http.Client
	discovery     *service.Discovery
	authorization *service.AuthorizationEnvironment
	tokens        service.TokenSource
	playFab       *playfab.Client

	closeOnce sync.Once
	closeErr  error
}

func newServiceSession(ctx context.Context, live oauth2.TokenSource, httpClient *http.Client) (*serviceSession, error) {
	serviceContext := auth.WithContextClient(ctx, httpClient)
	discovery, err := service.Default(serviceContext)
	if err != nil {
		return nil, fmt.Errorf("discover service environments: %w", err)
	}
	authorization := &service.AuthorizationEnvironment{}
	if err := discovery.Environment(authorization); err != nil {
		return nil, fmt.Errorf("load authorization environment: %w", err)
	}
	authorization.HTTPClient = httpClient

	xboxClient, err := (xsapi.ClientConfig{
		HTTPClient: httpClient,
		RTAMode:    xsapi.RTADisabled,
	}).New(serviceContext, auth.ContextSession(serviceContext, live))
	if err != nil {
		return nil, fmt.Errorf("login to Xbox Live: %w", err)
	}
	playFabClient, err := playfab.LoginWithXbox(serviceContext, authorization.PlayFabTitleID, xboxClient, playfab.ClientConfig{
		HTTPClient:    httpClient,
		CreateAccount: true,
	})
	closeErr := xboxClient.Close()
	if err != nil {
		return nil, errors.Join(fmt.Errorf("login to PlayFab: %w", err), closeErr)
	}
	if closeErr != nil {
		_ = playFabClient.Close()
		return nil, fmt.Errorf("close Xbox Live client: %w", closeErr)
	}

	return &serviceSession{
		live:          live,
		httpClient:    httpClient,
		discovery:     discovery,
		authorization: authorization,
		tokens:        authorization.TokenSource(playFabClient, service.TokenConfig{}),
		playFab:       playFabClient,
	}, nil
}

func (s *serviceSession) Token() (*oauth2.Token, error) {
	return s.live.Token()
}

func (s *serviceSession) MultiplayerToken(ctx context.Context, key *ecdsa.PublicKey) (string, error) {
	return s.authorization.MultiplayerToken(ctx, s.tokens, key)
}

func (s *serviceSession) Close() error {
	s.closeOnce.Do(func() {
		s.closeErr = s.playFab.Close()
	})
	return s.closeErr
}

type gatheringsEnvironment struct {
	ServiceURI string `json:"serviceUri"`
}

func (*gatheringsEnvironment) ServiceName() string { return "gatherings" }

type signalingEnvironment struct {
	ServiceURI string `json:"serviceUri"`
}

func (*signalingEnvironment) ServiceName() string { return "signaling" }

type featuredServer struct {
	Name         string
	Address      string
	ExperienceID string
}

type resolvedDestination struct {
	Address   string
	Transport string
}

func (s *serviceSession) resolve(ctx context.Context, selector, name string) (*Target, error) {
	var gatherings gatheringsEnvironment
	if err := s.discovery.Environment(&gatherings); err != nil {
		return nil, fmt.Errorf("load gatherings environment: %w", err)
	}
	if err := validateHTTPSServiceURI(gatherings.ServiceURI); err != nil {
		return nil, fmt.Errorf("invalid gatherings environment: %w", err)
	}
	token, err := s.tokens.ServiceToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("request Minecraft services token: %w", err)
	}

	experienceID := name
	displayName := name
	if _, err := uuid.Parse(experienceID); err != nil {
		servers, listErr := listFeaturedServers(ctx, s.httpClient, gatherings.ServiceURI, token)
		if listErr != nil {
			return nil, fmt.Errorf("list featured and creator experiences: %w", listErr)
		}
		server, ok := findFeaturedServer(servers, name)
		if !ok {
			return nil, fmt.Errorf("experience %q was not found", name)
		}
		displayName = server.Name
		experienceID = server.ExperienceID
		if experienceID == "" {
			if server.Address == "" {
				return nil, fmt.Errorf("experience %q has no experience ID or server address", server.Name)
			}
			return s.newTarget(selector, displayName, "", resolvedDestination{Address: server.Address, Transport: transportRakNet}, nil), nil
		}
	}
	id, err := uuid.Parse(experienceID)
	if err != nil {
		return nil, fmt.Errorf("experience %q returned invalid ID %q: %w", displayName, experienceID, err)
	}
	destination, err := joinExperience(ctx, s.httpClient, gatherings.ServiceURI, token, id)
	if err != nil {
		return nil, fmt.Errorf("join experience %q: %w", displayName, err)
	}
	if destination.Transport == transportRakNet {
		return s.newTarget(selector, displayName, id.String(), destination, nil), nil
	}

	var signaling signalingEnvironment
	if err := s.discovery.Environment(&signaling); err != nil {
		return nil, fmt.Errorf("load signaling environment: %w", err)
	}
	if err := validateSignalingServiceURI(signaling.ServiceURI); err != nil {
		return nil, fmt.Errorf("invalid signaling environment: %w", err)
	}
	connection, err := dialSignaling(ctx, s.httpClient, signaling.ServiceURI, token, destination.Transport)
	if err != nil {
		return nil, fmt.Errorf("connect %s signaling: %w", destination.Transport, err)
	}
	network := authenticatedNetherNet{NetherNet: minecraft.NetherNet{Signaling: connection}}
	return s.newTargetWithSignaling(selector, displayName, id.String(), destination, network, connection.Close), nil
}

func networkTarget(network minecraft.Network) func(*Target) {
	return func(target *Target) {
		target.Network = network
	}
}

func (s *serviceSession) newTarget(selector, name, experienceID string, destination resolvedDestination, configure func(*Target)) *Target {
	target := &Target{
		Selector:     selector,
		Name:         name,
		ExperienceID: experienceID,
		Address:      destination.Address,
		Transport:    destination.Transport,
		TokenSource:  s,
	}
	if configure != nil {
		configure(target)
	}
	target.closeFn = s.Close
	return target
}

func (s *serviceSession) newTargetWithSignaling(selector, name, experienceID string, destination resolvedDestination, network minecraft.Network, closeSignaling func() error) *Target {
	target := s.newTarget(selector, name, experienceID, destination, networkTarget(network))
	target.closeFn = func() error {
		return errors.Join(closeSignaling(), s.Close())
	}
	return target
}

type authenticatedNetherNet struct {
	minecraft.NetherNet
}

func (n authenticatedNetherNet) DialContextIdentity(ctx context.Context, address, token string, key *ecdsa.PrivateKey) (net.Conn, error) {
	n.Dialer.Identity = &nethernet.Identity{PrivateKey: key, Token: token, Domain: identityDomain}
	return n.DialContext(ctx, address)
}

func listFeaturedServers(ctx context.Context, client *http.Client, baseURI string, token *service.Token) ([]featuredServer, error) {
	endpoint, err := serviceURL(baseURI, "/api/v2.0/discovery/blob/client")
	if err != nil {
		return nil, err
	}
	var response dataEnvelope[struct {
		Items []struct {
			Title struct {
				Neutral string `json:"NEUTRAL"`
			} `json:"Title"`
			DisplayProperties struct {
				URL          string `json:"url"`
				Port         int    `json:"port"`
				ExperienceID string `json:"experienceId"`
			} `json:"DisplayProperties"`
		} `json:"Items"`
	}]
	if err := requestJSON(ctx, client, http.MethodPost, endpoint, nil, token, &response); err != nil {
		return nil, err
	}
	servers := make([]featuredServer, 0, len(response.Data.Items))
	for _, item := range response.Data.Items {
		address := ""
		if item.DisplayProperties.URL != "" && validPort(item.DisplayProperties.Port) {
			address = net.JoinHostPort(item.DisplayProperties.URL, strconv.Itoa(item.DisplayProperties.Port))
		}
		servers = append(servers, featuredServer{
			Name:         item.Title.Neutral,
			Address:      address,
			ExperienceID: item.DisplayProperties.ExperienceID,
		})
	}
	return servers, nil
}

func findFeaturedServer(servers []featuredServer, name string) (featuredServer, bool) {
	for _, server := range servers {
		if strings.EqualFold(server.Name, name) || server.ExperienceID == name {
			return server, true
		}
	}
	return featuredServer{}, false
}

func joinExperience(ctx context.Context, client *http.Client, baseURI string, token *service.Token, id uuid.UUID) (resolvedDestination, error) {
	endpoint, err := serviceURL(baseURI, "/api/v2.0/join/experience")
	if err != nil {
		return resolvedDestination{}, err
	}
	var response resultEnvelope[struct {
		NetworkProtocol string `json:"networkProtocol"`
		IPv4Address     string `json:"ipV4Address"`
		Port            int    `json:"port"`
		NetherNetID     string `json:"netherNetId"`
	}]
	if err := requestJSON(ctx, client, http.MethodPost, endpoint, map[string]any{"experienceId": id}, token, &response); err != nil {
		return resolvedDestination{}, err
	}
	result := response.Data
	if result.NetherNetID != "" {
		networkID, err := uuid.Parse(result.NetherNetID)
		if err != nil {
			return resolvedDestination{}, fmt.Errorf("invalid NetherNet ID %q: %w", result.NetherNetID, err)
		}
		transport := transportNetherNet
		if strings.EqualFold(result.NetworkProtocol, protocolJSONRPCName) || strings.EqualFold(result.NetworkProtocol, "NETHERNET_JSONRPC") {
			transport = transportJSONRPC
		}
		return resolvedDestination{Address: networkID.String(), Transport: transport}, nil
	}
	if net.ParseIP(result.IPv4Address) == nil {
		return resolvedDestination{}, fmt.Errorf("no usable destination returned for network protocol %q", result.NetworkProtocol)
	}
	if !validPort(result.Port) {
		return resolvedDestination{}, fmt.Errorf("invalid port %d", result.Port)
	}
	return resolvedDestination{Address: net.JoinHostPort(result.IPv4Address, strconv.Itoa(result.Port)), Transport: transportRakNet}, nil
}

type resultEnvelope[T any] struct {
	Data T `json:"result"`
}

type dataEnvelope[T any] struct {
	Data T `json:"data"`
}

func requestJSON(ctx context.Context, client *http.Client, method, requestURL string, payload any, token *service.Token, output any) error {
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("encode request: %w", err)
		}
		body = bytes.NewReader(encoded)
	}
	request, err := http.NewRequestWithContext(ctx, method, requestURL, body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	request.Header.Set("Accept", "*/*")
	request.Header.Set("Accept-Language", "en-US,en;q=0.5")
	request.Header.Set("Cache-Control", "no-cache")
	request.Header.Set("User-Agent", minecraftUserAgent)
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if token != nil {
		token.SetAuthHeader(request)
	}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, maxServiceResponse))
		return fmt.Errorf("service returned HTTP %s", response.Status)
	}
	limited := io.LimitReader(response.Body, maxServiceResponse+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if len(data) > maxServiceResponse {
		return fmt.Errorf("service response exceeds %d bytes", maxServiceResponse)
	}
	if err := json.Unmarshal(data, output); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func serviceURL(baseURI, path string) (string, error) {
	base, err := url.Parse(baseURI)
	if err != nil {
		return "", fmt.Errorf("parse service URI: %w", err)
	}
	if base.Scheme != "https" && base.Scheme != "wss" || base.Host == "" {
		return "", fmt.Errorf("service URI must be an absolute HTTPS or WSS URL")
	}
	base.Path = strings.TrimRight(base.Path, "/") + path
	base.RawQuery = ""
	base.Fragment = ""
	return base.String(), nil
}

func validateHTTPSServiceURI(value string) error {
	parsed, err := url.Parse(value)
	if err != nil {
		return err
	}
	if parsed.Scheme != "https" || parsed.Host == "" {
		return errors.New("service URI must be an absolute HTTPS URL")
	}
	return nil
}

func validateSignalingServiceURI(value string) error {
	_, err := serviceURL(value, "")
	return err
}

func validPort(port int) bool {
	return port >= 1 && port <= 65535
}
