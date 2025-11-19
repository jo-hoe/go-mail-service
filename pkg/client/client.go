package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client represents a mail service client
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// MailRequest represents the structure for sending mail
type MailRequest struct {
	To          string `json:"to"`
	Subject     string `json:"subject"`
	HtmlContent string `json:"content"`
	From        string `json:"from,omitempty"`
	FromName    string `json:"fromName,omitempty"`
}

// MailResponse represents the response from the mail service
type MailResponse struct {
	To          string `json:"to"`
	Subject     string `json:"subject"`
	HtmlContent string `json:"content"`
	From        string `json:"from,omitempty"`
	FromName    string `json:"fromName,omitempty"`
}

// ErrorResponse represents an error response from the service
type ErrorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func (e ErrorResponse) Error() string {
	return fmt.Sprintf("mail service error (code %d): %s", e.Code, e.Message)
}

// NewClient creates a new mail service client
func NewClient(baseURL string, options ...ClientOption) *Client {
	client := &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	for _, option := range options {
		option(client)
	}

	return client
}

// ClientOption represents configuration options for the client
type ClientOption func(*Client)

// WithTimeout sets a custom timeout for HTTP requests
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// SendMail sends an email using the mail service
func (c *Client) SendMail(ctx context.Context, request MailRequest) (*MailResponse, error) {
	// Validate required fields
	if request.To == "" {
		return nil, fmt.Errorf("'to' field is required")
	}
	if request.Subject == "" {
		return nil, fmt.Errorf("'subject' field is required")
	}
	if request.HtmlContent == "" {
		return nil, fmt.Errorf("'content' field is required")
	}
	// Note: From and FromName are optional - the service will use defaults if not provided

	// Prepare request body
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/v1/sendmail", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Handle error responses
	if resp.StatusCode != http.StatusOK {
		var errorResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			return nil, fmt.Errorf("received HTTP %d but failed to decode error response: %w", resp.StatusCode, err)
		}
		errorResp.Code = resp.StatusCode
		return nil, errorResp
	}

	// Parse successful response
	var mailResp MailResponse
	if err := json.NewDecoder(resp.Body).Decode(&mailResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &mailResp, nil
}

// HealthCheck performs a health check against the service
func (c *Client) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}

	return nil
}
