package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"bytes"
	"sync"
	"time"
)

// Credentials holds the client ID and access key for CloudInfra auth.
type Credentials struct {
	ClientID  string
	AccessKey string
}

// Client handles authentication with the CloudInfra API.
type Client struct {
	baseURL    string
	httpClient *http.Client

	mu     sync.RWMutex
	token  string
	expiry time.Time
}

// tokenResponse represents the CloudInfra auth response.
// Example: {"success":true,"data":{"token":"eyJ...","csrf":"...","expires":"Thu, 19 Mar 2026 13:36:19 GMT","expiresIn":1800}}
type tokenResponse struct {
	Data struct {
		Token     string `json:"token"`
		CSRF      string `json:"csrf"`
		Expires   string `json:"expires"`
		ExpiresIn int    `json:"expiresIn"` // seconds
	} `json:"data"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// NewClient creates a new CloudInfra auth client.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetToken exchanges client credentials for a JWT token.
// It caches the token and refreshes it when near expiry.
func (c *Client) GetToken(ctx context.Context, creds Credentials) (string, error) {
	c.mu.RLock()
	if c.token != "" && time.Now().Before(c.expiry.Add(-1*time.Minute)) {
		token := c.token
		c.mu.RUnlock()
		return token, nil
	}
	c.mu.RUnlock()

	return c.refreshToken(ctx, creds)
}

// refreshToken fetches a new JWT token from CloudInfra.
func (c *Client) refreshToken(ctx context.Context, creds Credentials) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	if c.token != "" && time.Now().Before(c.expiry.Add(-1*time.Minute)) {
		return c.token, nil
	}

	url := c.baseURL + "/auth/external"
	payload := map[string]string{"clientId": creds.ClientID, "accessKey": creds.AccessKey}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshaling auth request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("creating auth request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("executing auth request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading auth response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("auth request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var tokenResp tokenResponse
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return "", fmt.Errorf("parsing auth response: %w", err)
	}

	if !tokenResp.Success {
		return "", fmt.Errorf("authentication failed: %s", tokenResp.Message)
	}

	c.token = tokenResp.Data.Token

	// Use expiresIn (seconds) for expiry calculation
	if tokenResp.Data.ExpiresIn > 0 {
		c.expiry = time.Now().Add(time.Duration(tokenResp.Data.ExpiresIn) * time.Second)
	} else {
		c.expiry = time.Now().Add(25 * time.Minute)
	}

	return c.token, nil
}
