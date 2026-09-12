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
