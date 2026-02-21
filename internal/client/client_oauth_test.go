// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func generateTestP8Key(t *testing.T) []byte {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate test key: %v", err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatalf("failed to marshal key: %v", err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
}

func TestCreateClientAssertion_ValidKey(t *testing.T) {
	pemKey := generateTestP8Key(t)
	ts := &appleTokenSource{
		config: &ClientConfig{
			TeamID:     "TEAM123",
			ClientID:   "CLIENT456",
			KeyID:      "KEY789",
			PrivateKey: pemKey,
			Scope:      "business.api",
		},
	}

	assertion, err := ts.createClientAssertion()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if assertion == "" {
		t.Fatal("expected non-empty assertion")
	}

	token, _, err := jwt.NewParser().ParseUnverified(assertion, jwt.MapClaims{})
	if err != nil {
		t.Fatalf("failed to parse JWT: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("failed to cast claims")
	}

	if iss, _ := claims["iss"].(string); iss != "TEAM123" {
		t.Errorf("expected iss=TEAM123, got %s", iss)
	}
	if sub, _ := claims["sub"].(string); sub != "CLIENT456" {
		t.Errorf("expected sub=CLIENT456, got %s", sub)
	}

	aud, _ := claims["aud"].([]any)
	if len(aud) != 1 || aud[0] != audienceURL {
		t.Errorf("expected aud=[%s], got %v", audienceURL, aud)
	}

	kid, _ := token.Header["kid"].(string)
	if kid != "KEY789" {
		t.Errorf("expected kid=KEY789, got %s", kid)
	}

	if token.Method.Alg() != "ES256" {
		t.Errorf("expected algorithm ES256, got %s", token.Method.Alg())
	}
}

func TestCreateClientAssertion_InvalidKey(t *testing.T) {
	ts := &appleTokenSource{
		config: &ClientConfig{
			TeamID:     "TEAM123",
			ClientID:   "CLIENT456",
			KeyID:      "KEY789",
			PrivateKey: []byte("not-a-valid-key"),
			Scope:      "business.api",
		},
	}

	_, err := ts.createClientAssertion()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse private key") {
		t.Errorf("expected 'failed to parse private key' error, got %q", err.Error())
	}
}

func TestCreateOrGetAssertion_CachesAssertion(t *testing.T) {
	pemKey := generateTestP8Key(t)
	ts := &appleTokenSource{
		config: &ClientConfig{
			TeamID:     "TEAM123",
			ClientID:   "CLIENT456",
			KeyID:      "KEY789",
			PrivateKey: pemKey,
			Scope:      "business.api",
		},
	}

	first, err := ts.createOrGetAssertion()
	if err != nil {
		t.Fatalf("first call error: %v", err)
	}

	second, err := ts.createOrGetAssertion()
	if err != nil {
		t.Fatalf("second call error: %v", err)
	}

	if first != second {
		t.Error("expected cached assertion to be reused, but got different values")
	}
}

func TestTokenSource_Token_Success(t *testing.T) {
	pemKey := generateTestP8Key(t)

	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		ct := r.Header.Get("Content-Type")
		if ct != "application/x-www-form-urlencoded" {
			t.Errorf("expected form content type, got %s", ct)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"test-access-token","token_type":"Bearer","expires_in":3600,"scope":"business.api"}`))
	}))
	defer tokenServer.Close()

	ts := &appleTokenSource{
		config: &ClientConfig{
			TeamID:     "TEAM123",
			ClientID:   "CLIENT456",
			KeyID:      "KEY789",
			PrivateKey: pemKey,
			Scope:      "business.api",
		},
		tokenClient: tokenServer.Client(),
	}

	origURL := defaultTokenURL
	defer func() {
		if defaultTokenURL != origURL {
			t.Log("note: defaultTokenURL was not restored")
		}
	}()

	req, _ := http.NewRequest(http.MethodPost, tokenServer.URL, nil)
	_ = req

	ts.tokenClient = &http.Client{
		Transport: &rewriteTransport{
			base:    http.DefaultTransport,
			rewrite: tokenServer.URL,
		},
	}

	token, err := ts.Token()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.AccessToken != "test-access-token" {
		t.Errorf("expected access token 'test-access-token', got %s", token.AccessToken)
	}
	if token.TokenType != "Bearer" {
		t.Errorf("expected token type Bearer, got %s", token.TokenType)
	}
	if token.Expiry.Before(time.Now()) {
		t.Error("expected token expiry in the future")
	}
}

type rewriteTransport struct {
	base    http.RoundTripper
	rewrite string
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	req.URL.Host = strings.TrimPrefix(t.rewrite, "http://")
	return t.base.RoundTrip(req)
}

func TestTokenSource_Token_AuthError(t *testing.T) {
	pemKey := generateTestP8Key(t)

	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid_client","error_description":"The client credentials are invalid"}`))
	}))
	defer tokenServer.Close()

	ts := &appleTokenSource{
		config: &ClientConfig{
			TeamID:     "TEAM123",
			ClientID:   "CLIENT456",
			KeyID:      "KEY789",
			PrivateKey: pemKey,
			Scope:      "business.api",
		},
		tokenClient: &http.Client{
			Transport: &rewriteTransport{
				base:    http.DefaultTransport,
				rewrite: tokenServer.URL,
			},
		},
	}

	_, err := ts.Token()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid_client") {
		t.Errorf("expected 'invalid_client' in error, got %q", err.Error())
	}
}

func TestNewUUIDv4(t *testing.T) {
	uuidPattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		u := newUUIDv4()
		if !uuidPattern.MatchString(u) {
			t.Fatalf("UUID %q does not match v4 pattern", u)
		}
		if seen[u] {
			t.Fatalf("duplicate UUID generated: %s", u)
		}
		seen[u] = true
	}
}
