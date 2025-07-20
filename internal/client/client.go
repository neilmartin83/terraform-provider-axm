package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client represents the Apple Device Management API client.
type Client struct {
	auth    *AppleOAuthClient
	baseURL string
}

// ErrorResponse represents the error details that an API returns in the response body whenever the API request isn’t successful.
type ErrorResponse struct {
	Errors []Error `json:"errors"`
}

// Error represents the details about an error that returns when an API request isn’t successful.
type Error struct {
	ID     string       `json:"id"`
	Status string       `json:"status"`
	Code   string       `json:"code"`
	Title  string       `json:"title"`
	Detail string       `json:"detail"`
	Source *ErrorSource `json:"source,omitempty"`
	Links  *ErrorLinks  `json:"links,omitempty"`
	Meta   interface{}  `json:"meta,omitempty"`
}

// ErrorSource represents one of two possible types of values — source.Parameter when a query parameter produces the error, or source.JsonPointer when a problem with the entity produces the error.
type ErrorSource struct {
	Pointer   string `json:"pointer,omitempty"`
	Parameter string `json:"parameter,omitempty"`
}

// ErrorLinks provides links related to the error.
type ErrorLinks struct {
	About      string                `json:"about,omitempty"`
	Associated *ErrorLinksAssociated `json:"associated,omitempty"`
}

// ErrorLinksAssociated provides additional information about associated errors.
type ErrorLinksAssociated struct {
	Href string                 `json:"href"`
	Meta map[string]interface{} `json:"meta,omitempty"`
}

// PagedDocumentLinks represents links related to the response document, including paging links.
type PagedDocumentLinks struct {
	First string `json:"first,omitempty"`
	Next  string `json:"next,omitempty"`
	Self  string `json:"self"`
}

// RelationshipLinks represents links related to the response document, including self-links.
type RelationshipLinks struct {
	Include string `json:"include,omitempty"`
	Related string `json:"related,omitempty"`
	Self    string `json:"self,omitempty"`
}

// ResourceLinks represents self-links to requested resources.
type ResourceLinks struct {
	Self string `json:"self,omitempty"`
}

// DocumentLinks represents self-links to documents that can contain information for one or more resources.
type DocumentLinks struct {
	Self string `json:"self"`
}

// Meta represents metadata in a response.
type Meta struct {
	Paging Paging `json:"paging"`
}

// Paging represents paging details, such as the total number of resources and the per-page limit.
type Paging struct {
	Limit      int    `json:"limit"`
	NextCursor string `json:"nextCursor,omitempty"`
	Total      int    `json:"total,omitempty"`
}

// Data represents a generic data structure used in relationships.
type Data struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// NewClient creates a new Client instance with the provided configuration.
func NewClient(baseURL, teamID, clientID, keyID, scope, p8Key string) (*Client, error) {
	config := &ClientConfig{
		TeamID:     teamID,
		ClientID:   clientID,
		KeyID:      keyID,
		Scope:      scope,
		PrivateKey: []byte(p8Key),
	}

	auth, err := NewAppleOAuthClient(config)
	if err != nil {
		return nil, err
	}

	return &Client{
		auth:    auth,
		baseURL: baseURL,
	}, nil
}

// handleErrorResponse processes error responses from the API.
func (c *Client) handleErrorResponse(resp *http.Response) error {
	var errResp ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		return fmt.Errorf("failed to decode error response: %w", err)
	}

	if len(errResp.Errors) > 0 {
		err := errResp.Errors[0]
		return fmt.Errorf("%s: %s (code: %s, status: %s, id: %s)",
			err.Title, err.Detail, err.Code, err.Status, err.ID)
	}

	return fmt.Errorf("unknown error occurred with status %d", resp.StatusCode)
}

// doRequest performs an authenticated HTTP request and handles rate limiting via Retry-After.
func (c *Client) doRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	for {
		if err := c.auth.Authenticate(ctx, req); err != nil {
			return nil, fmt.Errorf("authentication failed: %w", err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusTooManyRequests {
			return resp, nil
		}

		retryAfter := resp.Header.Get("Retry-After")
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("warning: failed to close response body: %v\n", err)
		}
		if retryAfter != "" {
			seconds, err := time.ParseDuration(retryAfter + "s")
			if err == nil {
				fmt.Printf("Received 429. Retrying after %s...\n", seconds)
				time.Sleep(seconds)
				continue
			}
		}
		return resp, fmt.Errorf("received 429 Too Many Requests, and no valid Retry-After header")
	}
}
