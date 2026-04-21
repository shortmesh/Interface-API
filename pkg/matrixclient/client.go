package matrixclient

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"interface-api/pkg/config"
)

func New() (*Client, error) {
	baseURL := os.Getenv("MATRIX_CLIENT_URL")
	if baseURL == "" {
		return nil, fmt.Errorf("MATRIX_CLIENT_URL environment variable is not set")
	}

	if err := config.ValidateExternalURL(baseURL, "MATRIX_CLIENT_URL"); err != nil {
		return nil, err
	}

	clientID := os.Getenv("ADMIN_CLIENT_ID")
	if clientID == "" {
		return nil, fmt.Errorf("ADMIN_CLIENT_ID environment variable is not set")
	}

	clientSecret := os.Getenv("ADMIN_CLIENT_SECRET")
	if clientSecret == "" {
		return nil, fmt.Errorf("ADMIN_CLIENT_SECRET environment variable is not set")
	}

	return &Client{
		baseURL:      baseURL,
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}, nil
}

func generateNonce() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

func (c *Client) calculateSignature(method, path, timestamp, nonce, body string) string {
	stringToSign := c.clientID + method + path + timestamp + nonce + body
	h := hmac.New(sha256.New, []byte(c.clientSecret))
	h.Write([]byte(stringToSign))
	return hex.EncodeToString(h.Sum(nil))
}

func (c *Client) addAuthHeaders(req *http.Request, body string) error {
	timestamp := fmt.Sprintf("%d", time.Now().UTC().Unix())
	nonce, err := generateNonce()
	if err != nil {
		return err
	}
	signature := c.calculateSignature(req.Method, req.URL.Path, timestamp, nonce, body)

	req.Header.Set("X-ShortMesh-ID", c.clientID)
	req.Header.Set("X-ShortMesh-Timestamp", timestamp)
	req.Header.Set("X-ShortMesh-Nonce", nonce)
	req.Header.Set("X-ShortMesh-Signature", signature)

	return nil
}

func (c *Client) StoreCredentials(req *StoreCredentialsRequest) (*StoreCredentialsResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	path := "/api/v1/store"
	httpReq, err := http.NewRequest("POST", fmt.Sprintf("%s%s", c.baseURL, path), bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if err := c.addAuthHeaders(httpReq, string(body)); err != nil {
		return nil, fmt.Errorf("failed to add auth headers: %w", err)
	}

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

	var storeResp StoreCredentialsResponse
	if err := json.Unmarshal(respBody, &storeResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &storeResp, nil
}

func (c *Client) AddDevice(req *AddDeviceRequest) (*AddDeviceResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	path := "/api/v1/devices"
	httpReq, err := http.NewRequest("POST", fmt.Sprintf("%s%s", c.baseURL, path), bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if err := c.addAuthHeaders(httpReq, string(body)); err != nil {
		return nil, fmt.Errorf("failed to add auth headers: %w", err)
	}

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

	var deviceResp AddDeviceResponse
	if err := json.Unmarshal(respBody, &deviceResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &deviceResp, nil
}

func (c *Client) DeleteDevice(req *DeleteDeviceRequest) (*DeleteDeviceResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	path := "/api/v1/devices"
	httpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s%s", c.baseURL, path), bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if err := c.addAuthHeaders(httpReq, string(body)); err != nil {
		return nil, fmt.Errorf("failed to add auth headers: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(respBody))
	}

	var deleteResp DeleteDeviceResponse
	if err := json.Unmarshal(respBody, &deleteResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &deleteResp, nil
}

func (c *Client) ListDevices(req *ListDevicesRequest) (ListDevicesResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	path := "/api/v1/devices"
	httpReq, err := http.NewRequest("GET", fmt.Sprintf("%s%s", c.baseURL, path), bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if err := c.addAuthHeaders(httpReq, string(body)); err != nil {
		return nil, fmt.Errorf("failed to add auth headers: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(respBody))
	}

	var listResp ListDevicesResponse
	if err := json.Unmarshal(respBody, &listResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if listResp == nil {
		listResp = ListDevicesResponse{}
	}

	return listResp, nil
}

func (c *Client) SendMessage(deviceID string, req *SendMessageRequest) (*SendMessageResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	path := fmt.Sprintf("/api/v1/devices/%s/message", deviceID)
	httpReq, err := http.NewRequest("POST", fmt.Sprintf("%s%s", c.baseURL, path), bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if err := c.addAuthHeaders(httpReq, string(body)); err != nil {
		return nil, fmt.Errorf("failed to add auth headers: %w", err)
	}

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

	var sendResp SendMessageResponse
	if err := json.Unmarshal(respBody, &sendResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &sendResp, nil
}

func (c *Client) DeleteToken(req *DeleteTokenRequest) (*DeleteTokenResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	path := "/api/v1/users"
	httpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s%s", c.baseURL, path), bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if err := c.addAuthHeaders(httpReq, string(body)); err != nil {
		return nil, fmt.Errorf("failed to add auth headers: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(respBody))
	}

	var deleteResp DeleteTokenResponse
	if err := json.Unmarshal(respBody, &deleteResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &deleteResp, nil
}
