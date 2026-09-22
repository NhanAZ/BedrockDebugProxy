package authcache

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

type errorTokenSource struct {
	err error
}

func (s errorTokenSource) Token() (*oauth2.Token, error) {
	return nil, s.err
}

func TestSourcePersistsAndReusesToken(t *testing.T) {
	path := filepath.Join(t.TempDir(), "auth-token.json")
	want := &oauth2.Token{
		AccessToken:  "access",
		RefreshToken: "refresh",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(time.Hour).UTC().Round(0),
	}
	deviceCalls := 0
	factory := sourceFactory{
		device: func(io.Writer) oauth2.TokenSource {
			deviceCalls++
			return oauth2.StaticTokenSource(want)
		},
		refresh: func(token *oauth2.Token, _ io.Writer) oauth2.TokenSource {
			if token.RefreshToken != want.RefreshToken {
				t.Fatalf("cached refresh token = %q", token.RefreshToken)
			}
			return oauth2.StaticTokenSource(token)
		},
	}

	first, err := newSource(path, io.Discard, factory)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := first.Token(); err != nil {
		t.Fatal(err)
	}
	if deviceCalls != 1 {
		t.Fatalf("device calls = %d, want 1", deviceCalls)
	}
	if info, err := os.Stat(path); err != nil || info.Size() == 0 {
		t.Fatalf("cache info = %v, %v", info, err)
	}

	second, err := newSource(path, io.Discard, factory)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := second.Token(); err != nil {
		t.Fatal(err)
	}
	if deviceCalls != 1 {
		t.Fatalf("device calls after cache reuse = %d, want 1", deviceCalls)
	}
}

