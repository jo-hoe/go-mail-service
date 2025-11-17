package mailjet

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/jo-hoe/go-mail-service/internal/mail"
)

// MailjetService implements MailService
type MailjetService struct {
	config *MailjetConfig
	client *http.Client
}

// NewMailjetService creates a MailjetService
func NewMailjetService(config *MailjetConfig) *MailjetService {
	return &MailjetService{
		config: config,
		client: &http.Client{},
	}
}

// mailjetMessage represents a single message in Mailjet's API format
type mailjetMessage struct {
	From        mailjetEmail   `json:"From"`
	To          []mailjetEmail `json:"To"`
	Subject     string         `json:"Subject"`
	TextPart    string         `json:"TextPart,omitempty"`
	HTMLPart    string         `json:"HTMLPart,omitempty"`
}

// mailjetEmail represents an email address with optional name
type mailjetEmail struct {
	Email string `json:"Email"`
	Name  string `json:"Name,omitempty"`
}

// mailjetRequest represents the Mailjet API request payload
type mailjetRequest struct {
	Messages []mailjetMessage `json:"Messages"`
}

// mailjetResponse represents the Mailjet API response
type mailjetResponse struct {
	Messages []mailjetMessageResponse `json:"Messages"`
}

// mailjetMessageResponse represents a single message response
type mailjetMessageResponse struct {
	Status string                  `json:"Status"`
	To     []mailjetRecipientResponse `json:"To,omitempty"`
	Errors []mailjetError          `json:"Errors,omitempty"`
}

// mailjetRecipientResponse represents recipient info in response
type mailjetRecipientResponse struct {
	Email       string `json:"Email"`
	MessageUUID string `json:"MessageUUID"`
	MessageID   int64  `json:"MessageID"`
	MessageHref string `json:"MessageHref"`
}

// mailjetError represents an error in the response
type mailjetError struct {
	ErrorIdentifier string   `json:"ErrorIdentifier"`
	ErrorCode       string   `json:"ErrorCode"`
	StatusCode      int      `json:"StatusCode"`
	ErrorMessage    string   `json:"ErrorMessage"`
	ErrorRelatedTo  []string `json:"ErrorRelatedTo,omitempty"`
}

func (service *MailjetService) SendMail(ctx context.Context, attributes mail.MailAttributes) error {
	message := service.createMessage(attributes)
	return service.sendRequest(ctx, message)
}

// createMessage creates a Mailjet message from mail attributes
func (service *MailjetService) createMessage(attributes mail.MailAttributes) mailjetMessage {
	from := mailjetEmail{
		Email: service.config.OriginAddress,
		Name:  service.config.OriginName,
	}

	// Parse recipients (comma-separated)
	toEmails := []mailjetEmail{}
	mailAddresses := strings.Split(attributes.To, ",")
	for _, mailAddress := range mailAddresses {
		mailAddress = strings.TrimSpace(mailAddress)
		if mailAddress != "" {
			toEmails = append(toEmails, mailjetEmail{
				Email: mailAddress,
			})
		}
	}

	return mailjetMessage{
		From:     from,
		To:       toEmails,
		Subject:  attributes.Subject,
		HTMLPart: attributes.HtmlContent,
	}
}

func (service *MailjetService) sendRequest(ctx context.Context, message mailjetMessage) error {
	// Create request payload
	payload := mailjetRequest{
		Messages: []mailjetMessage{message},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.mailjet.com/v3.1/send", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	
	// Set Basic Authentication
	auth := base64.StdEncoding.EncodeToString([]byte(service.config.APIKeyPublic + ":" + service.config.APIKeyPrivate))
	req.Header.Set("Authorization", "Basic "+auth)

	// Send request
	resp, err := service.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("mailjet API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var mailjetResp mailjetResponse
	if err := json.Unmarshal(body, &mailjetResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for errors in response
	if len(mailjetResp.Messages) > 0 {
		msg := mailjetResp.Messages[0]
		if msg.Status == "error" && len(msg.Errors) > 0 {
			firstError := msg.Errors[0]
			return fmt.Errorf("mailjet error [%s]: %s", firstError.ErrorCode, firstError.ErrorMessage)
		}
	}

	return nil
}
