# Go Mail Service Client

A simple Go client library for interacting with the Go Mail Service API.

## Installation

```bash
go get github.com/jo-hoe/go-mail-service/pkg/client@latest
```

**Note:** The client library is a separate Go module with zero external dependencies. It only uses Go's standard library, so you won't inherit any unnecessary dependencies from the mail service implementation (like SendGrid, Echo framework, or validators).

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/jo-hoe/go-mail-service/pkg/client"
)

func main() {
    // Create a new client
    mailClient := client.NewClient("http://localhost:80")

    // Create a mail request
    request := client.MailRequest{
        To:          "recipient@example.com",
        Subject:     "Hello from Go Mail Service",
        HtmlContent: "<h1>Hello!</h1><p>This is a test email.</p>",
        From:        "sender@example.com",
        FromName:    "Test Sender",
    }

    // Send the email
    response, err := mailClient.SendMail(context.Background(), request)
    if err != nil {
        log.Fatalf("Failed to send email: %v", err)
    }

    fmt.Printf("Email sent successfully to: %s\n", response.To)
}
```

## API Reference

### Creating a Client

#### `NewClient(baseURL string, options ...ClientOption) *Client`

Creates a new mail service client with the given base URL and optional configuration.

```go
// Basic client
client := client.NewClient("http://localhost:80")

// Client with custom timeout
client := client.NewClient("http://localhost:80", client.WithTimeout(10*time.Second))

// Client with custom HTTP client
httpClient := &http.Client{Timeout: 15 * time.Second}
client := client.NewClient("http://localhost:80", client.WithHTTPClient(httpClient))
```

### Client Options

#### `WithTimeout(timeout time.Duration) ClientOption`

Sets a custom timeout for HTTP requests (default: 30 seconds).

#### `WithHTTPClient(httpClient *http.Client) ClientOption`

Sets a custom HTTP client for making requests.

### Sending Mail

#### `SendMail(ctx context.Context, request MailRequest) (*MailResponse, error)`

Sends an email using the mail service.

**MailRequest fields:**
- `To` (required): Recipient email address(es), comma-separated for multiple recipients
- `Subject` (required): Email subject line
- `HtmlContent` (required): Email body content in HTML format
- `From` (optional): Sender email address
- `FromName` (optional): Display name for the sender

```go
request := client.MailRequest{
    To:          "user@example.com",
    Subject:     "Test Email",
    HtmlContent: "<p>Hello World!</p>",
    From:        "sender@company.com",
    FromName:    "Company Name",
}

response, err := mailClient.SendMail(context.Background(), request)
if err != nil {
    // Handle error
    if errorResp, ok := err.(client.ErrorResponse); ok {
        fmt.Printf("HTTP Error %d: %s\n", errorResp.Code, errorResp.Message)
    } else {
        fmt.Printf("Error: %v\n", err)
    }
    return
}

fmt.Printf("Email sent to: %s\n", response.To)
```

### Health Check

#### `HealthCheck(ctx context.Context) error`

Performs a health check against the mail service.

```go
err := mailClient.HealthCheck(context.Background())
if err != nil {
    log.Printf("Mail service is not healthy: %v", err)
} else {
    log.Println("Mail service is healthy!")
}
```

## Advanced Usage

### Multiple Recipients

Send emails to multiple recipients by providing a comma-separated list of email addresses:

```go
request := client.MailRequest{
    To:          "user1@example.com,user2@example.com,user3@example.com",
    Subject:     "Team Notification",
    HtmlContent: "<h2>Team Update</h2><p>This is for the whole team.</p>",
}
```

### Context with Timeout

Use context for request timeouts and cancellation:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

response, err := mailClient.SendMail(ctx, request)
```

### Error Handling

The client provides structured error handling:

```go
response, err := mailClient.SendMail(context.Background(), request)
if err != nil {
    switch e := err.(type) {
    case client.ErrorResponse:
        // HTTP error from the service
        fmt.Printf("Service error (HTTP %d): %s\n", e.Code, e.Message)
    default:
        // Other errors (network, validation, etc.)
        fmt.Printf("Client error: %v\n", err)
    }
    return
}
```

## Error Types

### `ErrorResponse`

Represents an HTTP error response from the mail service:

```go
type ErrorResponse struct {
    Message string `json:"message"`
    Code    int    `json:"code"`
}
```

Common HTTP status codes:
- `400 Bad Request`: Invalid request data (missing required fields, invalid email format, etc.)
- `500 Internal Server Error`: Server-side error

## Examples

The code examples above demonstrate the main usage patterns for the client library.

## Mock Server for Testing

The client library includes a mock mail server (`MockMailServer`) that can be used for testing your application without connecting to a real mail service. This is particularly useful for unit tests and integration tests.

### Creating a Mock Server

```go
import (
    "context"
    "testing"
    "github.com/jo-hoe/go-mail-service/pkg/client"
)

func TestMyEmailFeature(t *testing.T) {
    // Create and start the mock server
    mockServer := client.NewMockMailServer()
    defer mockServer.Close()

    // Create a client pointing to the mock server
    mailClient := client.NewClient(mockServer.URL())

    // Your test code here...
}
```

### Mock Server Features

The mock server provides several useful features for testing:

#### Recording Sent Emails

All emails sent to the mock server are recorded and can be retrieved for verification:

