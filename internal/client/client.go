// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/oauth2"
)

const (
	maxRateLimitRetries   = 5
	maxRetryAfterDuration = 60 * time.Second
)

// Logger is an interface for logging HTTP requests, responses, and authentication events
type Logger interface {
	LogRequest(ctx context.Context, method, url string, body []byte)
	LogResponse(ctx context.Context, statusCode int, headers http.Header, body []byte)
	LogAuth(ctx context.Context, message string, fields map[string]any)
}

// Client represents the Apple Device Management API client.
type Client struct {
	httpClient  *http.Client
	tokenSource *appleTokenSource
	oauthTS     oauth2.TokenSource
	baseURL     string
	logger      Logger
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
	Meta   any          `json:"meta,omitempty"`
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
	Meta map[string]any `json:"meta,omitempty"`
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

	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	ts := newTokenSource(config)
	initialToken := ts.loadCachedOAuthToken()
	reusableTS := oauth2.ReuseTokenSource(initialToken, ts)

	return &Client{
		httpClient:  oauth2.NewClient(context.Background(), reusableTS),
		tokenSource: ts,
		oauthTS:     reusableTS,
		baseURL:     baseURL,
	}, nil
}

// SetLogger sets the logger for the client.
func (c *Client) SetLogger(logger Logger) {
	c.logger = logger
	if c.tokenSource != nil {
		c.tokenSource.logger = logger
	}
}

// TestAuth forces authentication and returns the JWT client assertion, its expiry, and the OAuth token.
func (c *Client) TestAuth() (assertion string, assertionExpiry time.Time, token *oauth2.Token, err error) {
	assertion, err = c.tokenSource.createOrGetAssertion()
	if err != nil {
		return "", time.Time{}, nil, fmt.Errorf("assertion failed: %w", err)
	}
	assertionExpiry = c.tokenSource.assertionExpiry
	token, err = c.oauthTS.Token()
	if err != nil {
		return assertion, assertionExpiry, nil, fmt.Errorf("token failed: %w", err)
	}
	return assertion, assertionExpiry, token, nil
}

// handleErrorResponse processes error responses from the API.
func (c *Client) handleErrorResponse(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("HTTP %d: failed to read error response body: %w", resp.StatusCode, err)
	}

	var errResp ErrorResponse
	if err := json.Unmarshal(body, &errResp); err != nil {
		// Response is not JSON (e.g. HTML error page from auth failure or wrong URL)
		snippet := string(body)
		if len(snippet) > 200 {
			snippet = snippet[:200] + "..."
		}
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, snippet)
	}

	if len(errResp.Errors) > 0 {
		e := errResp.Errors[0]
		return fmt.Errorf("%s: %s (code: %s, status: %s, id: %s)",
			e.Title, e.Detail, e.Code, e.Status, e.ID)
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
		_ = req.Body.Close()
		req.Body = io.NopCloser(bytes.NewBuffer(requestBody))
	}

	attempts := 0

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		if requestBody != nil {
			req.Body = io.NopCloser(bytes.NewReader(requestBody))
			req.ContentLength = int64(len(requestBody))
		}

		if c.logger != nil {
			c.logger.LogRequest(ctx, req.Method, req.URL.String(), requestBody)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusTooManyRequests {
			if c.logger != nil && resp.Body != nil {
				responseBody, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				_ = resp.Body.Close()
				c.logger.LogResponse(ctx, resp.StatusCode, resp.Header, responseBody)
				resp.Body = io.NopCloser(bytes.NewBuffer(responseBody))
			}
			return resp, nil
		}

		if resp.Body != nil {
			if c.logger != nil {
				responseBody, _ := io.ReadAll(resp.Body)
				c.logger.LogResponse(ctx, resp.StatusCode, resp.Header, responseBody)
			} else {
				_, _ = io.Copy(io.Discard, resp.Body)
			}
			_ = resp.Body.Close()
		}

		retryAfterDuration, err := parseRetryAfter(resp.Header.Get("Retry-After"))
		if err != nil {
			return nil, fmt.Errorf("received 429 Too Many Requests: %w", err)
		}

		if retryAfterDuration > maxRetryAfterDuration {
			return nil, fmt.Errorf("received 429 Too Many Requests with Retry-After of %v)", retryAfterDuration)
		}

		attempts++
		if attempts >= maxRateLimitRetries {
			return nil, fmt.Errorf("received 429 Too Many Requests after %d retries", attempts)
		}

		if c.logger != nil {
			c.logger.LogAuth(ctx, "Rate limited, waiting before retry", map[string]any{
				"retry_after_seconds": retryAfterDuration.Seconds(),
				"attempt":             attempts,
			})
		}
		if err := waitWithContext(ctx, retryAfterDuration); err != nil {
			return nil, err
		}
	}
}

func parseRetryAfter(header string) (time.Duration, error) {
	if header == "" {
		return 0, errors.New("missing Retry-After header")
	}

	if seconds, err := strconv.Atoi(header); err == nil {
		return time.Duration(seconds) * time.Second, nil
	}

	if retryTime, err := http.ParseTime(header); err == nil {
		delay := time.Until(retryTime)
		if delay <= 0 {
			return 0, errors.New("Retry-After time already elapsed")
		}
		return delay, nil
	}

	return 0, fmt.Errorf("invalid Retry-After header: %s", header)
}

func waitWithContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}

	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
