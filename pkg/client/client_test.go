package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	baseURL := "http://localhost:8080"
	client := NewClient(baseURL)

	if client.baseURL != baseURL {
		t.Errorf("Expected baseURL %s, got %s", baseURL, client.baseURL)
	}

	if client.httpClient.Timeout != 30*time.Second {
		t.Errorf("Expected default timeout 30s, got %v", client.httpClient.Timeout)
	}
}

func TestNewClientWithOptions(t *testing.T) {
	baseURL := "http://localhost:8080"
	customTimeout := 10 * time.Second
	client := NewClient(baseURL, WithTimeout(customTimeout))

	if client.httpClient.Timeout != customTimeout {
		t.Errorf("Expected timeout %v, got %v", customTimeout, client.httpClient.Timeout)
	}
}

func TestSendMail_Success(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/v1/sendmail" {
			t.Errorf("Expected path /v1/sendmail, got %s", r.URL.Path)
		}

		// Verify content type
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Parse request body
		var request MailRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		// Verify request data
		if request.To != "test@example.com" {
			t.Errorf("Expected To: test@example.com, got %s", request.To)
		}
		if request.Subject != "Test Subject" {
			t.Errorf("Expected Subject: Test Subject, got %s", request.Subject)
		}

		// Send response
		response := MailResponse(request)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	client := NewClient(server.URL)

	// Test request
	request := MailRequest{
		To:          "test@example.com",
		Subject:     "Test Subject",
		HtmlContent: "Test Body",
		From:        "sender@example.com",
		FromName:    "Test Sender",
	}

	response, err := client.SendMail(context.Background(), request)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if response.To != request.To {
		t.Errorf("Expected To: %s, got %s", request.To, response.To)
	}
	if response.Subject != request.Subject {
		t.Errorf("Expected Subject: %s, got %s", request.Subject, response.Subject)
	}
}

func TestSendMail_ValidationErrors(t *testing.T) {
	client := NewClient("http://localhost:8080")

	tests := []struct {
		name        string
		request     MailRequest
		expectedErr string
	}{
		{
			name:        "missing to field",
			request:     MailRequest{Subject: "Test", HtmlContent: "Body"},
			expectedErr: "'to' field is required",
		},
		{
			name:        "missing subject field",
			request:     MailRequest{To: "test@example.com", HtmlContent: "Body"},
			expectedErr: "'subject' field is required",
		},
		{
			name:        "missing content field",
			request:     MailRequest{To: "test@example.com", Subject: "Test"},
			expectedErr: "'content' field is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := client.SendMail(context.Background(), tt.request)
			if err == nil {
				t.Errorf("Expected error, got nil")
			}
			if err.Error() != tt.expectedErr {
				t.Errorf("Expected error '%s', got '%s'", tt.expectedErr, err.Error())
			}
		})
	}
}

func TestSendMail_HTTPError(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		errorResp := map[string]string{"message": "Invalid request"}
		_ = json.NewEncoder(w).Encode(errorResp)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	request := MailRequest{
		To:          "test@example.com",
		Subject:     "Test Subject",
		HtmlContent: "Test Body",
	}

	_, err := client.SendMail(context.Background(), request)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	// Check if it's an ErrorResponse
	if errorResp, ok := err.(ErrorResponse); ok {
		if errorResp.Code != http.StatusBadRequest {
			t.Errorf("Expected error code %d, got %d", http.StatusBadRequest, errorResp.Code)
		}
	} else {
		t.Errorf("Expected ErrorResponse type, got %T", err)
	}
}

func TestHealthCheck_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/" {
			t.Errorf("Expected path /, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.HealthCheck(context.Background())
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestHealthCheck_Failure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.HealthCheck(context.Background())
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestWithTimeout(t *testing.T) {
	customTimeout := 5 * time.Second
	client := NewClient("http://localhost:8080", WithTimeout(customTimeout))

	if client.httpClient.Timeout != customTimeout {
		t.Errorf("Expected timeout %v, got %v", customTimeout, client.httpClient.Timeout)
	}
}

func TestWithHTTPClient(t *testing.T) {
	customClient := &http.Client{Timeout: 15 * time.Second}
	client := NewClient("http://localhost:8080", WithHTTPClient(customClient))

	if client.httpClient != customClient {
		t.Errorf("Expected custom HTTP client to be set")
	}
}
