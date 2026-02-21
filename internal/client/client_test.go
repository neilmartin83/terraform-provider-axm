// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  ClientConfig
		wantErr string
	}{
		{
			name:    "missing_client_id",
			config:  ClientConfig{TeamID: "team", KeyID: "key", PrivateKey: []byte("pk"), Scope: "business.api"},
			wantErr: "client_id is required",
		},
		{
			name:    "missing_team_id",
			config:  ClientConfig{ClientID: "client", KeyID: "key", PrivateKey: []byte("pk"), Scope: "business.api"},
			wantErr: "team_id is required",
		},
		{
			name:    "missing_key_id",
			config:  ClientConfig{ClientID: "client", TeamID: "team", PrivateKey: []byte("pk"), Scope: "business.api"},
			wantErr: "key_id is required",
		},
		{
			name:    "missing_private_key",
			config:  ClientConfig{ClientID: "client", TeamID: "team", KeyID: "key", Scope: "business.api"},
			wantErr: "private_key is required",
		},
		{
			name:    "missing_scope",
			config:  ClientConfig{ClientID: "client", TeamID: "team", KeyID: "key", PrivateKey: []byte("pk")},
			wantErr: "scope is required",
		},
		{
			name:   "valid_config",
			config: ClientConfig{ClientID: "client", TeamID: "team", KeyID: "key", PrivateKey: []byte("pk"), Scope: "business.api"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(&tt.config)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func TestParseRetryAfter(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		wantDur  time.Duration
		wantErr  string
		tolerance time.Duration
	}{
		{
			name:    "integer_seconds",
			header:  "5",
			wantDur: 5 * time.Second,
		},
		{
			name:    "zero_seconds",
			header:  "0",
			wantDur: 0,
		},
		{
			name:    "empty_header",
			header:  "",
			wantErr: "missing Retry-After header",
		},
		{
			name:    "invalid_value",
			header:  "abc",
			wantErr: "invalid Retry-After header",
		},
		{
			name:      "http_date_in_future",
			header:    time.Now().Add(30 * time.Second).UTC().Format(http.TimeFormat),
			wantDur:   30 * time.Second,
			tolerance: 5 * time.Second,
		},
		{
			name:    "http_date_in_past",
			header:  time.Now().Add(-10 * time.Second).UTC().Format(http.TimeFormat),
			wantErr: "already elapsed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dur, err := parseRetryAfter(tt.header)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.tolerance > 0 {
				diff := dur - tt.wantDur
				if diff < 0 {
					diff = -diff
				}
				if diff > tt.tolerance {
					t.Fatalf("expected duration ~%v (±%v), got %v", tt.wantDur, tt.tolerance, dur)
				}
			} else {
				if dur != tt.wantDur {
					t.Fatalf("expected duration %v, got %v", tt.wantDur, dur)
				}
			}
		})
	}
}

