package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Logger is an interface for logging HTTP requests, responses, and authentication events
type Logger interface {
	LogRequest(ctx context.Context, method, url string, body []byte)
	LogResponse(ctx context.Context, statusCode int, headers http.Header, body []byte)
	LogAuth(ctx context.Context, message string, fields map[string]interface{})
}

// Client represents the Apple Device Management API client.
type Client struct {
	auth    *AppleOAuthClient
	baseURL string
	logger  Logger
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

// SetLogger sets the logger for the client
func (c *Client) SetLogger(logger Logger) {
	c.logger = logger
	if c.auth != nil {
		c.auth.logger = logger
	}
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
	var requestBody []byte
	if req.Body != nil {
		var err error
		requestBody, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}
		if err := req.Body.Close(); err != nil && c.logger != nil {
			c.logger.LogAuth(ctx, "Failed to close request body", map[string]interface{}{
				"error": err.Error(),
			})
		}
		req.Body = io.NopCloser(bytes.NewBuffer(requestBody))
	}

	if c.logger != nil {
		c.logger.LogRequest(ctx, req.Method, req.URL.String(), requestBody)
	}

	for {
		if err := c.auth.Authenticate(ctx, req); err != nil {
			return nil, fmt.Errorf("authentication failed: %w", err)
		}
		resp, err := c.auth.GetHTTPClient().Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusTooManyRequests {
			if c.logger != nil && resp.Body != nil {
				responseBody, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				if err := resp.Body.Close(); err != nil {
					c.logger.LogAuth(ctx, "Failed to close response body", map[string]interface{}{
						"error": err.Error(),
					})
				}
				c.logger.LogResponse(ctx, resp.StatusCode, resp.Header, responseBody)
				resp.Body = io.NopCloser(bytes.NewBuffer(responseBody))
			}
			return resp, nil
		}

		if c.logger != nil {
			responseBody := []byte{}
			if resp.Body != nil {
				responseBody, _ = io.ReadAll(resp.Body)
			}
			c.logger.LogResponse(ctx, resp.StatusCode, resp.Header, responseBody)
		}

		retryAfter := resp.Header.Get("Retry-After")

		if err := resp.Body.Close(); err != nil && c.logger != nil {
			c.logger.LogAuth(ctx, "Failed to close response body", map[string]interface{}{
				"error": err.Error(),
			})
		}

		if retryAfter != "" {
			seconds, err := time.ParseDuration(retryAfter + "s")
			if err == nil {
				if seconds > 60*time.Second {
					if c.logger != nil {
						c.logger.LogAuth(ctx, "Retry-After exceeded maximum wait time", map[string]interface{}{
							"retry_after_seconds": seconds.Seconds(),
							"max_wait_seconds":    60,
						})
					}
					return nil, fmt.Errorf("received 429 Too Many Requests with Retry-After of %v)", seconds)
				}

				if c.logger != nil {
					c.logger.LogAuth(ctx, "Rate limited, waiting before retry", map[string]interface{}{
						"retry_after_seconds": seconds.Seconds(),
					})
				}
				fmt.Printf("Received 429. Retrying after %s...\n", seconds)
				time.Sleep(seconds)
				continue
			}

			if c.logger != nil {
				c.logger.LogAuth(ctx, "Failed to parse Retry-After header", map[string]interface{}{
					"retry_after_header": retryAfter,
				})
			}
		}

		return nil, fmt.Errorf("received 429 Too Many Requests, and no valid Retry-After header")
	}
}
