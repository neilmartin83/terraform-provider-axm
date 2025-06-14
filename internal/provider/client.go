package axm

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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

type OrgDevicesResponse struct {
	Data  []OrgDevice `json:"data"`
	Links Links       `json:"links"`
	Meta  Meta        `json:"meta"`
}

type OrgDeviceResponse struct {
	Data  OrgDevice `json:"data"`
	Links Links     `json:"links"`
}

type OrgDevice struct {
	Type          string          `json:"type"`
	ID            string          `json:"id"`
	Attributes    DeviceAttribute `json:"attributes"`
	Relationships Relationships   `json:"relationships"`
	Links         Links           `json:"links"`
}

type DeviceAttribute struct {
	SerialNumber       string   `json:"serialNumber"`
	AddedToOrgDateTime string   `json:"addedToOrgDateTime"`
	UpdatedDateTime    string   `json:"updatedDateTime"`
	DeviceModel        string   `json:"deviceModel"`
	ProductFamily      string   `json:"productFamily"`
	ProductType        string   `json:"productType"`
	DeviceCapacity     string   `json:"deviceCapacity"`
	PartNumber         string   `json:"partNumber"`
	OrderNumber        string   `json:"orderNumber"`
	Color              string   `json:"color"`
	Status             string   `json:"status"`
	OrderDateTime      string   `json:"orderDateTime"`
	IMEI               []string `json:"imei"`
	MEID               []string `json:"meid"`
	EID                string   `json:"eid"`
	PurchaseSourceID   string   `json:"purchaseSourceId"`
	PurchaseSourceType string   `json:"purchaseSourceType"`
}

type Relationships struct {
	AssignedServer AssignedServer `json:"assignedServer"`
}

type AssignedServer struct {
	Links Links `json:"links"`
}

type Links struct {
	Self    string `json:"self"`
	Related string `json:"related,omitempty"`
}

type Meta struct {
	Paging Paging `json:"paging"`
}

type Paging struct {
	Limit      int    `json:"limit"`
	NextCursor string `json:"nextCursor,omitempty"`
	Total      int    `json:"total,omitempty"`
}

func NewClient(baseURL, teamID, clientID, keyID, p8Key string) (*Client, error) {
	privKey, err := parsePrivateKey(p8Key)
	if err != nil {
		return nil, err
	}

	client := &Client{
		httpClient: http.DefaultClient,
		baseURL:    baseURL,
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
	now := time.Now().UTC()
	expiration := now.Add(180 * 24 * time.Hour)

	claims := jwt.MapClaims{
		"iss": c.teamID,
		"sub": c.clientID,
		"aud": "https://account.apple.com/auth/oauth2/v2/token",
		"iat": now.Unix(),
		"exp": expiration.Unix(),
		"jti": uuid.New().String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = c.keyID

	signedToken, err := token.SignedString(c.privateKey)
	if err != nil {
		return fmt.Errorf("failed to sign JWT: %w", err)
	}

	form := map[string]string{
		"grant_type":            "client_credentials",
		"client_id":             c.clientID,
		"client_assertion_type": "urn:ietf:params:oauth:client-assertion-type:jwt-bearer",
		"client_assertion":      signedToken,
		"scope":                 "business.api",
	}

	req, err := http.NewRequest("POST", "https://account.apple.com/auth/oauth2/v2/token", bytes.NewBufferString(encodeForm(form)))
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("auth HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return fmt.Errorf("auth failed: %s", string(body))
	}

	var respBody struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &respBody); err != nil {
		return fmt.Errorf("failed to parse auth response: %w", err)
	}

	c.token = respBody.AccessToken
	c.tokenExpiry = now.Add(time.Duration(respBody.ExpiresIn) * time.Second)

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

func (c *Client) GetOrgDevices(ctx context.Context) ([]OrgDevice, error) {
	var allDevices []OrgDevice
	nextCursor := ""
	limit := 100

	for {
		req, err := http.NewRequestWithContext(ctx, "GET",
			fmt.Sprintf("%s/v1/orgDevices", c.baseURL), nil)
		if err != nil {
			return nil, err
		}

		q := req.URL.Query()
		q.Add("limit", fmt.Sprintf("%d", limit))
		if nextCursor != "" {
			q.Add("cursor", nextCursor)
		}
		req.URL.RawQuery = q.Encode()

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
		req.Header.Set("Accept", "application/json")

		resp, err := c.doRequest(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("error %d: %s", resp.StatusCode, bodyBytes)
		}

		var response OrgDevicesResponse
		if err := json.Unmarshal(bodyBytes, &response); err != nil {
			return nil, fmt.Errorf("failed to decode response JSON: %w", err)
		}

		allDevices = append(allDevices, response.Data...)

		nextCursor = response.Meta.Paging.NextCursor
		if nextCursor == "" {
			break
		}
	}

	return allDevices, nil
}

func (c *Client) GetOrgDevice(ctx context.Context, id string) (*OrgDevice, error) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s/v1/orgDevices/%s", c.baseURL, id), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error %d: %s", resp.StatusCode, bodyBytes)
	}

	var response OrgDeviceResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return nil, fmt.Errorf("failed to decode response JSON: %w", err)
	}

	return &response.Data, nil
}
