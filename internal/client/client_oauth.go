// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

const (
	defaultTokenURL      = "https://account.apple.com/auth/oauth2/token"
	audienceURL          = "https://account.apple.com/auth/oauth2/v2/token"
	assertionMaxLifetime = 180 * 24 * time.Hour
	tokenRefreshBuffer   = 5 * time.Minute
	assertionCacheDir    = ".axm/cache"
)

// ClientConfig holds the credentials and settings required to authenticate with the Apple API.
type ClientConfig struct {
	ClientID   string `json:"client_id"`
	TeamID     string `json:"team_id"`
	KeyID      string `json:"key_id"`
	PrivateKey []byte `json:"private_key"`
	Scope      string `json:"scope"`
}

// TokenResponse represents the JSON response from Apple's OAuth token endpoint.
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

// AuthErrorResponse represents an error response from Apple's OAuth token endpoint.
type AuthErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
	ErrorURI         string `json:"error_uri,omitempty"`
}

// CachedAssertion represents a JWT client assertion persisted to disk for reuse across provider runs.
type CachedAssertion struct {
	Assertion string    `json:"assertion"`
	ExpiresAt time.Time `json:"expires_at"`
	ClientID  string    `json:"client_id"`
	TeamID    string    `json:"team_id"`
	KeyID     string    `json:"key_id"`
}

// CachedToken represents an OAuth access token persisted to disk for reuse across provider runs.
type CachedToken struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresAt   time.Time `json:"expires_at"`
	Scope       string    `json:"scope"`
	ClientID    string    `json:"client_id"`
	TeamID      string    `json:"team_id"`
	KeyID       string    `json:"key_id"`
}

// appleTokenSource implements oauth2.TokenSource by creating JWT client assertions
// and exchanging them for access tokens at Apple's OAuth endpoint.
type appleTokenSource struct {
	config          *ClientConfig
	tokenClient     *http.Client
	assertion       string
	assertionExpiry time.Time
	logger          Logger
}

