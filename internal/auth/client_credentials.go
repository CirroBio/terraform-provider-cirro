package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
	Error       string `json:"error"`
	ErrorDesc   string `json:"error_description"`
}

// ClientCredentials manages OAuth2 client_credentials token refresh.
type ClientCredentials struct {
	clientID     string
	clientSecret string
	tokenURL     string

	mu          sync.Mutex
	accessToken string
	expiry      time.Time
}

// NewClientCredentials creates a token manager for the given tenant base URL.
// tokenURL should be the full URL to the token endpoint, e.g. https://app.cirro.bio/auth/token.
func NewClientCredentials(clientID, clientSecret, tokenURL string) *ClientCredentials {
	return &ClientCredentials{
		clientID:     clientID,
		clientSecret: clientSecret,
		tokenURL:     tokenURL,
	}
}

// GetToken returns a valid access token, refreshing if necessary.
func (c *ClientCredentials) GetToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.accessToken != "" && time.Now().Before(c.expiry) {
		return c.accessToken, nil
	}

	return c.refresh(ctx)
}

func (c *ClientCredentials) refresh(ctx context.Context) (string, error) {
	basic := base64.StdEncoding.EncodeToString(
		[]byte(c.clientID + ":" + c.clientSecret),
	)

	body := url.Values{"grant_type": {"client_credentials"}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.tokenURL,
		strings.NewReader(body.Encode()))
	if err != nil {
		return "", fmt.Errorf("build token request: %w", err)
	}
	req.Header.Set("Authorization", "Basic "+basic)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read token response: %w", err)
	}

	var tr tokenResponse
	if err := json.Unmarshal(raw, &tr); err != nil {
		return "", fmt.Errorf("parse token response: %w", err)
	}

	if tr.AccessToken == "" {
		return "", fmt.Errorf("authentication failed: %s – %s", tr.Error, tr.ErrorDesc)
	}

	expiresIn := tr.ExpiresIn
	if expiresIn == 0 {
		expiresIn = 3600
	}
	// Expire 30 seconds early to avoid using a token that's about to expire.
	c.accessToken = tr.AccessToken
	c.expiry = time.Now().Add(time.Duration(expiresIn)*time.Second - 30*time.Second)

	return c.accessToken, nil
}
