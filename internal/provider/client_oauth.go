package axm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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
)

type AppleOAuthClient struct {
	config     *ClientConfig
	httpClient *http.Client
	token      *TokenInfo
	mu         sync.RWMutex
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

type JWTClaims struct {
	Issuer    string `json:"iss"`
	Subject   string `json:"sub"`
	Audience  string `json:"aud"`
	ExpiresAt int64  `json:"exp"`
	IssuedAt  int64  `json:"iat"`
	JTI       string `json:"jti"`
}

type TokenRequest struct {
	GrantType           string `json:"grant_type"`
	ClientID            string `json:"client_id"`
	ClientAssertion     string `json:"client_assertion"`
	ClientAssertionType string `json:"client_assertion_type"`
	Scope               string `json:"scope"`
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

// NewAppleClient creates a new client with the provided configuration
func NewAppleOAuthClient(config *ClientConfig) (*AppleOAuthClient, error) {
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &AppleOAuthClient{
		config:     config,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		mu:         sync.RWMutex{},
	}, nil
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

// RequestNewToken gets a new access token using the client assertion
func (c *AppleOAuthClient) RequestNewToken(ctx context.Context) (*TokenInfo, error) {
	assertion, err := c.CreateClientAssertion()
	if err != nil {
		return nil, fmt.Errorf("failed to create client assertion: %w", err)
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

	if c.token == nil || time.Now().After(c.token.ExpiresAt.Add(-TokenRefreshBuffer)) {
		newToken, err := c.RequestNewToken(ctx)
		if err != nil {
			return nil, err
		}
		c.token = newToken
	}

	return c.token, nil
}

// IsTokenValid checks if the current token is valid
func (c *AppleOAuthClient) IsTokenValid() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.token == nil {
		return false
	}

	return time.Now().Before(c.token.ExpiresAt.Add(-5 * time.Minute))
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
