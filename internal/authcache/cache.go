package authcache

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/sandertv/gophertunnel/minecraft/auth"
	"golang.org/x/oauth2"
)

const maxCacheBytes = 64 << 10

type sourceFactory struct {
	device  func(io.Writer) oauth2.TokenSource
	refresh func(*oauth2.Token, io.Writer) oauth2.TokenSource
}

type Source struct {
	mu        sync.Mutex
	path      string
	output    io.Writer
	factory   sourceFactory
	source    oauth2.TokenSource
	fromCache bool
	announced bool
}

// DefaultPath returns the per-user path used for the Microsoft OAuth token cache.
func DefaultPath() (string, error) {
	root, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("locate user configuration directory: %w", err)
	}
	return filepath.Join(root, "BedrockDebugProxy", "auth-token.json"), nil
}

// Remove deletes the local Microsoft OAuth token cache. It does not revoke the
// Microsoft session or sign out any other application using the account.
func Remove(path string) (bool, error) {
	if path == "" {
		return false, errors.New("authentication cache path is empty")
	}
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("inspect authentication cache: %w", err)
	}
	if !info.Mode().IsRegular() {
		return false, fmt.Errorf("authentication cache %q is not a regular file", path)
	}
	if err := os.Remove(path); err != nil {
		return false, fmt.Errorf("remove authentication cache: %w", err)
	}
	return true, nil
}

// New returns a token source that reuses and refreshes a cached Microsoft OAuth
// token. Device authentication is requested only when no usable cached token is
// available. The cache contains a refresh token and must be treated as a secret.
func New(path string, output io.Writer) (*Source, error) {
	return newSource(path, output, sourceFactory{
		device:  auth.WriterTokenSource,
		refresh: auth.RefreshTokenSourceWriter,
	})
}

func newSource(path string, output io.Writer, factory sourceFactory) (*Source, error) {
	if path == "" {
		return nil, errors.New("authentication cache path is empty")
	}
	if output == nil {
		output = io.Discard
	}
	if factory.device == nil || factory.refresh == nil {
		return nil, errors.New("authentication token source factory is incomplete")
	}

	s := &Source{path: path, output: output, factory: factory}
	token, err := readToken(path)
	if errors.Is(err, os.ErrNotExist) {
		s.source = factory.device(output)
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	s.source = factory.refresh(token, output)
	s.fromCache = true
	return s, nil
}

func (s *Source) Token() (*oauth2.Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	token, err := s.source.Token()
	if err != nil && s.fromCache {
		_, _ = fmt.Fprintln(s.output, "Cached Microsoft authentication could not be refreshed. Starting device authentication.")
		s.source = s.factory.device(s.output)
		s.fromCache = false
		s.announced = false
		token, err = s.source.Token()
	}
	if err != nil {
		return nil, err
	}
	if token == nil {
		return nil, errors.New("authentication token source returned a nil token")
	}
	if err := writeToken(s.path, token); err != nil {
		return nil, err
	}
	if !s.announced {
		if s.fromCache {
			_, _ = fmt.Fprintln(s.output, "Using cached Microsoft authentication.")
		} else {
			_, _ = fmt.Fprintf(s.output, "Microsoft authentication cached at %s.\n", s.path)
		}
		s.announced = true
	}
	return token, nil
}

func readToken(path string) (*oauth2.Token, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("authentication cache %q is not a regular file", path)
	}
	if info.Size() > maxCacheBytes {
		return nil, fmt.Errorf("authentication cache %q exceeds %d bytes", path, maxCacheBytes)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open authentication cache: %w", err)
	}
	defer func() { _ = f.Close() }()

	var token oauth2.Token
	if err := json.NewDecoder(io.LimitReader(f, maxCacheBytes+1)).Decode(&token); err != nil {
		return nil, fmt.Errorf("decode authentication cache %q: %w", path, err)
	}
	if token.AccessToken == "" && token.RefreshToken == "" {
		return nil, fmt.Errorf("authentication cache %q contains no usable token", path)
	}
	return &token, nil
}

func writeToken(path string, token *oauth2.Token) error {
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return fmt.Errorf("create authentication cache directory: %w", err)
	}
	temporary, err := os.CreateTemp(directory, ".auth-token-*.tmp")
	if err != nil {
		return fmt.Errorf("create authentication cache temporary file: %w", err)
	}
	temporaryPath := temporary.Name()
	cleanup := func() {
		_ = temporary.Close()
		_ = os.Remove(temporaryPath)
	}
	if err := temporary.Chmod(0o600); err != nil {
		cleanup()
		return fmt.Errorf("restrict authentication cache permissions: %w", err)
	}
	//nolint:gosec // The documented per-user cache intentionally persists this OAuth secret.
	if err := json.NewEncoder(temporary).Encode(token); err != nil {
		cleanup()
		return fmt.Errorf("encode authentication cache: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		cleanup()
		return fmt.Errorf("sync authentication cache: %w", err)
	}
	if err := temporary.Close(); err != nil {
		_ = os.Remove(temporaryPath)
		return fmt.Errorf("close authentication cache: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err == nil {
		return nil
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		_ = os.Remove(temporaryPath)
		return fmt.Errorf("replace authentication cache: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		_ = os.Remove(temporaryPath)
		return fmt.Errorf("publish authentication cache: %w", err)
	}
	return nil
}
