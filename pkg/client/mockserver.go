package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
)

// MockMailServer is a mock implementation of the mail service for testing
type MockMailServer struct {
	server       *httptest.Server
	sentMails    []MailRequest
	mu           sync.RWMutex
	healthStatus int
	sendStatus   int
	errorMessage string
}

// NewMockMailServer creates and starts a new mock mail server
func NewMockMailServer() *MockMailServer {
	mock := &MockMailServer{
		sentMails:    make([]MailRequest, 0),
		healthStatus: http.StatusOK,
		sendStatus:   http.StatusOK,
	}

	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		mock.mu.RLock()
		status := mock.healthStatus
		mock.mu.RUnlock()

		w.WriteHeader(status)
		if status == http.StatusOK {
			_, _ = w.Write([]byte("OK"))
		}
	})

	// Send mail endpoint
	mux.HandleFunc("/v1/sendmail", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		mock.mu.RLock()
		sendStatus := mock.sendStatus
		errorMsg := mock.errorMessage
		mock.mu.RUnlock()

		var request MailRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"message": "Invalid request body",
			})
			return
		}

		// Store the mail
		mock.mu.Lock()
		mock.sentMails = append(mock.sentMails, request)
		mock.mu.Unlock()

		// Return configured status
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(sendStatus)

		if sendStatus != http.StatusOK {
			_ = json.NewEncoder(w).Encode(map[string]string{
				"message": errorMsg,
			})
			return
		}

		// Return the mail request as response
		response := MailResponse(request)
		_ = json.NewEncoder(w).Encode(response)
	})

	mock.server = httptest.NewServer(mux)
	return mock
}

// URL returns the base URL of the mock server
func (m *MockMailServer) URL() string {
	return m.server.URL
}

// Close shuts down the mock server
func (m *MockMailServer) Close() {
	m.server.Close()
}

// GetSentMails returns a copy of all mails sent to the mock server
func (m *MockMailServer) GetSentMails() []MailRequest {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent external modifications
	mails := make([]MailRequest, len(m.sentMails))
	copy(mails, m.sentMails)
	return mails
}

// GetLastSentMail returns the most recently sent mail, or nil if no mails were sent
func (m *MockMailServer) GetLastSentMail() *MailRequest {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.sentMails) == 0 {
		return nil
	}
	mail := m.sentMails[len(m.sentMails)-1]
	return &mail
}

// Reset clears all stored mails and resets status codes to defaults
func (m *MockMailServer) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sentMails = make([]MailRequest, 0)
	m.healthStatus = http.StatusOK
	m.sendStatus = http.StatusOK
	m.errorMessage = ""
}

// SetHealthStatus configures the HTTP status code returned by the health check endpoint
func (m *MockMailServer) SetHealthStatus(status int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.healthStatus = status
}

// SetSendMailStatus configures the HTTP status code and error message returned by the sendmail endpoint
func (m *MockMailServer) SetSendMailStatus(status int, errorMessage string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sendStatus = status
	m.errorMessage = errorMessage
}

// SentMailCount returns the number of mails sent to the mock server
func (m *MockMailServer) SentMailCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sentMails)
}
