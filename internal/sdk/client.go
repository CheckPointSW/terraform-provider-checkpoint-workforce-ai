package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/CheckPointSW/terraform-provider-checkpoint-workforce-ai/internal/sdk/auth"
)

const (
	maxRetries       = 3
	retryBaseDelay   = 2 * time.Second
	rateLimitBackoff = 30 * time.Second // Start with 30s, not the full 5min block
)

// BaseClient provides HTTP operations with auth, retry, and rate-limit handling.
type BaseClient struct {
	BaseURL    string
	Token      string
	AuthClient *auth.Client
	Creds      auth.Credentials
	HTTPClient *http.Client
}

// NewBaseClient creates a new base client.
func NewBaseClient(baseURL, token string, authClient *auth.Client, creds auth.Credentials) *BaseClient {
	return &BaseClient{
		BaseURL:    baseURL,
		Token:      token,
		AuthClient: authClient,
		Creds:      creds,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// refreshToken gets a fresh JWT token.
func (c *BaseClient) refreshToken(ctx context.Context) error {
	token, err := c.AuthClient.GetToken(ctx, c.Creds)
	if err != nil {
		return err
	}
	c.Token = token
	return nil
}

// DoRequest executes an HTTP request with retry and rate-limit handling.
func (c *BaseClient) DoRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		reqBody = strings.NewReader(string(jsonBytes))
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Rebuild reader on retry
		if attempt > 0 && body != nil {
			jsonBytes, _ := json.Marshal(body)
			reqBody = strings.NewReader(string(jsonBytes))
		}

		req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, reqBody)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.Token)
		req.Header.Set("x-api-source", "terraform")

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			if attempt < maxRetries {
				time.Sleep(retryBaseDelay * time.Duration(1<<attempt))
				continue
			}
			return nil, fmt.Errorf("executing request: %w", err)
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("reading response: %w", err)
		}

		switch {
		case resp.StatusCode == http.StatusUnauthorized:
			// Token expired, refresh and retry
			if err := c.refreshToken(ctx); err != nil {
				return nil, fmt.Errorf("refreshing token: %w", err)
			}
			continue

		case resp.StatusCode == http.StatusTooManyRequests:
			// Rate limited
			if attempt < maxRetries {
				time.Sleep(rateLimitBackoff * time.Duration(1<<attempt))
				continue
			}
			return nil, fmt.Errorf("rate limited after %d retries: %s", maxRetries, string(respBody))

		case resp.StatusCode >= 200 && resp.StatusCode < 300:
			return respBody, nil

		case resp.StatusCode >= 500 && attempt < maxRetries:
			// Server error, retry
			time.Sleep(retryBaseDelay * time.Duration(1<<attempt))
			continue

		default:
			return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
		}
	}

	return nil, fmt.Errorf("request failed after %d retries", maxRetries+1)
}

// Get performs a GET request.
func (c *BaseClient) Get(ctx context.Context, path string) ([]byte, error) {
	return c.DoRequest(ctx, http.MethodGet, path, nil)
}

// Post performs a POST request.
func (c *BaseClient) Post(ctx context.Context, path string, body interface{}) ([]byte, error) {
	return c.DoRequest(ctx, http.MethodPost, path, body)
}

// Put performs a PUT request.
func (c *BaseClient) Put(ctx context.Context, path string, body interface{}) ([]byte, error) {
	return c.DoRequest(ctx, http.MethodPut, path, body)
}

// Patch performs a PATCH request.
func (c *BaseClient) Patch(ctx context.Context, path string, body interface{}) ([]byte, error) {
	return c.DoRequest(ctx, http.MethodPatch, path, body)
}

// Delete performs a DELETE request.
func (c *BaseClient) Delete(ctx context.Context, path string) ([]byte, error) {
	return c.DoRequest(ctx, http.MethodDelete, path, nil)
}