```go
mockServer := client.NewMockMailServer()
defer mockServer.Close()

mailClient := client.NewClient(mockServer.URL())

// Send an email
request := client.MailRequest{
    To:          "test@example.com",
    Subject:     "Test Email",
    HtmlContent: "<p>Test content</p>",
}
mailClient.SendMail(context.Background(), request)

// Verify the email was sent
sentMails := mockServer.GetSentMails()
if len(sentMails) != 1 {
    t.Errorf("Expected 1 email, got %d", len(sentMails))
}

// Check email details
if sentMails[0].To != "test@example.com" {
    t.Errorf("Unexpected recipient: %s", sentMails[0].To)
}
```

#### Getting the Last Sent Email

Convenient method to retrieve just the most recent email:

```go
// Send multiple emails
mailClient.SendMail(context.Background(), request1)
mailClient.SendMail(context.Background(), request2)

// Get only the last one
lastMail := mockServer.GetLastSentMail()
if lastMail != nil {
    fmt.Printf("Last email was to: %s\n", lastMail.To)
}
```

#### Counting Sent Emails

```go
count := mockServer.SentMailCount()
fmt.Printf("Total emails sent: %d\n", count)
```

#### Simulating Error Responses

Configure the mock server to return specific error responses:

```go
mockServer := client.NewMockMailServer()
defer mockServer.Close()

// Simulate a service error
mockServer.SetSendMailStatus(http.StatusServiceUnavailable, "Service temporarily unavailable")

mailClient := client.NewClient(mockServer.URL())

// This will now return an error
_, err := mailClient.SendMail(context.Background(), request)
if err != nil {
    errorResp, ok := err.(client.ErrorResponse)
    if ok {
        fmt.Printf("Got expected error: %s\n", errorResp.Message)
    }
}
```

#### Simulating Health Check Failures

```go
mockServer := client.NewMockMailServer()
defer mockServer.Close()

// Simulate unhealthy service
mockServer.SetHealthStatus(http.StatusServiceUnavailable)

mailClient := client.NewClient(mockServer.URL())

err := mailClient.HealthCheck(context.Background())
if err != nil {
    fmt.Println("Health check failed as expected")
}
```

#### Resetting the Mock Server

Clear all recorded emails and reset status codes:

```go
mockServer := client.NewMockMailServer()
defer mockServer.Close()

// Send some emails...
mailClient.SendMail(context.Background(), request)

// Reset everything
mockServer.Reset()

// Now the server has no recorded emails and default status codes
if mockServer.SentMailCount() != 0 {
    t.Error("Expected no emails after reset")
}
```

### Complete Testing Example

```go
func TestEmailNotification(t *testing.T) {
    // Setup mock server
    mockServer := client.NewMockMailServer()
    defer mockServer.Close()

    // Create your application's email sender with the mock client
    mailClient := client.NewClient(mockServer.URL())
    
    // Test your business logic that sends emails
    err := sendWelcomeEmail(mailClient, "newuser@example.com")
    if err != nil {
        t.Fatalf("Failed to send welcome email: %v", err)
    }

    // Verify the email was sent correctly
    sentMails := mockServer.GetSentMails()
    if len(sentMails) != 1 {
        t.Fatalf("Expected 1 email, got %d", len(sentMails))
    }

    email := sentMails[0]
    if email.To != "newuser@example.com" {
        t.Errorf("Wrong recipient: %s", email.To)
    }
    if email.Subject != "Welcome to Our Service" {
        t.Errorf("Wrong subject: %s", email.Subject)
    }
    if !strings.Contains(email.HtmlContent, "Welcome") {
        t.Error("Email content doesn't contain welcome message")
    }
}

func TestEmailErrorHandling(t *testing.T) {
    mockServer := client.NewMockMailServer()
    defer mockServer.Close()

    // Simulate service failure
    mockServer.SetSendMailStatus(http.StatusInternalServerError, "Database error")

    mailClient := client.NewClient(mockServer.URL())
    
    // Your code should handle this error gracefully
    err := sendWelcomeEmail(mailClient, "user@example.com")
    if err == nil {
        t.Error("Expected error when service is down")
    }

    // Verify error was properly propagated
    if errorResp, ok := err.(client.ErrorResponse); ok {
        if errorResp.Code != http.StatusInternalServerError {
            t.Errorf("Wrong error code: %d", errorResp.Code)
        }
    }
}
```

### Thread Safety

The mock server is thread-safe and can handle concurrent requests, making it suitable for testing concurrent email operations:

```go
func TestConcurrentEmails(t *testing.T) {
    mockServer := client.NewMockMailServer()
    defer mockServer.Close()

    mailClient := client.NewClient(mockServer.URL())

    // Send emails concurrently
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            request := client.MailRequest{
                To:          fmt.Sprintf("user%d@example.com", id),
                Subject:     "Test",
                HtmlContent: "Content",
            }
            mailClient.SendMail(context.Background(), request)
        }(i)
    }
    wg.Wait()

    // All 10 emails should be recorded
    if mockServer.SentMailCount() != 10 {
        t.Errorf("Expected 10 emails, got %d", mockServer.SentMailCount())
    }
}
```

## Testing

Run the client tests (including mock server tests):

```bash
go test ./pkg/client/...
```

Run with verbose output:

```bash
go test -v ./pkg/client/...
```

Run with coverage:

```bash
go test -cover ./pkg/client/...
```

## License

This client library is part of the Go Mail Service project and follows the same license terms.
