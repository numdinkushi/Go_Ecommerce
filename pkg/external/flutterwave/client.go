package flutterwave

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

const (
	baseURL = "https://api.flutterwave.com"
)

type Client struct {
	httpClient *http.Client
	secretKey  string
}

type VerifyAccountRequest struct {
	AccountNumber string `json:"account_number"`
	AccountBank   string `json:"account_bank"`
}

type Bank struct {
	ID   int    `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type GetBanksResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    []Bank `json:"data"`
}

type ErrorResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type VerifyAccountResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		AccountNumber string `json:"account_number"`
		AccountName   string `json:"account_name"`
	} `json:"data"`
}

func NewClient(secretKey string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		secretKey: secretKey,
	}
}

func (c *Client) GetBanks(country string) ([]Bank, error) {
	url := fmt.Sprintf("%s/banks?country=%s", baseURL, country)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.secretKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check if response is HTML (indicates error page or wrong endpoint)
	if len(body) > 0 && body[0] == '<' {
		return nil, fmt.Errorf("API returned HTML instead of JSON (status %d). Check endpoint URL. Response preview: %s", resp.StatusCode, string(body[:min(200, len(body))]))
	}

	// Check status code first
	if resp.StatusCode != http.StatusOK {
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Message != "" {
			return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, errorResp.Message)
		}
		// If not JSON, return the raw response
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Try to parse as JSON
	var apiResponse GetBanksResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response (status %d): %w. Response preview: %s", resp.StatusCode, err, string(body[:min(200, len(body))]))
	}

	if apiResponse.Status != "success" {
		return nil, fmt.Errorf("API error: %s", apiResponse.Message)
	}

	return apiResponse.Data, nil
}

func (c *Client) VerifyAccount(accountNumber, bankCode string) (*VerifyAccountResponse, error) {
	url := fmt.Sprintf("%s/v3/accounts/resolve", baseURL)

	requestBody := VerifyAccountRequest{
		AccountNumber: accountNumber,
		AccountBank:   bankCode,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.secretKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err == nil {
			return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, errorResp.Message)
		}
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var apiResponse VerifyAccountResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if apiResponse.Status != "success" {
		return nil, fmt.Errorf("verification failed: %s", apiResponse.Message)
	}

	return &apiResponse, nil
}