// Token creates a new access token by generating or reusing a JWT client assertion
// and exchanging it at Apple's token endpoint.
func (s *appleTokenSource) Token() (*oauth2.Token, error) {
	assertion, err := s.createOrGetAssertion()
	if err != nil {
		return nil, fmt.Errorf("failed to get valid assertion: %w", err)
	}

	if s.logger != nil {
		s.logger.LogAuth(context.Background(), "Requesting new access token", map[string]any{
			"reason": "token expired or missing",
		})
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", s.config.ClientID)
	data.Set("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
	data.Set("client_assertion", assertion)
	data.Set("scope", s.config.Scope)

	req, err := http.NewRequest(http.MethodPost, defaultTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.tokenClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		var apiErr AuthErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
			return nil, fmt.Errorf("token request failed with status %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("token request failed: %s - %s", apiErr.Error, apiErr.ErrorDescription)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	token := &oauth2.Token{
		AccessToken: tokenResp.AccessToken,
		TokenType:   tokenResp.TokenType,
		Expiry:      time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second).Add(-tokenRefreshBuffer),
	}

	_ = s.saveCachedToken(token)

	if s.logger != nil {
		s.logger.LogAuth(context.Background(), "Successfully obtained new access token", map[string]any{
			"expires_at": token.Expiry.Add(tokenRefreshBuffer),
		})
	}

	return token, nil
}

// createOrGetAssertion returns a valid JWT assertion, creating a new one if necessary.
func (s *appleTokenSource) createOrGetAssertion() (string, error) {
	if s.assertion != "" && time.Now().Before(s.assertionExpiry.Add(-tokenRefreshBuffer)) {
		if s.logger != nil {
			s.logger.LogAuth(context.Background(), "Using cached client assertion", map[string]any{
				"expires_at": s.assertionExpiry,
			})
		}
		return s.assertion, nil
	}

	if s.logger != nil {
		s.logger.LogAuth(context.Background(), "Creating new client assertion", map[string]any{
			"reason": "assertion expired or missing",
		})
	}

	newAssertion, err := s.createClientAssertion()
	if err != nil {
		return "", fmt.Errorf("failed to create client assertion: %w", err)
	}

	s.assertion = newAssertion
	s.assertionExpiry = time.Now().Add(assertionMaxLifetime)

	_ = s.saveCachedAssertion()

	if s.logger != nil {
		s.logger.LogAuth(context.Background(), "Successfully created new client assertion", map[string]any{
			"expires_at": s.assertionExpiry,
		})
	}

	return s.assertion, nil
}

// createClientAssertion generates a signed JWT client assertion for Apple's OAuth endpoint.
func (s *appleTokenSource) createClientAssertion() (string, error) {
	now := time.Now()

	claims := jwt.RegisteredClaims{
		Issuer:    s.config.TeamID,
		Subject:   s.config.ClientID,
		Audience:  jwt.ClaimStrings{audienceURL},
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(assertionMaxLifetime)),
		ID:        newUUIDv4(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = s.config.KeyID

	key, err := jwt.ParseECPrivateKeyFromPEM(s.config.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	signedToken, err := token.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// newUUIDv4 generates a random UUID v4 string using crypto/rand.
func newUUIDv4() string {
	var uuid [16]byte
	_, _ = rand.Read(uuid[:])
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	uuid[8] = (uuid[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16])
}

// validateConfig checks that all required fields are present in the client configuration.
func validateConfig(config *ClientConfig) error {
	if config.ClientID == "" {
		return errors.New("client_id is required")
	}
	if config.TeamID == "" {
		return errors.New("team_id is required")
	}
	if config.KeyID == "" {
		return errors.New("key_id is required")
	}
	if len(config.PrivateKey) == 0 {
		return errors.New("private_key is required")
	}
	if config.Scope == "" {
		return errors.New("scope is required")
	}
	return nil
}

// newTokenSource creates and initializes an appleTokenSource with disk-cached assertion.
func newTokenSource(config *ClientConfig) *appleTokenSource {
	ts := &appleTokenSource{
		config:      config,
		tokenClient: &http.Client{Timeout: 30 * time.Second},
	}
	_ = ts.loadCachedAssertion()
	return ts
}

// loadCachedOAuthToken loads a cached token from disk and returns it as an oauth2.Token.
// Returns nil if no valid cached token exists.
func (s *appleTokenSource) loadCachedOAuthToken() *oauth2.Token {
	cacheFile, err := s.getTokenCacheFilePath()
	if err != nil {
		return nil
	}

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			if s.logger != nil {
				s.logger.LogAuth(context.Background(), "No cached token found on disk", map[string]any{
					"cache_file": cacheFile,
				})
			}
		}
		return nil
	}

	var cached CachedToken
	if err := json.Unmarshal(data, &cached); err != nil {
		return nil
	}

	if cached.ClientID != s.config.ClientID ||
		cached.TeamID != s.config.TeamID ||
		cached.KeyID != s.config.KeyID {
		if s.logger != nil {
			s.logger.LogAuth(context.Background(), "Cached token config mismatch", map[string]any{
				"cache_file": cacheFile,
			})
		}
		return nil
	}

	if !time.Now().Before(cached.ExpiresAt.Add(-tokenRefreshBuffer)) {
		if s.logger != nil {
			s.logger.LogAuth(context.Background(), "Cached token expired, removing", map[string]any{
				"cache_file": cacheFile,
				"expires_at": cached.ExpiresAt,
			})
		}
		_ = os.Remove(cacheFile)
		return nil
	}

	if s.logger != nil {
		s.logger.LogAuth(context.Background(), "Loaded valid cached token from disk", map[string]any{
			"cache_file": cacheFile,
			"expires_at": cached.ExpiresAt,
		})
	}

	return &oauth2.Token{
		AccessToken: cached.AccessToken,
		TokenType:   cached.TokenType,
		Expiry:      cached.ExpiresAt.Add(-tokenRefreshBuffer),
	}
}

// getConfigHash creates a unique hash from the client configuration.
func (s *appleTokenSource) getConfigHash() string {
	data := fmt.Sprintf("%s:%s:%s", s.config.ClientID, s.config.TeamID, s.config.KeyID)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])[:16]
}

// getCacheFilePath returns the path to the assertion cache file.
func (s *appleTokenSource) getCacheFilePath() (string, error) {
	cacheDir := filepath.Join(os.TempDir(), assertionCacheDir)
	configHash := s.getConfigHash()
	return filepath.Join(cacheDir, fmt.Sprintf("assertion_%s.json", configHash)), nil
}

// getTokenCacheFilePath returns the path to the token cache file.
func (s *appleTokenSource) getTokenCacheFilePath() (string, error) {
	cacheDir := filepath.Join(os.TempDir(), assertionCacheDir)
	configHash := s.getConfigHash()
	return filepath.Join(cacheDir, fmt.Sprintf("token_%s.json", configHash)), nil
}

// loadCachedAssertion loads a cached assertion from disk if valid.
func (s *appleTokenSource) loadCachedAssertion() error {
	cacheFile, err := s.getCacheFilePath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			if s.logger != nil {
				s.logger.LogAuth(context.Background(), "No cached assertion found on disk", map[string]any{
					"cache_file": cacheFile,
				})
			}
			return nil
		}
		return fmt.Errorf("failed to read cache file: %w", err)
	}

	var cached CachedAssertion
	if err := json.Unmarshal(data, &cached); err != nil {
		return fmt.Errorf("failed to unmarshal cache: %w", err)
	}

	if cached.ClientID != s.config.ClientID ||
		cached.TeamID != s.config.TeamID ||
		cached.KeyID != s.config.KeyID {
		if s.logger != nil {
			s.logger.LogAuth(context.Background(), "Cached assertion config mismatch", map[string]any{
				"cache_file": cacheFile,
			})
		}
		return errors.New("cached assertion config mismatch")
	}

	if time.Now().Before(cached.ExpiresAt.Add(-tokenRefreshBuffer)) {
		s.assertion = cached.Assertion
		s.assertionExpiry = cached.ExpiresAt
		if s.logger != nil {
			s.logger.LogAuth(context.Background(), "Loaded valid cached assertion from disk", map[string]any{
				"cache_file": cacheFile,
				"expires_at": cached.ExpiresAt,
			})
		}
		return nil
	}

	if s.logger != nil {
		s.logger.LogAuth(context.Background(), "Cached assertion expired, removing", map[string]any{
			"cache_file": cacheFile,
			"expires_at": cached.ExpiresAt,
		})
	}
	_ = os.Remove(cacheFile)
	return nil
}