func TestHandleErrorResponse(t *testing.T) {
	c := &Client{}

	t.Run("json_error_response", func(t *testing.T) {
		body := `{"errors":[{"id":"err-1","status":"404","code":"NOT_FOUND","title":"Resource Not Found","detail":"The device was not found"}]}`
		resp := &http.Response{
			StatusCode: 404,
			Body:       io.NopCloser(strings.NewReader(body)),
		}
		err := c.handleErrorResponse(resp)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		for _, want := range []string{"Resource Not Found", "The device was not found", "NOT_FOUND", "404", "err-1"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("expected error to contain %q, got %q", want, err.Error())
			}
		}
	})

	t.Run("empty_errors_array", func(t *testing.T) {
		body := `{"errors":[]}`
		resp := &http.Response{
			StatusCode: 500,
			Body:       io.NopCloser(strings.NewReader(body)),
		}
		err := c.handleErrorResponse(resp)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "unknown error occurred") {
			t.Errorf("expected 'unknown error occurred', got %q", err.Error())
		}
	})

	t.Run("non_json_response", func(t *testing.T) {
		body := `<html><body>Unauthorized</body></html>`
		resp := &http.Response{
			StatusCode: 401,
			Body:       io.NopCloser(strings.NewReader(body)),
		}
		err := c.handleErrorResponse(resp)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "HTTP 401") {
			t.Errorf("expected 'HTTP 401' in error, got %q", err.Error())
		}
		if !strings.Contains(err.Error(), "Unauthorized") {
			t.Errorf("expected body snippet in error, got %q", err.Error())
		}
	})

	t.Run("long_non_json_response", func(t *testing.T) {
		body := strings.Repeat("x", 300)
		resp := &http.Response{
			StatusCode: 500,
			Body:       io.NopCloser(strings.NewReader(body)),
		}
		err := c.handleErrorResponse(resp)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "...") {
			t.Errorf("expected truncation marker '...' in error, got %q", err.Error())
		}
	})

	t.Run("unreadable_body", func(t *testing.T) {
		resp := &http.Response{
			StatusCode: 500,
			Body:       io.NopCloser(&errorReader{}),
		}
		err := c.handleErrorResponse(resp)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "failed to read error response body") {
			t.Errorf("expected read error message, got %q", err.Error())
		}
	})
}

type errorReader struct{}

func (r *errorReader) Read(_ []byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

func TestDoRequest_RateLimitRetry(t *testing.T) {
	var requestCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := requestCount.Add(1)
		if count <= 2 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	req, _ := http.NewRequest(http.MethodGet, server.URL+"/test", nil)
	resp, err := c.doRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	if got := requestCount.Load(); got != 3 {
		t.Fatalf("expected 3 requests, got %d", got)
	}
}

func TestDoRequest_RateLimitExceedsMaxRetries(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	req, _ := http.NewRequest(http.MethodGet, server.URL+"/test", nil)
	_, err := c.doRequest(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "after 5 retries") {
		t.Fatalf("expected max retries error, got %q", err.Error())
	}
}

func TestDoRequest_RateLimitExceedsMaxDuration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "120")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	req, _ := http.NewRequest(http.MethodGet, server.URL+"/test", nil)
	_, err := c.doRequest(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "Retry-After of") {
		t.Fatalf("expected duration exceeded error, got %q", err.Error())
	}
}

func TestDoRequest_RateLimitMissingHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	req, _ := http.NewRequest(http.MethodGet, server.URL+"/test", nil)
	_, err := c.doRequest(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "missing Retry-After header") {
		t.Fatalf("expected missing header error, got %q", err.Error())
	}
}

func TestDoRequest_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/test", nil)
	_, err := c.doRequest(ctx, req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDoRequest_PreservesRequestBody(t *testing.T) {
	var requestCount atomic.Int32
	var bodiesReceived []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		bodiesReceived = append(bodiesReceived, string(body))
		count := requestCount.Add(1)
		if count <= 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	reqBody := `{"test":"data"}`
	req, _ := http.NewRequest(http.MethodPost, server.URL+"/test", bytes.NewReader([]byte(reqBody)))
	resp, err := c.doRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if len(bodiesReceived) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(bodiesReceived))
	}
	for i, body := range bodiesReceived {
		if body != reqBody {
			t.Errorf("request %d: expected body %q, got %q", i, reqBody, body)
		}
	}
}

func TestWaitWithContext(t *testing.T) {
	tests := []struct {
		name    string
		dur     time.Duration
		cancel  bool
		wantErr bool
	}{
		{name: "zero_duration", dur: 0},
		{name: "negative_duration", dur: -1 * time.Second},
		{name: "normal_wait", dur: 10 * time.Millisecond},
		{name: "context_cancelled", dur: 10 * time.Second, cancel: true, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.cancel {
				var cancelFn context.CancelFunc
				ctx, cancelFn = context.WithCancel(ctx)
				cancelFn()
			}
			err := waitWithContext(ctx, tt.dur)
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
