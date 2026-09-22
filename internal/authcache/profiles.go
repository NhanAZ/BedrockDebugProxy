package authcache

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	defaultAccountName = "default"
	registryFileName   = "accounts.json"
	accountsDirectory  = "accounts"
	maxRegistryBytes   = 64 << 10
)

var accountNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,31}$`)

// AccountStore manages the small local registry that selects the account used
// by run. The legacy auth-token.json remains the default account so existing
// installations continue to work without a migration prompt.
type AccountStore struct {
	root         string
	legacyPath   string
	registryPath string
	accountsDir  string
}

type accountRegistry struct {
	Version  int                      `json:"version"`
	Default  string                   `json:"default,omitempty"`
	Accounts map[string]accountRecord `json:"accounts,omitempty"`
}

type accountRecord struct {
	Name            string    `json:"name"`
	CreatedAt       time.Time `json:"created_at,omitempty"`
	MaxCacheAgeSecs int64     `json:"max_cache_age_seconds,omitempty"`
}

// Account describes a locally configured account without exposing its token.
type Account struct {
	Name          string
	TokenPath     string
	Default       bool
	Cached        bool
	CacheState    string
	CacheModified time.Time
	CacheExpires  time.Time
	AccessExpires time.Time
	MaxCacheAge   time.Duration
}

// NewAccountStore returns the per-user account registry.
func NewAccountStore() (*AccountStore, error) {
	root, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("locate user configuration directory: %w", err)
	}
	return newAccountStore(filepath.Join(root, "BedrockDebugProxy")), nil
}

func newAccountStore(root string) *AccountStore {
	return &AccountStore{
		root:         root,
		legacyPath:   filepath.Join(root, "auth-token.json"),
		registryPath: filepath.Join(root, registryFileName),
		accountsDir:  filepath.Join(root, accountsDirectory),
	}
}

// Default returns the account selected for ordinary proxy runs. When no
// registry exists, the legacy cache is treated as the default account.
func (s *AccountStore) Default() (Account, error) {
	registry, err := s.load()
	if err != nil {
		return Account{}, err
	}
	key := registry.Default
	if key == "" {
		key = defaultAccountName
	}
	record, ok := registry.Accounts[key]
	if !ok {
		record = accountRecord{Name: defaultAccountName}
	}
	return s.accountStatus(key, record, key == registry.Default || (registry.Default == "" && key == defaultAccountName)), nil
}

// List returns accounts in stable display order. It never returns token data.
func (s *AccountStore) List() ([]Account, error) {
	registry, err := s.load()
	if err != nil {
		return nil, err
	}
	if len(registry.Accounts) == 0 {
		return nil, nil
	}
	keys := make([]string, 0, len(registry.Accounts))
	for key := range registry.Accounts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	accounts := make([]Account, 0, len(keys))
	for _, key := range keys {
		accounts = append(accounts, s.accountStatus(key, registry.Accounts[key], key == registry.Default))
	}
	sort.SliceStable(accounts, func(i, j int) bool {
		if accounts[i].Default != accounts[j].Default {
			return accounts[i].Default
		}
		return strings.ToLower(accounts[i].Name) < strings.ToLower(accounts[j].Name)
	})
	return accounts, nil
}

// TokenPath validates a local account label and returns its canonical key and
// token path. It does not modify the registry, which keeps failed device
// authentication from leaving an empty account behind.
func (s *AccountStore) TokenPath(name string) (string, string, error) {
	key, _, err := canonicalAccountName(name)
	if err != nil {
		return "", "", err
	}
	return key, s.tokenPath(key), nil
}

// Upsert records a successfully authenticated account. The first account
// automatically becomes the default. Existing defaults are preserved.
func (s *AccountStore) Upsert(name string, maxAge time.Duration, setMaxAge bool) error {
	key, label, err := canonicalAccountName(name)
	if err != nil {
		return err
	}
	registry, err := s.load()
	if err != nil {
		return err
	}
	if registry.Accounts == nil {
		registry.Accounts = make(map[string]accountRecord)
	}
	record, exists := registry.Accounts[key]
	if !exists {
		record = accountRecord{Name: label, CreatedAt: time.Now().UTC()}
	}
	if record.Name == "" {
		record.Name = label
	}
	if setMaxAge {
		record.MaxCacheAgeSecs = int64(maxAge / time.Second)
	}
	registry.Accounts[key] = record
	if registry.Default == "" {
		registry.Default = key
	}
	return s.save(registry)
}

// Use makes an existing account the default account for future runs.
func (s *AccountStore) Use(name string) (Account, error) {
	key, _, err := canonicalAccountName(name)
	if err != nil {
		return Account{}, err
	}
	registry, err := s.load()
	if err != nil {
		return Account{}, err
	}
	record, ok := registry.Accounts[key]
	if !ok {
		return Account{}, fmt.Errorf("account %q was not found; run bedrock-debug-proxy login %s first", name, name)
	}
	registry.Default = key
	if err := s.save(registry); err != nil {
		return Account{}, err
	}
	return s.accountStatus(key, record, true), nil
}

// Rename moves one cached account to a new local label without changing its
// token, cache age policy, or selected-default state. The destination must not
// already exist so an explicit rename cannot overwrite another account.
func (s *AccountStore) Rename(oldName, newName string) error {
	oldKey, _, err := canonicalAccountName(oldName)
	if err != nil {
		return err
	}
	newKey, newLabel, err := canonicalAccountName(newName)
	if err != nil {
		return err
	}
	if oldKey == newKey {
		return errors.New("old and new account names are the same")
	}
	registry, err := s.load()
	if err != nil {
		return err
	}
	record, ok := registry.Accounts[oldKey]
	if !ok {
		return fmt.Errorf("account %q was not found", oldName)
	}
	if _, exists := registry.Accounts[newKey]; exists {
		return fmt.Errorf("account %q already exists", newName)
	}
	oldPath := s.tokenPath(oldKey)
	newPath := s.tokenPath(newKey)
	info, err := os.Stat(oldPath)
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("account %q has no cached Microsoft authentication", oldName)
	}
	if err != nil {
		return fmt.Errorf("inspect account cache: %w", err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("account cache %q is not a regular file", oldName)
	}
	if err := os.MkdirAll(filepath.Dir(newPath), 0o700); err != nil {
		return fmt.Errorf("create account directory: %w", err)
	}
	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("rename account cache: %w", err)
	}
	rollback := func() {
		_ = os.Rename(newPath, oldPath)
	}
	record.Name = newLabel
	registry.Accounts[newKey] = record
	delete(registry.Accounts, oldKey)
	if registry.Default == oldKey {
		registry.Default = newKey
	}
	if err := s.save(registry); err != nil {
		rollback()
		return fmt.Errorf("save renamed account: %w", err)
	}
	return nil
}

// Remove deletes one account's local token and registry entry. It returns the
// selected replacement default, if one was chosen automatically.
func (s *AccountStore) Remove(name string) (removed bool, replacement string, err error) {
	key, _, err := canonicalAccountName(name)
	if err != nil {
		return false, "", err
	}
	registry, err := s.load()
	if err != nil {
		return false, "", err
	}
	if _, ok := registry.Accounts[key]; !ok {
		return false, "", fmt.Errorf("account %q was not found", name)
	}
	path := s.tokenPath(key)
	removed, err = Remove(path)
	if err != nil {
		return false, "", err
	}
	delete(registry.Accounts, key)
	if registry.Default == key {
		keys := make([]string, 0, len(registry.Accounts))
		for candidate := range registry.Accounts {
			keys = append(keys, candidate)
		}
		sort.Strings(keys)
		if len(keys) > 0 {
			registry.Default = keys[0]
			replacement = registry.Accounts[keys[0]].Name
		} else {
			registry.Default = ""
		}
	}
	if len(registry.Accounts) == 0 {
		if err := removeRegistry(s.registryPath); err != nil {
			return false, "", err
		}
		return removed, replacement, nil
	}
	if err := s.save(registry); err != nil {
		return false, "", err
	}
	return removed, replacement, nil
}

// CacheExpired reports whether the local max-age policy requires a fresh
// device authentication. Provider token expiry remains provider-controlled.
func (a Account) CacheExpired(now time.Time) bool {
	return a.MaxCacheAge > 0 && !a.CacheModified.IsZero() && !now.Before(a.CacheExpires)
}

func (s *AccountStore) accountStatus(key string, record accountRecord, isDefault bool) Account {
	path := s.tokenPath(key)
	status := Account{
		Name:        record.Name,
		TokenPath:   path,
		Default:     isDefault,
		MaxCacheAge: time.Duration(record.MaxCacheAgeSecs) * time.Second,
		CacheState:  "missing",
	}
	if status.Name == "" {
		status.Name = key
	}
	info, err := os.Stat(path)
	if err != nil {
		return status
	}
	status.Cached = true
	status.CacheModified = info.ModTime()
	status.CacheState = "cached"
	if status.MaxCacheAge > 0 {
		status.CacheExpires = info.ModTime().Add(status.MaxCacheAge)
		if !time.Now().Before(status.CacheExpires) {
			status.CacheState = "reauthentication required"
		}
	}
	if token, err := readToken(path); err == nil {
		status.AccessExpires = token.Expiry
	} else {
		status.CacheState = "invalid"
	}
	return status
}

func (s *AccountStore) tokenPath(key string) string {
	if key == defaultAccountName {
		return s.legacyPath
	}
	return filepath.Join(s.accountsDir, key+".json")
}

func (s *AccountStore) load() (accountRegistry, error) {
	registry := accountRegistry{Version: 1, Accounts: make(map[string]accountRecord)}
	file, err := os.Open(s.registryPath)
	if errors.Is(err, os.ErrNotExist) {
		if info, statErr := os.Stat(s.legacyPath); statErr == nil && info.Mode().IsRegular() {
			registry.Default = defaultAccountName
			registry.Accounts[defaultAccountName] = accountRecord{Name: defaultAccountName, CreatedAt: info.ModTime()}
		}
		return registry, nil
	}
	if err != nil {
		return registry, fmt.Errorf("open account registry: %w", err)
	}
	defer func() { _ = file.Close() }()
	if info, statErr := file.Stat(); statErr != nil {
		return registry, fmt.Errorf("inspect account registry: %w", statErr)
	} else if info.Size() > maxRegistryBytes {
		return registry, fmt.Errorf("account registry exceeds %d bytes", maxRegistryBytes)
	}
	if err := json.NewDecoder(file).Decode(&registry); err != nil {
		return registry, fmt.Errorf("decode account registry: %w", err)
	}
	if registry.Version != 1 {
		return registry, fmt.Errorf("unsupported account registry version %d", registry.Version)
	}
	if registry.Accounts == nil {
		registry.Accounts = make(map[string]accountRecord)
	}
	if registry.Default == "" && len(registry.Accounts) > 0 {
		return registry, errors.New("account registry has accounts but no default account")
	}
	if registry.Default != "" {
		key, _, err := canonicalAccountName(registry.Default)
		if err != nil || key != registry.Default {
			return registry, errors.New("account registry contains an unsafe default account key")
		}
		if _, ok := registry.Accounts[registry.Default]; !ok {
			return registry, errors.New("account registry selects a missing default account")
		}
	}
	for key, record := range registry.Accounts {
		canonicalKey, _, err := canonicalAccountName(key)
		if err != nil || canonicalKey != key {
			return registry, errors.New("account registry contains an unsafe account key")
		}
		if record.Name != "" {
			labelKey, _, labelErr := canonicalAccountName(record.Name)
			if labelErr != nil || labelKey != key {
				return registry, fmt.Errorf("account registry label for %q is invalid", key)
			}
		}
	}
	return registry, nil
}

func (s *AccountStore) save(registry accountRegistry) error {
	if registry.Version == 0 {
		registry.Version = 1
	}
	if registry.Accounts == nil {
		registry.Accounts = make(map[string]accountRecord)
	}
	if err := os.MkdirAll(s.root, 0o700); err != nil {
		return fmt.Errorf("create account directory: %w", err)
	}
	temporary, err := os.CreateTemp(s.root, ".accounts-*.tmp")
	if err != nil {
		return fmt.Errorf("create account registry temporary file: %w", err)
	}
	temporaryPath := temporary.Name()
	cleanup := func() {
		_ = temporary.Close()
		_ = os.Remove(temporaryPath)
	}
	if err := temporary.Chmod(0o600); err != nil {
		cleanup()
		return fmt.Errorf("restrict account registry permissions: %w", err)
	}
	if err := json.NewEncoder(temporary).Encode(registry); err != nil {
		cleanup()
		return fmt.Errorf("encode account registry: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		cleanup()
		return fmt.Errorf("sync account registry: %w", err)
	}
	if err := temporary.Close(); err != nil {
		_ = os.Remove(temporaryPath)
		return fmt.Errorf("close account registry: %w", err)
	}
	if err := os.Rename(temporaryPath, s.registryPath); err != nil {
		_ = os.Remove(temporaryPath)
		return fmt.Errorf("publish account registry: %w", err)
	}
	return nil
}

func removeRegistry(path string) error {
	err := os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("remove account registry: %w", err)
	}
	return nil
}

func canonicalAccountName(name string) (key, label string, err error) {
	label = strings.TrimSpace(name)
	if label == "" {
		return "", "", errors.New("account name is required")
	}
	if !accountNamePattern.MatchString(label) {
		return "", "", errors.New("account name must be 1-32 characters using letters, digits, dot, underscore, or hyphen")
	}
	return strings.ToLower(label), label, nil
}

// ParseMaxCacheAge parses the optional local cache policy accepted by login.
// Empty, zero, and "none" mean that the provider controls token refresh and
// the local cache has no additional age limit.
func ParseMaxCacheAge(value string) (time.Duration, error) {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" || value == "0" || value == "none" {
		return 0, nil
	}
	if strings.HasSuffix(value, "d") {
		number := strings.TrimSuffix(value, "d")
		days, err := strconv.ParseInt(number, 10, 64)
		maxDays := int64((1<<63 - 1) / int64(24*time.Hour))
		if err != nil || days <= 0 || days > maxDays {
			return 0, fmt.Errorf("invalid cache age %q; use a duration such as 7d or 168h", value)
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	duration, err := time.ParseDuration(value)
	if err != nil || duration <= 0 {
		return 0, fmt.Errorf("invalid cache age %q; use a duration such as 7d or 168h", value)
	}
	return duration, nil
}
