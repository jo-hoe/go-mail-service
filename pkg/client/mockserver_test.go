package client

import (
	"context"
	"net/http"
	"testing"
)

func TestMockMailServer_SendMail(t *testing.T) {
	// Create and start the mock server
	mockServer := NewMockMailServer()
	defer mockServer.Close()

	// Create a client pointing to the mock server
	client := NewClient(mockServer.URL())

	// Send a test email
	request := MailRequest{
		To:          "test@example.com",
		Subject:     "Test Subject",
		HtmlContent: "<h1>Test Email</h1>",
		From:        "sender@example.com",
		FromName:    "Test Sender",
	}

	response, err := client.SendMail(context.Background(), request)
	if err != nil {
		t.Fatalf("Failed to send mail: %v", err)
	}

	// Verify response
	if response.To != request.To {
		t.Errorf("Expected To: %s, got %s", request.To, response.To)
	}
	if response.Subject != request.Subject {
		t.Errorf("Expected Subject: %s, got %s", request.Subject, response.Subject)
	}

	// Verify the mail was recorded by the mock server
	sentMails := mockServer.GetSentMails()
	if len(sentMails) != 1 {
		t.Fatalf("Expected 1 sent mail, got %d", len(sentMails))
	}

	sentMail := sentMails[0]
	if sentMail.To != request.To {
		t.Errorf("Expected stored To: %s, got %s", request.To, sentMail.To)
	}
	if sentMail.Subject != request.Subject {
		t.Errorf("Expected stored Subject: %s, got %s", request.Subject, sentMail.Subject)
	}
}

func TestMockMailServer_MultipleMails(t *testing.T) {
	mockServer := NewMockMailServer()
	defer mockServer.Close()

	client := NewClient(mockServer.URL())

	// Send multiple emails
	requests := []MailRequest{
		{To: "user1@example.com", Subject: "Subject 1", HtmlContent: "Content 1"},
		{To: "user2@example.com", Subject: "Subject 2", HtmlContent: "Content 2"},
		{To: "user3@example.com", Subject: "Subject 3", HtmlContent: "Content 3"},
	}

	for _, req := range requests {
		_, err := client.SendMail(context.Background(), req)
		if err != nil {
			t.Fatalf("Failed to send mail: %v", err)
		}
	}

	// Verify all mails were recorded
	sentMails := mockServer.GetSentMails()
	if len(sentMails) != 3 {
		t.Fatalf("Expected 3 sent mails, got %d", len(sentMails))
	}

	// Verify count method
	if mockServer.SentMailCount() != 3 {
		t.Errorf("Expected SentMailCount() to return 3, got %d", mockServer.SentMailCount())
	}

	// Verify each mail
	for i, req := range requests {
		if sentMails[i].To != req.To {
			t.Errorf("Mail %d: Expected To: %s, got %s", i, req.To, sentMails[i].To)
		}
		if sentMails[i].Subject != req.Subject {
			t.Errorf("Mail %d: Expected Subject: %s, got %s", i, req.Subject, sentMails[i].Subject)
		}
	}
}

func TestMockMailServer_GetLastSentMail(t *testing.T) {
	mockServer := NewMockMailServer()
	defer mockServer.Close()

	// Initially, no mails should be sent
	if mail := mockServer.GetLastSentMail(); mail != nil {
		t.Errorf("Expected nil for last sent mail, got %v", mail)
	}

	client := NewClient(mockServer.URL())

	// Send first email
	request1 := MailRequest{To: "user1@example.com", Subject: "First", HtmlContent: "Content 1"}
	_, _ = client.SendMail(context.Background(), request1)

	lastMail := mockServer.GetLastSentMail()
	if lastMail == nil {
		t.Fatal("Expected last sent mail to not be nil")
	}
	if lastMail.Subject != "First" {
		t.Errorf("Expected last mail subject 'First', got '%s'", lastMail.Subject)
	}

	// Send second email
	request2 := MailRequest{To: "user2@example.com", Subject: "Second", HtmlContent: "Content 2"}
	_, _ = client.SendMail(context.Background(), request2)

	lastMail = mockServer.GetLastSentMail()
	if lastMail == nil {
		t.Fatal("Expected last sent mail to not be nil")
	}
	if lastMail.Subject != "Second" {
		t.Errorf("Expected last mail subject 'Second', got '%s'", lastMail.Subject)
	}
}