func TestSourceFallsBackToDeviceAuthenticationWhenRefreshFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "auth-token.json")
	if err := writeToken(path, &oauth2.Token{AccessToken: "expired", RefreshToken: "invalid"}); err != nil {
		t.Fatal(err)
	}
	replacement := &oauth2.Token{AccessToken: "replacement", RefreshToken: "replacement-refresh"}
	var output bytes.Buffer
	deviceCalls := 0
	source, err := newSource(path, &output, sourceFactory{
		device: func(io.Writer) oauth2.TokenSource {
			deviceCalls++
			return oauth2.StaticTokenSource(replacement)
		},
		refresh: func(*oauth2.Token, io.Writer) oauth2.TokenSource {
			return errorTokenSource{err: errors.New("refresh rejected")}
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	token, err := source.Token()
	if err != nil {
		t.Fatal(err)
	}
	if token.AccessToken != replacement.AccessToken || deviceCalls != 1 {
		t.Fatalf("token = %#v, device calls = %d", token, deviceCalls)
	}
	if !strings.Contains(output.String(), "Starting device authentication") {
		t.Fatalf("output = %q", output.String())
	}
	cached, err := readToken(path)
	if err != nil {
		t.Fatal(err)
	}
	if cached.RefreshToken != replacement.RefreshToken {
		t.Fatalf("cached token = %#v", cached)
	}
}

func TestSourceDoesNotRefreshCacheAgeForAnUnchangedToken(t *testing.T) {
	path := filepath.Join(t.TempDir(), "auth-token.json")
	token := &oauth2.Token{AccessToken: "access", RefreshToken: "refresh", Expiry: time.Now().Add(time.Hour).UTC().Round(0)}
	if err := writeToken(path, token); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-24 * time.Hour).UTC().Round(0)
	if err := os.Chtimes(path, old, old); err != nil {
		t.Fatal(err)
	}
	source, err := newSource(path, io.Discard, sourceFactory{
		device: func(io.Writer) oauth2.TokenSource {
			return oauth2.StaticTokenSource(token)
		},
		refresh: func(*oauth2.Token, io.Writer) oauth2.TokenSource {
			return oauth2.StaticTokenSource(token)
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := source.Token(); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !info.ModTime().Equal(old) {
		t.Fatalf("cache modified at %v, want unchanged %v", info.ModTime(), old)
	}
}

func TestReadTokenRejectsOversizedCache(t *testing.T) {
	path := filepath.Join(t.TempDir(), "auth-token.json")
	if err := os.WriteFile(path, bytes.Repeat([]byte{'x'}, maxCacheBytes+1), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := readToken(path); err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("readToken() error = %v", err)
	}
}

func TestRemoveDeletesCacheAndHandlesMissingPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "auth-token.json")
	if err := writeToken(path, &oauth2.Token{RefreshToken: "refresh"}); err != nil {
		t.Fatal(err)
	}
	removed, err := Remove(path)
	if err != nil || !removed {
		t.Fatalf("Remove() = %v, %v", removed, err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("cache stat error = %v", err)
	}
	removed, err = Remove(path)
	if err != nil || removed {
		t.Fatalf("Remove() missing = %v, %v", removed, err)
	}
}

func TestRemoveRejectsNonRegularPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "auth-token.json")
	if err := os.Mkdir(path, 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := Remove(path); err == nil || !strings.Contains(err.Error(), "not a regular file") {
		t.Fatalf("Remove() error = %v", err)
	}
}

func TestRewriteDeviceAuthPromptProvidesClickableLoginURL(t *testing.T) {
	got, ok := rewriteDeviceAuthPrompt("Authenticate at https://www.microsoft.com/link using the code MQR5BATV.\n")
	if !ok {
		t.Fatal("rewriteDeviceAuthPrompt() did not recognize the gophertunnel prompt")
	}
	want := "Authenticate at https://login.live.com/oauth20_remoteconnect.srf?otc=MQR5BATV (fallback: https://www.microsoft.com/link with code MQR5BATV)\n"
	if got != want {
		t.Fatalf("rewritten prompt = %q, want %q", got, want)
	}
}

func TestRewriteDeviceAuthPromptLeavesOtherOutputUnchanged(t *testing.T) {
	for _, input := range []string{
		"Authentication successful.\n",
		"Authenticate at https://www.microsoft.com/link without a code.\n",
		"Authenticate at https://www.microsoft.com/link using the code .\n",
	} {
		if got, ok := rewriteDeviceAuthPrompt(input); ok || got != "" {
			t.Fatalf("rewriteDeviceAuthPrompt(%q) = %q, %v; want no rewrite", input, got, ok)
		}
	}
}

func TestClickableAuthWriterRewritesPrompt(t *testing.T) {
	var output bytes.Buffer
	writer := newClickableAuthWriter(&output)
	input := []byte("Authenticate at https://www.microsoft.com/link using the code MQR5BATV.\n")
	n, err := writer.Write(input)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(input) {
		t.Fatalf("Write() bytes = %d, want %d", n, len(input))
	}
	if !strings.Contains(output.String(), "https://login.live.com/oauth20_remoteconnect.srf?otc=MQR5BATV") {
		t.Fatalf("writer output = %q", output.String())
	}
}

func TestAccountStoreTreatsLegacyCacheAsDefault(t *testing.T) {
	root := t.TempDir()
	store := newAccountStore(root)
	if err := writeToken(filepath.Join(root, "auth-token.json"), &oauth2.Token{RefreshToken: "legacy"}); err != nil {
		t.Fatal(err)
	}
	account, err := store.Default()
	if err != nil {
		t.Fatal(err)
	}
	if account.Name != "default" || !account.Default || account.TokenPath != filepath.Join(root, "auth-token.json") {
		t.Fatalf("default account = %#v", account)
	}
	accounts, err := store.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 1 || accounts[0].Name != "default" || !accounts[0].Cached {
		t.Fatalf("accounts = %#v", accounts)
	}
}

func TestAccountStoreAddsSwitchesAndRemovesAccounts(t *testing.T) {
	root := t.TempDir()
	store := newAccountStore(root)
	if err := writeToken(filepath.Join(root, "accounts", "sampleaccount.json"), &oauth2.Token{RefreshToken: "sr"}); err != nil {
		t.Fatal(err)
	}
	if err := store.Upsert("SampleAccount", 0, false); err != nil {
		t.Fatal(err)
	}
	if err := writeToken(filepath.Join(root, "accounts", "nhanaz.json"), &oauth2.Token{RefreshToken: "nhan"}); err != nil {
		t.Fatal(err)
	}
	if err := store.Upsert("NhanAZ", 7*24*time.Hour, true); err != nil {
		t.Fatal(err)
	}
	defaultAccount, err := store.Default()
	if err != nil {
		t.Fatal(err)
	}
	if defaultAccount.Name != "SampleAccount" {
		t.Fatalf("default account after add = %#v", defaultAccount)
	}
	if _, err := store.Use("NhanAZ"); err != nil {
		t.Fatal(err)
	}
	defaultAccount, err = store.Default()
	if err != nil {
		t.Fatal(err)
	}
	if defaultAccount.Name != "NhanAZ" || defaultAccount.MaxCacheAge != 7*24*time.Hour {
		t.Fatalf("selected account = %#v", defaultAccount)
	}
	removed, replacement, err := store.Remove("NhanAZ")
	if err != nil || !removed || replacement != "SampleAccount" {
		t.Fatalf("Remove() = %v, %q, %v", removed, replacement, err)
	}
	defaultAccount, err = store.Default()
	if err != nil {
		t.Fatal(err)
	}
	if defaultAccount.Name != "SampleAccount" {
		t.Fatalf("replacement default = %#v", defaultAccount)
	}
}

func TestAccountStorePreservesLegacyAccountWhenAddingNamedAccount(t *testing.T) {
	root := t.TempDir()
	store := newAccountStore(root)
	if err := writeToken(filepath.Join(root, "auth-token.json"), &oauth2.Token{RefreshToken: "legacy"}); err != nil {
		t.Fatal(err)
	}
	if err := writeToken(filepath.Join(root, "accounts", "nhanaz.json"), &oauth2.Token{RefreshToken: "named"}); err != nil {
		t.Fatal(err)
	}
	if err := store.Upsert("NhanAZ", 0, false); err != nil {
		t.Fatal(err)
	}
	accounts, err := store.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 2 || accounts[0].Name != "default" || !accounts[0].Default || accounts[1].Name != "NhanAZ" {
		t.Fatalf("accounts = %#v", accounts)
	}
}

func TestAccountStoreRenamesLegacyCacheAndPreservesDefault(t *testing.T) {
	root := t.TempDir()
	store := newAccountStore(root)
	legacyPath := filepath.Join(root, "auth-token.json")
	if err := writeToken(legacyPath, &oauth2.Token{RefreshToken: "legacy"}); err != nil {
		t.Fatal(err)
	}
	if err := store.Rename("default", "SampleAccount"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(legacyPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("legacy cache still exists, stat error = %v", err)
	}
	namedPath := filepath.Join(root, "accounts", "sampleaccount.json")
	if _, err := os.Stat(namedPath); err != nil {
		t.Fatalf("named cache missing: %v", err)
	}
	account, err := store.Default()
	if err != nil || account.Name != "SampleAccount" || account.TokenPath != namedPath {
		t.Fatalf("renamed default = %#v, %v", account, err)
	}
	cached, err := readToken(namedPath)
	if err != nil || cached.RefreshToken != "legacy" {
		t.Fatalf("renamed token = %#v, %v", cached, err)
	}
}

func TestAccountStoreRenameRejectsExistingDestination(t *testing.T) {
	root := t.TempDir()
	store := newAccountStore(root)
	if err := writeToken(filepath.Join(root, "auth-token.json"), &oauth2.Token{RefreshToken: "legacy"}); err != nil {
		t.Fatal(err)
	}
	if err := writeToken(filepath.Join(root, "accounts", "nhanaz.json"), &oauth2.Token{RefreshToken: "named"}); err != nil {
		t.Fatal(err)
	}
	if err := store.Upsert("NhanAZ", 0, false); err != nil {
		t.Fatal(err)
	}
	if err := store.Rename("default", "NhanAZ"); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("Rename() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "auth-token.json")); err != nil {
		t.Fatalf("legacy cache after rejected rename: %v", err)
	}
}

func TestAccountStoreRejectsUnsafeNames(t *testing.T) {
	store := newAccountStore(t.TempDir())
	for _, name := range []string{"", "..", "bad/name", "bad name", strings.Repeat("a", 33)} {
		if _, _, err := store.TokenPath(name); err == nil {
			t.Fatalf("TokenPath(%q) unexpectedly succeeded", name)
		}
	}
}

func TestAccountStoreRejectsUnsafeRegistryKeys(t *testing.T) {
	root := t.TempDir()
	registry := `{"version":1,"default":"../escape","accounts":{"../escape":{"name":"../escape"}}}`
	if err := os.WriteFile(filepath.Join(root, registryFileName), []byte(registry), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := newAccountStore(root).List(); err == nil {
		t.Fatal("List accepted an unsafe registry key")
	}
}

func TestParseMaxCacheAge(t *testing.T) {
	for _, test := range []struct {
		value string
		want  time.Duration
	}{
		{"", 0}, {"none", 0}, {"7d", 7 * 24 * time.Hour}, {"168h", 7 * 24 * time.Hour},
	} {
		got, err := ParseMaxCacheAge(test.value)
		if err != nil || got != test.want {
			t.Fatalf("ParseMaxCacheAge(%q) = %v, %v; want %v", test.value, got, err, test.want)
		}
	}
	for _, value := range []string{"not-a-duration", "7days", "999999999999999999999d"} {
		if _, err := ParseMaxCacheAge(value); err == nil {
			t.Fatalf("ParseMaxCacheAge accepted invalid duration %q", value)
		}
	}
}

func TestNewSourceForceDeviceIgnoresExistingCache(t *testing.T) {
	path := filepath.Join(t.TempDir(), "auth-token.json")
	if err := writeToken(path, &oauth2.Token{RefreshToken: "cached"}); err != nil {
		t.Fatal(err)
	}
	deviceCalls := 0
	refreshCalls := 0
	source, err := newSourceWithOptions(path, io.Discard, sourceFactory{
		device: func(io.Writer) oauth2.TokenSource {
			deviceCalls++
			return oauth2.StaticTokenSource(&oauth2.Token{RefreshToken: "fresh"})
		},
		refresh: func(*oauth2.Token, io.Writer) oauth2.TokenSource {
			refreshCalls++
			return oauth2.StaticTokenSource(&oauth2.Token{RefreshToken: "refreshed"})
		},
	}, Options{ForceDevice: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := source.Token(); err != nil {
		t.Fatal(err)
	}
	if deviceCalls != 1 || refreshCalls != 0 {
		t.Fatalf("device calls = %d, refresh calls = %d", deviceCalls, refreshCalls)
	}
	cached, err := readToken(path)
	if err != nil || cached.RefreshToken != "fresh" {
		t.Fatalf("cached token = %#v, error = %v", cached, err)
	}
}
