package masclient

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Client struct {
	baseURL       string
	adminBaseURL  string
	clientID      string
	clientSecret  string
	tokenLifetime int
	httpClient    *http.Client
}

func New() (*Client, error) {
	baseURL := os.Getenv("MAS_URL")
	if baseURL == "" {
		return nil, fmt.Errorf("MAS_URL environment variable is not set")
	}

	adminBaseURL := os.Getenv("MAS_ADMIN_URL")
	if adminBaseURL == "" {
		return nil, fmt.Errorf("MAS_ADMIN_URL environment variable is not set")
	}

	clientID := os.Getenv("ADMIN_CLIENT_ID")
	if clientID == "" {
		return nil, fmt.Errorf("ADMIN_CLIENT_ID environment variable is not set")
	}

	clientSecret := os.Getenv("ADMIN_CLIENT_SECRET")
	if clientSecret == "" {
		return nil, fmt.Errorf("ADMIN_CLIENT_SECRET environment variable is not set")
	}

	tokenLifetime := 15552000 // 6 months in seconds (180 days)

	return &Client{
		baseURL:       baseURL,
		adminBaseURL:  adminBaseURL,
		clientID:      clientID,
		clientSecret:  clientSecret,
		tokenLifetime: tokenLifetime,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (c *Client) GetAdminToken() (string, error) {
	data := "grant_type=client_credentials&scope=urn:mas:admin"

	httpReq, err := http.NewRequest("POST", fmt.Sprintf("%s/oauth2/token", c.baseURL), bytes.NewBufferString(data))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	auth := base64.StdEncoding.EncodeToString(fmt.Appendf(nil, "%s:%s", c.clientID, c.clientSecret))
	httpReq.Header.Set("Authorization", fmt.Sprintf("Basic %s", auth))
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(respBody))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return tokenResp.AccessToken, nil
}

func (c *Client) CreateUser(adminToken, username string) (*CreateUserResponse, error) {
	reqBody := map[string]string{"username": username}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", fmt.Sprintf("%s/api/admin/v1/users", c.adminBaseURL), bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", adminToken))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(respBody))
	}

	var createUserResp CreateUserResponse
	if err := json.Unmarshal(respBody, &createUserResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &createUserResp, nil
}

func (c *Client) CreatePersonalSession(adminToken, userID, deviceID string) (*CreatePersonalSessionResponse, error) {
	scope := fmt.Sprintf("openid urn:matrix:org.matrix.msc2967.client:api:* urn:matrix:org.matrix.msc2967.client:device:%s", deviceID)

	reqBody := map[string]any{
		"actor_user_id": userID,
		"human_name":    fmt.Sprintf("App session (%s)", deviceID),
		"scope":         scope,
		"expires_in":    c.tokenLifetime,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", fmt.Sprintf("%s/api/admin/v1/personal-sessions", c.adminBaseURL), bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", adminToken))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(respBody))
	}

	var sessionResp CreatePersonalSessionResponse
	if err := json.Unmarshal(respBody, &sessionResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &sessionResp, nil
}