func TestMockMailServer_Reset(t *testing.T) {
	mockServer := NewMockMailServer()
	defer mockServer.Close()

	client := NewClient(mockServer.URL())

	// Send some emails
	request := MailRequest{To: "test@example.com", Subject: "Test", HtmlContent: "Content"}
	_, _ = client.SendMail(context.Background(), request)
	_, _ = client.SendMail(context.Background(), request)

	// Verify mails were recorded
	if mockServer.SentMailCount() != 2 {
		t.Errorf("Expected 2 mails before reset, got %d", mockServer.SentMailCount())
	}

	// Reset the server
	mockServer.Reset()

	// Verify mails were cleared
	if mockServer.SentMailCount() != 0 {
		t.Errorf("Expected 0 mails after reset, got %d", mockServer.SentMailCount())
	}

	if mail := mockServer.GetLastSentMail(); mail != nil {
		t.Errorf("Expected nil for last sent mail after reset, got %v", mail)
	}
}

func TestMockMailServer_HealthCheck(t *testing.T) {
	mockServer := NewMockMailServer()
	defer mockServer.Close()

	client := NewClient(mockServer.URL())

	// Default health check should succeed
	err := client.HealthCheck(context.Background())
	if err != nil {
		t.Errorf("Expected health check to succeed, got error: %v", err)
	}

	// Configure health check to fail
	mockServer.SetHealthStatus(http.StatusServiceUnavailable)

	err = client.HealthCheck(context.Background())
	if err == nil {
		t.Error("Expected health check to fail, but it succeeded")
	}

	// Reset and verify health check succeeds again
	mockServer.Reset()
	err = client.HealthCheck(context.Background())
	if err != nil {
		t.Errorf("Expected health check to succeed after reset, got error: %v", err)
	}
}

func TestMockMailServer_ErrorResponse(t *testing.T) {
	mockServer := NewMockMailServer()
	defer mockServer.Close()

	client := NewClient(mockServer.URL())

	// Configure the server to return an error
	mockServer.SetSendMailStatus(http.StatusBadRequest, "Invalid email format")

	request := MailRequest{
		To:          "test@example.com",
		Subject:     "Test",
		HtmlContent: "Content",
	}

	_, err := client.SendMail(context.Background(), request)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Verify the error is an ErrorResponse
	errorResp, ok := err.(ErrorResponse)
	if !ok {
		t.Fatalf("Expected ErrorResponse, got %T", err)
	}

	if errorResp.Code != http.StatusBadRequest {
		t.Errorf("Expected error code %d, got %d", http.StatusBadRequest, errorResp.Code)
	}

	if errorResp.Message != "Invalid email format" {
		t.Errorf("Expected error message 'Invalid email format', got '%s'", errorResp.Message)
	}

	// Verify the mail was still recorded (even though it returned an error)
	if mockServer.SentMailCount() != 1 {
		t.Errorf("Expected mail to be recorded even with error, got count %d", mockServer.SentMailCount())
	}
}

func TestMockMailServer_ConcurrentRequests(t *testing.T) {
	mockServer := NewMockMailServer()
	defer mockServer.Close()

	client := NewClient(mockServer.URL())

	// Send multiple emails concurrently
	numGoroutines := 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			request := MailRequest{
				To:          "test@example.com",
				Subject:     "Test",
				HtmlContent: "Content",
			}
			_, err := client.SendMail(context.Background(), request)
			if err != nil {
				t.Errorf("Goroutine %d: Failed to send mail: %v", id, err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify all mails were recorded
	if mockServer.SentMailCount() != numGoroutines {
		t.Errorf("Expected %d mails, got %d", numGoroutines, mockServer.SentMailCount())
	}
}

func ExampleMockMailServer() {
	// Create a mock mail server
	mockServer := NewMockMailServer()
	defer mockServer.Close()

	// Create a client pointing to the mock server
	client := NewClient(mockServer.URL())

	// Send an email
	request := MailRequest{
		To:          "user@example.com",
		Subject:     "Welcome!",
		HtmlContent: "<h1>Welcome to our service</h1>",
		From:        "noreply@example.com",
		FromName:    "Example Service",
	}

	_, err := client.SendMail(context.Background(), request)
	if err != nil {
		panic(err)
	}

	// Verify the email was sent
	sentMails := mockServer.GetSentMails()
	if len(sentMails) == 1 {
		mail := sentMails[0]
		println("Sent mail to:", mail.To)
		println("Subject:", mail.Subject)
	}
}

func ExampleMockMailServer_errorScenario() {
	// Create a mock mail server
	mockServer := NewMockMailServer()
	defer mockServer.Close()

	// Configure the server to simulate an error
	mockServer.SetSendMailStatus(http.StatusServiceUnavailable, "Service temporarily unavailable")

	// Create a client
	client := NewClient(mockServer.URL())

	// Try to send an email
	request := MailRequest{
		To:          "user@example.com",
		Subject:     "Test",
		HtmlContent: "Content",
	}

	_, err := client.SendMail(context.Background(), request)
	if err != nil {
		println("Expected error occurred:", err.Error())
	}

	// Reset the server to normal operation
	mockServer.Reset()

	// Now sending should work
	_, err = client.SendMail(context.Background(), request)
	if err == nil {
		println("Mail sent successfully after reset")
	}
}