// saveCachedAssertion saves the current assertion to disk.
func (s *appleTokenSource) saveCachedAssertion() error {
	if s.assertion == "" {
		return nil
	}

	cacheFile, err := s.getCacheFilePath()
	if err != nil {
		return err
	}

	cacheDir := filepath.Dir(cacheFile)
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	cached := CachedAssertion{
		Assertion: s.assertion,
		ExpiresAt: s.assertionExpiry,
		ClientID:  s.config.ClientID,
		TeamID:    s.config.TeamID,
		KeyID:     s.config.KeyID,
	}

	data, err := json.MarshalIndent(cached, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	if err := os.WriteFile(cacheFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	if s.logger != nil {
		s.logger.LogAuth(context.Background(), "Saved assertion to disk cache", map[string]any{
			"cache_file": cacheFile,
			"expires_at": s.assertionExpiry,
		})
	}

	return nil
}

// saveCachedToken saves a token to disk.
func (s *appleTokenSource) saveCachedToken(token *oauth2.Token) error {
	cacheFile, err := s.getTokenCacheFilePath()
	if err != nil {
		return err
	}

	cacheDir := filepath.Dir(cacheFile)
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	cached := CachedToken{
		AccessToken: token.AccessToken,
		TokenType:   token.TokenType,
		ExpiresAt:   token.Expiry.Add(tokenRefreshBuffer),
		Scope:       s.config.Scope,
		ClientID:    s.config.ClientID,
		TeamID:      s.config.TeamID,
		KeyID:       s.config.KeyID,
	}

	data, err := json.MarshalIndent(cached, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token cache: %w", err)
	}

	if err := os.WriteFile(cacheFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write token cache file: %w", err)
	}

	if s.logger != nil {
		s.logger.LogAuth(context.Background(), "Saved token to disk cache", map[string]any{
			"cache_file": cacheFile,
			"expires_at": cached.ExpiresAt,
		})
	}

	return nil
}
