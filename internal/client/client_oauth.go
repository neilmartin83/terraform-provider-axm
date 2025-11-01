package client

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	defaultTokenURL    = "https://account.apple.com/auth/oauth2/token"
	audienceURL        = "https://account.apple.com/auth/oauth2/v2/token"
	AssertionExpiry    = 180 * 24 * time.Hour
	TokenRefreshBuffer = 5 * time.Minute
	assertionCacheDir  = ".axm/cache"
)

type AppleOAuthClient struct {
	config          *ClientConfig
	httpClient      *http.Client
	token           *TokenInfo
	assertion       string
	assertionExpiry time.Time
	mu              sync.RWMutex
}

type ClientConfig struct {
	ClientID   string `json:"client_id"`
	TeamID     string `json:"team_id"`
	KeyID      string `json:"key_id"`
	PrivateKey []byte `json:"private_key"`
	Scope      string `json:"scope"`
}

type TokenInfo struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresIn   int       `json:"expires_in"`
	Scope       string    `json:"scope"`
	ExpiresAt   time.Time `json:"-"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

type AuthErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
	ErrorURI         string `json:"error_uri,omitempty"`
}

type CachedAssertion struct {
	Assertion string    `json:"assertion"`
	ExpiresAt time.Time `json:"expires_at"`
	ClientID  string    `json:"client_id"`
	TeamID    string    `json:"team_id"`
	KeyID     string    `json:"key_id"`
}

type CachedToken struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresAt   time.Time `json:"expires_at"`
	Scope       string    `json:"scope"`
	ClientID    string    `json:"client_id"`
	TeamID      string    `json:"team_id"`
	KeyID       string    `json:"key_id"`
}

// NewAppleClient creates a new client with the provided configuration
func NewAppleOAuthClient(config *ClientConfig) (*AppleOAuthClient, error) {
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	client := &AppleOAuthClient{
		config:     config,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		mu:         sync.RWMutex{},
	}

	_ = client.loadCachedAssertion()
	_ = client.loadCachedToken()

	return client, nil
}

// CreateClientAssertion generates a signed JWT token
func (c *AppleOAuthClient) CreateClientAssertion() (string, error) {
	now := time.Now()

	claims := jwt.RegisteredClaims{
		Issuer:    c.config.TeamID,
		Subject:   c.config.ClientID,
		Audience:  jwt.ClaimStrings{audienceURL},
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(AssertionExpiry)),
		ID:        uuid.New().String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = c.config.KeyID

	key, err := jwt.ParseECPrivateKeyFromPEM(c.config.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	signedToken, err := token.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// GetHTTPClient returns the configured HTTP client
func (c *AppleOAuthClient) GetHTTPClient() *http.Client {
	return c.httpClient
}

// RequestNewToken gets a new access token using the client assertion
func (c *AppleOAuthClient) RequestNewToken(ctx context.Context) (*TokenInfo, error) {
	assertion, err := c.createOrGetAssertion()
	if err != nil {
		return nil, fmt.Errorf("failed to get valid assertion: %w", err)
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", c.config.ClientID)
	data.Set("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
	data.Set("client_assertion", assertion)
	data.Set("scope", c.config.Scope)

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		defaultTokenURL,
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("warning: failed to close response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		var apiErr AuthErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
			return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("token request failed: %s - %s", apiErr.Error, apiErr.ErrorDescription)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	token := &TokenInfo{
		AccessToken: tokenResp.AccessToken,
		TokenType:   tokenResp.TokenType,
		ExpiresIn:   tokenResp.ExpiresIn,
		Scope:       tokenResp.Scope,
		ExpiresAt:   time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}

	c.token = token
	_ = c.saveCachedToken()

	return token, nil
}

// GetValidToken returns a valid token, refreshing if necessary
func (c *AppleOAuthClient) GetValidToken(ctx context.Context) (*TokenInfo, error) {
	c.mu.RLock()
	token := c.token
	c.mu.RUnlock()

	if token != nil && time.Now().Before(token.ExpiresAt.Add(-TokenRefreshBuffer)) {
		return token, nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.token != nil && time.Now().Before(c.token.ExpiresAt.Add(-TokenRefreshBuffer)) {
		return c.token, nil
	}

	newToken, err := c.RequestNewToken(ctx)
	if err != nil {
		return nil, err
	}
	c.token = newToken

	return c.token, nil
}

// createOrGetAssertion is a lock-free version for use when already holding a lock
func (c *AppleOAuthClient) createOrGetAssertion() (string, error) {
	if c.assertion != "" && time.Now().Before(c.assertionExpiry.Add(-TokenRefreshBuffer)) {
		return c.assertion, nil
	}

	newAssertion, err := c.CreateClientAssertion()
	if err != nil {
		return "", fmt.Errorf("failed to create client assertion: %w", err)
	}

	c.assertion = newAssertion
	c.assertionExpiry = time.Now().Add(AssertionExpiry)

	_ = c.saveCachedAssertion()

	return c.assertion, nil
}

// IsTokenValid checks if the current token is valid
func (c *AppleOAuthClient) IsTokenValid() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.token == nil {
		return false
	}

	return time.Now().Before(c.token.ExpiresAt.Add(-TokenRefreshBuffer))
}

// Helper function to validate client configuration
func validateConfig(config *ClientConfig) error {
	if config.ClientID == "" {
		return fmt.Errorf("client_id is required")
	}
	if config.TeamID == "" {
		return fmt.Errorf("team_id is required")
	}
	if config.KeyID == "" {
		return fmt.Errorf("key_id is required")
	}
	if len(config.PrivateKey) == 0 {
		return fmt.Errorf("private_key is required")
	}
	if config.Scope == "" {
		return fmt.Errorf("scope is required")
	}
	return nil
}

// Authenticate adds authentication to an HTTP request
func (c *AppleOAuthClient) Authenticate(ctx context.Context, req *http.Request) error {
	token, err := c.GetValidToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get valid token: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	return nil
}

// getCacheFilePath returns the path to the assertion cache file
func (c *AppleOAuthClient) getCacheFilePath() (string, error) {
	var cacheDir string

	homeDir, err := os.UserHomeDir()
	if err != nil {
		cacheDir = filepath.Join(os.TempDir(), assertionCacheDir)
	} else {
		cacheDir = filepath.Join(homeDir, assertionCacheDir)
	}

	configHash := c.getConfigHash()
	cacheFile := filepath.Join(cacheDir, fmt.Sprintf("assertion_%s.json", configHash))

	return cacheFile, nil
}

// getTokenCacheFilePath returns the path to the token cache file
func (c *AppleOAuthClient) getTokenCacheFilePath() (string, error) {
	var cacheDir string

	homeDir, err := os.UserHomeDir()
	if err != nil {
		cacheDir = filepath.Join(os.TempDir(), assertionCacheDir)
	} else {
		cacheDir = filepath.Join(homeDir, assertionCacheDir)
	}

	configHash := c.getConfigHash()
	cacheFile := filepath.Join(cacheDir, fmt.Sprintf("token_%s.json", configHash))

	return cacheFile, nil
}

// getConfigHash creates a unique hash from the client configuration
func (c *AppleOAuthClient) getConfigHash() string {
	data := fmt.Sprintf("%s:%s:%s", c.config.ClientID, c.config.TeamID, c.config.KeyID)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])[:16] // Use first 16 chars
}

// loadCachedAssertion loads a cached assertion from disk if valid
func (c *AppleOAuthClient) loadCachedAssertion() error {
	cacheFile, err := c.getCacheFilePath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read cache file: %w", err)
	}

	var cached CachedAssertion
	if err := json.Unmarshal(data, &cached); err != nil {
		return fmt.Errorf("failed to unmarshal cache: %w", err)
	}

	if cached.ClientID != c.config.ClientID ||
		cached.TeamID != c.config.TeamID ||
		cached.KeyID != c.config.KeyID {
		return fmt.Errorf("cached assertion config mismatch")
	}

	if time.Now().Before(cached.ExpiresAt.Add(-TokenRefreshBuffer)) {
		c.assertion = cached.Assertion
		c.assertionExpiry = cached.ExpiresAt
		return nil
	}

	_ = os.Remove(cacheFile)
	return nil
}

// saveCachedAssertion saves the current assertion to disk
func (c *AppleOAuthClient) saveCachedAssertion() error {
	if c.assertion == "" {
		return nil
	}

	cacheFile, err := c.getCacheFilePath()
	if err != nil {
		return err
	}

	cacheDir := filepath.Dir(cacheFile)
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	cached := CachedAssertion{
		Assertion: c.assertion,
		ExpiresAt: c.assertionExpiry,
		ClientID:  c.config.ClientID,
		TeamID:    c.config.TeamID,
		KeyID:     c.config.KeyID,
	}

	data, err := json.MarshalIndent(cached, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	if err := os.WriteFile(cacheFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// loadCachedToken loads a cached token from disk if valid
func (c *AppleOAuthClient) loadCachedToken() error {
	cacheFile, err := c.getTokenCacheFilePath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read token cache file: %w", err)
	}

	var cached CachedToken
	if err := json.Unmarshal(data, &cached); err != nil {
		return fmt.Errorf("failed to unmarshal token cache: %w", err)
	}

	if cached.ClientID != c.config.ClientID ||
		cached.TeamID != c.config.TeamID ||
		cached.KeyID != c.config.KeyID {
		return fmt.Errorf("cached token config mismatch")
	}

	if time.Now().Before(cached.ExpiresAt.Add(-TokenRefreshBuffer)) {
		c.token = &TokenInfo{
			AccessToken: cached.AccessToken,
			TokenType:   cached.TokenType,
			Scope:       cached.Scope,
			ExpiresAt:   cached.ExpiresAt,
		}
		return nil
	}

	_ = os.Remove(cacheFile)
	return nil
}

// saveCachedToken saves the current token to disk
func (c *AppleOAuthClient) saveCachedToken() error {
	if c.token == nil {
		return nil
	}

	cacheFile, err := c.getTokenCacheFilePath()
	if err != nil {
		return err
	}

	cacheDir := filepath.Dir(cacheFile)
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	cached := CachedToken{
		AccessToken: c.token.AccessToken,
		TokenType:   c.token.TokenType,
		ExpiresAt:   c.token.ExpiresAt,
		Scope:       c.token.Scope,
		ClientID:    c.config.ClientID,
		TeamID:      c.config.TeamID,
		KeyID:       c.config.KeyID,
	}

	data, err := json.MarshalIndent(cached, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token cache: %w", err)
	}

	if err := os.WriteFile(cacheFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write token cache file: %w", err)
	}

	return nil
}
