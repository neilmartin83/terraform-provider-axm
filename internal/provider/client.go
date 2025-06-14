package axm

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Client struct {
	httpClient  *http.Client
	baseURL     string
	token       string
	tokenExpiry time.Time

	teamID     string
	clientID   string
	keyID      string
	privateKey *ecdsa.PrivateKey
}

func NewClient(teamID, clientID, keyID, p8Key string) (*Client, error) {
	privKey, err := parsePrivateKey(p8Key)
	if err != nil {
		return nil, err
	}

	client := &Client{
		httpClient: http.DefaultClient,
		baseURL:    "https://api.apple.com/",
		teamID:     teamID,
		clientID:   clientID,
		keyID:      keyID,
		privateKey: privKey,
	}

	if err := client.authenticate(); err != nil {
		return nil, err
	}

	return client, nil
}

func parsePrivateKey(pemStr string) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return key.(*ecdsa.PrivateKey), nil
}

func (c *Client) authenticate() error {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Issuer:    c.teamID,
		Subject:   c.clientID,
		Audience:  jwt.ClaimStrings{"https://appleid.apple.com"},
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(5 * time.Minute)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = c.keyID

	signedToken, err := token.SignedString(c.privateKey)
	if err != nil {
		return err
	}

	form := map[string]string{
		"grant_type":            "client_credentials",
		"client_id":             c.clientID,
		"client_assertion_type": "urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
		"client_assertion":      signedToken,
		"scope":                 "your-scope-if-required",
	}

	req, _ := http.NewRequest("POST", "https://appleid.apple.com/auth/oauth2/token", bytes.NewBufferString(encodeForm(form)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("auth failed: %s", string(b))
	}

	var respBody struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return err
	}

	c.token = respBody.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(respBody.ExpiresIn) * time.Second)
	return nil
}

func encodeForm(data map[string]string) string {
	var buf bytes.Buffer
	for k, v := range data {
		buf.WriteString(fmt.Sprintf("%s=%s&", k, v))
	}
	return buf.String()[:buf.Len()-1]
}

func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	if c.token == "" || time.Now().After(c.tokenExpiry.Add(-10*time.Second)) {
		if err := c.authenticate(); err != nil {
			return nil, fmt.Errorf("failed to refresh token: %w", err)
		}
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	return c.httpClient.Do(req)
}
